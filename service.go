package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// PriceService is an interface that can fetch the price for any given symbol.
type PriceService interface {
	FetchPrice(context.Context, string) ([]byte, error)
	// FetchPriceServerStreaming(context.Context, []string) error
}

type priceService struct{}

// This is the business logic
func (s *priceService) FetchPrice(_ context.Context, symbol string) ([]byte, error) {

	url := os.Getenv("API_URL") + "symbol=" + symbol + "&interval=1day&outputsize=30&format=json"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-RapidAPI-Key", os.Getenv("XRapidAPIKey"))
	req.Header.Add("X-RapidAPI-Host", os.Getenv("XRapidAPIHost2"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("the given symbol (%s) is not available", symbol)
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	return body, nil

}

type loggingService struct {
	next PriceService
}

func (s loggingService) FetchPrice(ctx context.Context, symbol string) (body []byte, err error) {
	defer func(begin time.Time) {
		reqID := ctx.Value(requestIDKey)

		logrus.WithFields(logrus.Fields{
			"requestID": reqID,
			"took":      time.Since(begin),
			"err":       err,
			// "body":      body,
			"symbol": symbol,
		}).Info("FetchPrice")
	}(time.Now())

	return s.next.FetchPrice(ctx, symbol)
}
