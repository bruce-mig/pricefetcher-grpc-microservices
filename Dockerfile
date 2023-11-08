FROM golang:1.21-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

COPY . ./

RUN go build -o /price

EXPOSE 3000 4000

CMD ["/price"]