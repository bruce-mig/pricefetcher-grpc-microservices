package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bruce-mig/pricefetcher-grpc-microservices/types"
)

type APIFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

func makeAPIFunc(fn APIFunc) http.HandlerFunc {
	ctx := context.Background()
	return func(w http.ResponseWriter, r *http.Request) {
		ctx = context.WithValue(ctx, requestIDKey, reqID)

		if err := fn(ctx, w, r); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		}
	}
}

type JSONAPIServer struct {
	listenAddr string
	svc        PriceService
}

// constructor
func NewJSONAPIServer(listenAddr string, svc PriceService) *JSONAPIServer {
	return &JSONAPIServer{
		svc:        svc,
		listenAddr: listenAddr,
	}
}

func (s *JSONAPIServer) Run() {
	http.HandleFunc("/", makeAPIFunc(s.handleFetchPrice))
	http.ListenAndServe(s.listenAddr, nil)
}

func (s *JSONAPIServer) handleFetchPrice(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	symbol := r.URL.Query().Get("symbol")
	if len(symbol) == 0 {
		return fmt.Errorf("invalid ticker")
	}
	body, err := s.svc.FetchPrice(ctx, symbol)
	if err != nil {
		return err
	}

	var result types.PriceResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}

	resp := types.PriceResponse{
		Symbol:        result.Symbol,
		Name:          result.Name,
		DateTime:      result.DateTime,
		Price:         result.Price,
		PercentChange: result.PercentChange,
	}
	return writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, s int, v any) error {
	w.WriteHeader(s)
	return json.NewEncoder(w).Encode(v)
}
