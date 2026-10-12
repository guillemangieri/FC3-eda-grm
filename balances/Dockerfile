FROM golang:1.26
WORKDIR /app
RUN apt-get update && apt-get install -y librdkafka-dev
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY pkg ./pkg
RUN CGO_ENABLED=1 go build -o /usr/local/bin/balances ./cmd/balances
EXPOSE 3003
CMD ["/usr/local/bin/balances"]
