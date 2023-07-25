package event_streamer

import (
	"io"
	"log"

	"websocket-poc/internal/pkg"
	"websocket-poc/pkg/streamspb"
)

type Server struct {
	streamspb.UnimplementedEventStreamerServer
	reg *registry
}

func NewServer() *Server {
	return &Server{
		UnimplementedEventStreamerServer: streamspb.UnimplementedEventStreamerServer{},
		reg:                              newRegistry(),
	}
}

func (s *Server) Subscribe(stream streamspb.EventStreamer_SubscribeServer) error {
	log.Println("new stream started")
	state := pkg.NewSubscriptions()
	streamC := make(chan *streamspb.Response)
	s.reg.add(streamC)
	defer s.reg.remove(streamC)

	go func() {
		for res := range streamC {
			if !state.HasSubscription(res.Subscription) {
				continue
			}
			if err := stream.Send(res); err != nil {
				log.Fatal(err)
			}
		}
	}()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		switch req.Action {
		case streamspb.Action_ADD:
			log.Printf("adding %s\n", req.String())
			state.Add(req.Subscription)
		case streamspb.Action_REMOVE:
			log.Printf("removing %s\n", req.String())
			state.Remove(req.Subscription)
		}
	}
}

func (s *Server) BroadcastMessage(message *streamspb.Response) {
	s.reg.broadcast(message)
}
