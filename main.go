package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/bruce-mig/pricefetcher-grpc-microservices/client"
	pb "github.com/bruce-mig/pricefetcher-grpc-microservices/proto"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
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
		resp, err := grpcClient.FetchPrice(ctx, &pb.PriceRequest{Symbol: "MSFT"})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%+v\n", resp)
	}()

	go func() {
		time.Sleep(3 * time.Second)
		log.Print("Streaming started")
		symbols := &pb.SymbolsList{
			Symbols: []*pb.PriceRequest{
				{Symbol: "AAPL"},
				{Symbol: "GOOG"},
				{Symbol: "MSFT"},
			},
		}
		stream, err := grpcClient.FetchPriceServerStreaming(ctx, symbols)
		if err != nil {
			log.Fatalf("Could not send symbols: %v", err)
		}

		for {
			message, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while streaming %v", err)
			}
			log.Println(message)
		}
		log.Printf("Streaming finished")
	}()

	go makeGRPCServerAndRun(*grpcAddr, svc)

	jsonServer := NewJSONAPIServer(*jsonAddr, svc)
	jsonServer.Run()
}
