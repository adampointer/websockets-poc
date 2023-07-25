package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"websocket-poc/internal/app/websocket_adaptor"
	"websocket-poc/internal/pkg"

	"github.com/lesismal/nbio/nbhttp"
	"github.com/pkg/errors"
)

const defaultPort = 8080

func main() {
	port := pkg.GetPortFromEnv("HTTP_PORT", defaultPort)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := &http.ServeMux{}
	mux.HandleFunc("/ws", websocket_adaptor.OnWebsocket(ctx))

	svr := nbhttp.NewServer(nbhttp.Config{
		Network: "tcp",
		Addrs:   []string{fmt.Sprintf("localhost:%d", port)},
		Handler: mux,
	})

	err := svr.Start()
	if err != nil {
		log.Fatalf("nbio.Start failed: %v\n", err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	if err := svr.Shutdown(ctx); err != nil {
		log.Fatal(errors.Wrap(err, "shutdown GRPC server"))
	}
}
