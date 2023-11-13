package main

import (
	"flag"
	"log"
	"sync"

	"github.com/bruce-mig/pricefetcher-grpc-microservices/client"
	"github.com/joho/godotenv"
)

var wg *sync.WaitGroup

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	var (
		jsonAddr = flag.String("jsonAddr", ":3000", "listen address of the json service")
		grpcAddr = flag.String("grpcAddr", ":4000", "listen address of the grpc service")
		svc      = loggingService{&priceService{}}
	)

	flag.Parse()

	// grpcClient, err :=
	client.NewGRPCClient(":4000")
	if err != nil {
		log.Fatal(err)
	}

	// wg = &sync.WaitGroup{}

	// symbols := &pb.SymbolsList{
	// 	Symbols: []*pb.PriceRequest{
	// 		{Symbol: "AAPL"},
	// 		{Symbol: "GOOG"},
	// 		{Symbol: "MSFT"},
	// 	},
	// }

	// // wg.Add(1)
	// go client.CallFetchPrice(wg, grpcClient, symbols)

	// wg.Add(1)
	// go client.CallFetchPriceServerStreaming(wg, grpcClient, symbols)

	// // wg.Add(1)
	// // // go client.CallFetchPriceBidirectionalStreaming(wg, grpcClient, symbols)

	go makeGRPCServerAndRun(wg, *grpcAddr, svc)

	jsonServer := NewJSONAPIServer(*jsonAddr, svc)
	jsonServer.Run()
	// wg.Wait()
}
