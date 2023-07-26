package websocket_adaptor

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"websocket-poc/internal/pkg"
	"websocket-poc/pkg/streamspb"

	"github.com/google/uuid"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/pkg/errors"
)

const keepaliveTimeout = 60 * time.Second

type sessionState struct {
	sessionID uuid.UUID
}

func OnWebsocket(ctx context.Context) func(w http.ResponseWriter, r *http.Request) {
	client, err := newStreamerClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		u := upgrade()
		conn, err := u.Upgrade(w, r, nil)
		if err != nil {
			panic(err)
		}

		log.Println("OnOpen:", conn.RemoteAddr().String())

		state := &sessionState{sessionID: uuid.New()}
		conn.SetSession(state)
		dataC := make(chan *streamspb.Response)
		controlC := make(chan *streamspb.Request)
		sessions.add(state.sessionID, dataC, controlC)

		go session(dataC, controlC, client, conn)
	}
}

func upgrade() *websocket.Upgrader {
	u := websocket.NewUpgrader()

	u.OnMessage(func(c *websocket.Conn, messageType websocket.MessageType, data []byte) {
		var rpcRequest subscriptionRPCRequest
		if err := json.Unmarshal(data, &rpcRequest); err != nil {
			log.Fatal(err)
		}
		if err := c.WriteMessage(messageType, []byte(`{"jsonrpc": "2.0", "id": 1, "result": "123456"}`)); err != nil {
			log.Fatal(err)
		}

		state := c.Session().(*sessionState)
		s, err := sessions.get(state.sessionID)
		if err != nil {
			log.Fatal(err)
		}

		args := rpcRequest.getArgs()
		s.controlC <- &streamspb.Request{
			Subscription: &streamspb.SubscriptionID{
				Symbol:   args["pair"].(string),
				Exchange: args["exchange"].(string),
				Feature:  mapEventTypeToFeature(rpcRequest.getType()),
			},
			Action: streamspb.Action_ADD,
		}

		if err := c.SetReadDeadline(time.Now().Add(keepaliveTimeout)); err != nil {
			log.Fatal(err)
		}
	})

	u.OnClose(func(c *websocket.Conn, err error) {
		state := c.Session().(*sessionState)
		sessions.remove(state.sessionID)
		log.Println("OnClose:", c.RemoteAddr().String(), err)
	})

	u.SetPingHandler(func(c *websocket.Conn, s string) {
		log.Println("ping")
		if err := c.SetReadDeadline(time.Now().Add(keepaliveTimeout)); err != nil {
			log.Fatal(err)
		}
	})
	return u
}

func session(dataC chan *streamspb.Response, controlC chan *streamspb.Request, client *streamerClient, conn *websocket.Conn) {
	state := pkg.NewSubscriptions()

	for {
		select {
		case data := <-dataC:
			if data == nil {
				return
			}
			if err := onDataMessage(state, data, conn); err != nil {
				log.Fatal(errors.Wrap(err, "on data message"))
			}

		case message := <-controlC:
			if message == nil {
				return
			}
			if err := onControlMessage(state, message, client); err != nil {
				log.Fatal(errors.Wrap(err, "on control message"))
			}
		}
	}
}

func onControlMessage(state *pkg.Subscriptions, message *streamspb.Request, client *streamerClient) error {
	switch message.Action {
	case streamspb.Action_ADD:
		log.Printf("adding %s\n", message.String())
		state.Add(message.Subscription)
		if err := client.addSubscription(message.Subscription); err != nil {
			return errors.Wrap(err, "add subscription")
		}
	case streamspb.Action_REMOVE:
		log.Printf("removing %s\n", message.String())

		if message.Subscription == nil {
			state.ForEach(func(sub *streamspb.SubscriptionID) {
				state.Remove(sub)
				if err := client.removeSubscription(sub); err != nil {
					log.Fatal(errors.Wrap(err, "remove subscription from remote"))
				}
			})
		} else {
			state.Remove(message.Subscription)
			if err := client.removeSubscription(message.Subscription); err != nil {
				return errors.Wrap(err, "remove subscription")
			}
		}
	}
	return nil
}

func onDataMessage(state *pkg.Subscriptions, data *streamspb.Response, conn *websocket.Conn) error {
	if !state.HasSubscription(data.Subscription) {
		return nil
	}

	bs, err := marshalRpcResponse(data)
	if err != nil {
		return errors.Wrap(err, "marshal rpc response")
	}
	return conn.WriteMessage(websocket.TextMessage, bs)
}
