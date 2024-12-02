ARG GO_VERSION=latest

FROM golang:$GO_VERSION as builder

WORKDIR /go/src/app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -o /go/bin/ogen ./cmd/ogen/main.go

FROM golang:1.23.3

WORKDIR /

COPY --from=builder /go/bin/ogen ./ogen

ENTRYPOINT ["./ogen"]
