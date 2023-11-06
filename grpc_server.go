package main

import (
	"context"
	"math/rand"
	"net"

	pb "github.com/bruce-mig/pricefetcher-grpc-microservices/proto"
	"google.golang.org/grpc"
)

func makeGRPCServerAndRun(listenAddr string, svc PriceService) error {
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
	pb.UnimplementedPriceFetcherServer
}

func NewGRPCPriceFetcherServer(svc PriceService) *GRPCPriceFetcherServer {
	return &GRPCPriceFetcherServer{
		svc: svc,
	}
}

func (s *GRPCPriceFetcherServer) FetchPrice(ctx context.Context, req *pb.PriceRequest) (*pb.PriceResponse, error) {

	reqID := rand.Intn(100000)
	ctx = context.WithValue(ctx, "requestID", reqID)
	price, err := s.svc.FetchPrice(ctx, req.Ticker)
	if err != nil {
		return nil, err
	}

	resp := &pb.PriceResponse{
		Ticker: req.Ticker,
		Price:  float32(price),
	}
	return resp, err
}
