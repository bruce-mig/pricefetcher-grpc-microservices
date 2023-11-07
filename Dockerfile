FROM golang:1.21.3-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

COPY . ./

RUN go build -o /pricefetcher

EXPOSE 3000 4000

CMD ["/pricefetcher"]