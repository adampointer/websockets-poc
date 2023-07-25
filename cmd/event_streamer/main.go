package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"websocket-poc/internal/app/event_streamer"
	"websocket-poc/internal/pkg"
	"websocket-poc/pkg/streamspb"

	"google.golang.org/grpc"
)

const (
	defaultPort    = 9090
	defaultFeature = streamspb.Feature_SPOT_TICKER
)

func main() {
	port := pkg.GetPortFromEnv("GRPC_PORT", defaultPort)
	feature := getFeatureFromEnv()
	log.Printf("starting streamer for feature %s\n", feature.String())

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	srv := event_streamer.NewServer()
	go event_streamer.ProduceMessages(srv, feature)
	streamspb.RegisterEventStreamerServer(grpcServer, srv)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("error starting server")
	}
}

func getFeatureFromEnv() streamspb.Feature {
	if strVal, found := os.LookupEnv("FEATURE"); found {
		switch strVal {
		case streamspb.Feature_SPOT_TICKER.String(): // SPOT_TICKER
			return streamspb.Feature_SPOT_TICKER
		case streamspb.Feature_SPOT_TRADE.String(): // SPOT_TRADE
			return streamspb.Feature_SPOT_TRADE
		}
	}
	return defaultFeature
}
