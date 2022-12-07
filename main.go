package main

import (
	"flag"
	"github.com/tonx22/lancktask/pkg/service"
	"github.com/tonx22/lancktask/pkg/transport"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var err error
	var (
		modeFlag   = flag.String("mode", "server", "application launch mode")
		searchFlag = flag.String("search", "", "phone numbers to search")
		typeFlag   = flag.String("type", "unary", "kinds of service method")
		tokenFlag  = flag.String("token", "ea207d02-d460-43e5-aa54-6a71658b03e4", "authorization token")
	)
	flag.Parse()

	// optional mode as client
	if *modeFlag == "client" {
		err = service.RunAsClient(searchFlag, typeFlag, tokenFlag)
		if err != nil {
			log.Printf("Runtime error: %v", err)
		}
		return
	}

	// default mode as server
	svc, err := service.NewSearchService()
	if err != nil {
		log.Fatalf("Can't create SearchService: %v", err)
	}

	err = transport.StartNewGRPCServer(svc, 50051, *tokenFlag)
	if err != nil {
		log.Fatalf("Failed to start GRPC server: %v", err)
	}

	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, os.Interrupt)
	<-sigChan
}
