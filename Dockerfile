FROM golang:1.24 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0

RUN go build -o /currency cmd/currency.go

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /currency /currency

CMD ["/currency"]