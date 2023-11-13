package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	pb "github.com/bruce-mig/pricefetcher-grpc-microservices/proto"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

// Define a custom type for context keys
type ctxKey string

// Define a constant for the request ID key
const requestIDKey ctxKey = "requestID"

func makeGRPCServerAndRun(wg *sync.WaitGroup, listenAddr string, svc PriceService) error {
	defer wg.Done()
	grpcPriceFetcher := NewGRPCPriceFetcherServer(svc)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	pb.RegisterPriceFetcherServer(server, grpcPriceFetcher)
	return server.Serve(ln)

}

type GRPCPriceFetcherServer struct {
	svc PriceService
	pb.PriceFetcherServer
	// pb.UnimplementedPriceFetcherServer
}

func NewGRPCPriceFetcherServer(svc PriceService) *GRPCPriceFetcherServer {
	return &GRPCPriceFetcherServer{
		svc: svc,
	}
}

func (s *GRPCPriceFetcherServer) FetchPrice(ctx context.Context, req *pb.PriceRequest) (*pb.PriceResponse, error) {

	// Set the request ID in the context using the custom key type
	reqID := rand.Intn(10000)
	ctx = context.WithValue(ctx, requestIDKey, reqID)

	body, err := s.svc.FetchPrice(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}

	data := &pb.PriceResponse{}
	unmarshaler := protojson.UnmarshalOptions{DiscardUnknown: true}
	err = unmarshaler.Unmarshal(body, data)

	if err != nil {
		fmt.Printf("failed to unmarshal:%+v", err)
		return nil, err
	}

	resp := &pb.PriceResponse{
		Symbol:        data.Symbol,
		Name:          data.Name,
		Datetime:      data.Datetime,
		Close:         data.Close,
		PercentChange: data.PercentChange,
	}
	return resp, err
}

func (s *GRPCPriceFetcherServer) FetchPriceServerStreaming(req *pb.SymbolsList, stream pb.PriceFetcher_FetchPriceServerStreamingServer) error {
	log.Printf("Got request with symbols: %v", req.Symbols)
	ctx := context.Background()
	for _, symbol := range req.Symbols {
		//req ID
		reqID := rand.Intn(10000)
		ctx = context.WithValue(ctx, requestIDKey, reqID)

		resp, err := s.FetchPrice(ctx, symbol)
		if err != nil {
			return err
		}
		if err = stream.Send(resp); err != nil {
			return err
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}

func (s *GRPCPriceFetcherServer) FetchPriceBidirectionalStreaming(stream pb.PriceFetcher_FetchPriceBidirectionalStreamingServer) error {
	ctx := context.Background()
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Printf("Got request with symbol:%v", req.Symbol)

		//req ID
		reqID := rand.Intn(10000)
		ctx = context.WithValue(ctx, requestIDKey, reqID)

		resp, err := s.FetchPrice(ctx, req)
		if err != nil {
			return err
		}
		if err = stream.Send(resp); err != nil {
			return err
		}
	}
}
