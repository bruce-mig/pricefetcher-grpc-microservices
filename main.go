package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/bruce-mig/pricefetcher-grpc-microservices/client"
	pb "github.com/bruce-mig/pricefetcher-grpc-microservices/proto"
)

func main() {
	var (
		jsonAddr = flag.String("jsonAddr", ":3000", "listen address of the json service")
		grpcAddr = flag.String("grpcAddr", ":4000", "listen address of the grpc service")
		svc      = loggingService{&priceService{}}
		ctx      = context.Background()
	)

	flag.Parse()

	grpcClient, err := client.NewGRPCClient(":4000")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		time.Sleep(3 * time.Second)
		resp, err := grpcClient.FetchPrice(ctx, &pb.PriceRequest{Ticker: "BTC"})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", resp)
	}()

	go makeGRPCServerAndRun(*grpcAddr, svc)

	jsonServer := NewJSONAPIServer(*jsonAddr, svc)
	jsonServer.Run()
}
