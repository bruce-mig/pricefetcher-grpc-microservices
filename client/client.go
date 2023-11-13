package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	pb "github.com/bruce-mig/pricefetcher-grpc-microservices/proto"
	"github.com/bruce-mig/pricefetcher-grpc-microservices/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ctx, _ = context.WithTimeout(context.Background(), 3000*time.Second)

func NewGRPCClient(remoteAddr string) (pb.PriceFetcherClient, error) {
	conn, err := grpc.Dial(remoteAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	c := pb.NewPriceFetcherClient(conn)

	return c, nil
}

// gRPC Client
func CallFetchPrice(wg *sync.WaitGroup, grpcClient pb.PriceFetcherClient, symbols *pb.SymbolsList) {
	// defer wg.Done()
	time.Sleep(3 * time.Second)
	resp, err := grpcClient.FetchPrice(ctx, symbols.Symbols[0])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", resp)
}

func CallFetchPriceServerStreaming(wg *sync.WaitGroup, grpcClient pb.PriceFetcherClient, symbols *pb.SymbolsList) {
	// defer wg.Done()
	log.Print("Streaming started")

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
}

func CallFetchPriceBidirectionalStreaming(wg *sync.WaitGroup, grpcClient pb.PriceFetcherClient, symbols *pb.SymbolsList) {
	defer wg.Done()
	log.Printf("Bidirectional Streaming Started")
	stream, err := grpcClient.FetchPriceBidirectionalStreaming(ctx)
	if err != nil {
		log.Fatalf("Could not send symbols: %v", err)
	}
	waitCh := make(chan struct{})
	go func() {
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
		close(waitCh)
	}()

	for _, symbol := range symbols.Symbols {
		req := &pb.PriceRequest{
			Symbol: symbol.Symbol,
		}
		if err := stream.Send(req); err != nil {
			log.Fatalf("Error while sending %v", err)
		}
		time.Sleep(2 * time.Second)
	}
	stream.CloseSend()
	<-waitCh
	log.Printf("Bidirectional streaming finished")
}

// JSON client
type Client struct {
	endpoint string
}

func New(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
	}
}

func (c *Client) FetchPrice(ctx context.Context, symbol string) (*types.PriceResponse, error) {
	endpoint := fmt.Sprintf("%s?symbol=%s", c.endpoint, symbol)

	req, err := http.NewRequest("get", endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		httpErr := map[string]any{}
		if err := json.NewDecoder(resp.Body).Decode(&httpErr); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("service responded with non OK status code: %s", httpErr["error"])
	}

	priceResp := new(types.PriceResponse)
	if err := json.NewDecoder(resp.Body).Decode(priceResp); err != nil {
		return nil, err
	}

	return priceResp, nil
}
