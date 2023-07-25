package websocket_adaptor

import (
	"context"
	"io"
	"log"
	"net/url"
	"time"

	"google.golang.org/grpc/connectivity"

	"google.golang.org/grpc/credentials/insecure"

	"websocket-poc/internal/pkg"
	"websocket-poc/pkg/streamspb"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	spotTradesAddr  = "localhost:9090"
	spotTickersAddr = "localhost:9091"

	spotTradesAddrKey  = "SPOT_TRADES_SERVICE_ADDR"
	spotTickersAddrKey = "SPOT_TICKERS_SERVICE_ADDR"
)

var configMap = map[streamspb.Feature]*url.URL{
	streamspb.Feature_SPOT_TICKER: mustParseUrl(pkg.GetStringFromEnv(spotTickersAddrKey, spotTickersAddr)),
	streamspb.Feature_SPOT_TRADE:  mustParseUrl(pkg.GetStringFromEnv(spotTradesAddrKey, spotTradesAddr)),
}

var dialOpts = []grpc.DialOption{
	grpc.WithTransportCredentials(insecure.NewCredentials()),
}

type clientState struct {
	client        streamspb.EventStreamerClient
	sendC         chan *streamspb.Request
	reconnectC    chan struct{}
	conn          *grpc.ClientConn
	subscriptions *pkg.Subscriptions
}

type streamerClient struct {
	clients map[streamspb.Feature]*clientState
}

func newStreamerClient(ctx context.Context) (*streamerClient, error) {
	clients := make(map[streamspb.Feature]*clientState, len(configMap))
	sc := &streamerClient{
		clients: clients,
	}

	for feature, remoteURL := range configMap {
		conn, err := grpc.Dial(remoteURL.String(), dialOpts...)
		if err != nil {
			return nil, errors.Wrapf(err, "dial %s", remoteURL)
		}
		clients[feature] = &clientState{
			client:        streamspb.NewEventStreamerClient(conn),
			sendC:         make(chan *streamspb.Request),
			reconnectC:    make(chan struct{}),
			conn:          conn,
			subscriptions: pkg.NewSubscriptions(),
		}

		feature := feature
		go func() {
			if err := sc.startStream(ctx, clients[feature]); err != nil {
				log.Fatal(errors.Wrapf(err, "listen on %s", feature.String()))
			}
		}()
	}

	return sc, nil
}

func (c *streamerClient) startStream(ctx context.Context, state *clientState) error {
	go func() {
		if err := c.streamHandler(ctx, state); err != nil {
			log.Fatal(errors.Wrap(err, "handle stream"))
		}
	}()

	for {
		select {
		case <-state.reconnectC:
			if !isReconnected(state.conn, 1*time.Second, 60*time.Second) {
				return errors.New("failed to establish a connection within the defined timeout")
			}
			go func() {
				if err := c.streamHandler(ctx, state); err != nil {
					log.Fatal(errors.Wrap(err, "handle stream"))
				}
			}()
			state.subscriptions.ForEach(func(sub *streamspb.SubscriptionID) {
				state.sendC <- &streamspb.Request{
					Subscription: sub,
					Action:       streamspb.Action_ADD,
				}
			})
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *streamerClient) streamHandler(ctx context.Context, state *clientState) error {
	stream, err := state.client.Subscribe(ctx)
	if err != nil {
		return errors.Wrap(err, "create stream")
	}

	stopC := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(stopC)
				return
			}
			if err != nil {
				state.reconnectC <- struct{}{}
				close(stopC)
				return
			}
			go c.onData(in)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			close(stopC)
			return nil
		case <-stopC:
			return nil
		case msg := <-state.sendC:
			if err := stream.Send(msg); err != nil {
				return errors.Wrap(err, "stream send")
			}
		}
	}
}

func (c *streamerClient) onData(res *streamspb.Response) {
	sessions.broadcast(res)
}

func (c *streamerClient) addSubscription(sub *streamspb.SubscriptionID) error {
	c.clients[sub.Feature].sendC <- &streamspb.Request{
		Subscription: sub,
		Action:       streamspb.Action_ADD,
	}
	c.clients[sub.Feature].subscriptions.Add(sub)
	return nil
}

func (c *streamerClient) removeSubscription(sub *streamspb.SubscriptionID) error {
	c.clients[sub.Feature].sendC <- &streamspb.Request{
		Subscription: sub,
		Action:       streamspb.Action_REMOVE,
	}
	c.clients[sub.Feature].subscriptions.Remove(sub)
	return nil
}

func mustParseUrl(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		log.Fatal(err)
	}
	return u
}

func isReconnected(conn *grpc.ClientConn, check, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	ticker := time.NewTicker(check)

	for {
		select {
		case <-ticker.C:
			conn.Connect()

			if conn.GetState() == connectivity.Ready {
				return true
			}
		case <-ctx.Done():
			return false
		}
	}
}
