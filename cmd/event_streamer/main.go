package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"websocket-poc/internal/app/event_streamer"
	"websocket-poc/internal/pkg"
	"websocket-poc/pkg/streamspb"

	"google.golang.org/grpc"
)

const (
	defaultGrpcPort = 9090
	defaultHttpPort = 8080
	defaultFeature  = streamspb.Feature_SPOT_TICKER
)

func main() {
	grpcPort := pkg.GetPortFromEnv("GRPC_PORT", defaultGrpcPort)
	httpPort := pkg.GetPortFromEnv("HTTP_PORT", defaultHttpPort)

	feature := getFeatureFromEnv()
	log.Printf("starting streamer for feature %s\n", feature.String())

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", grpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("listening on %d\n", grpcPort)

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	srv := event_streamer.NewServer()
	go event_streamer.ProduceMessages(srv, feature)
	streamspb.RegisterEventStreamerServer(grpcServer, srv)

	go func() {
		http.Handle("/metrics", event_streamer.MetricsHandler())
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), nil))
	}()

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
