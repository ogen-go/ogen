FROM golang:1.21 as builder

WORKDIR /go/src/app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 go build -o /go/bin/ogen ./cmd/ogen/main.go

FROM scratch

WORKDIR /

COPY --from=builder /go/bin/ogen ./ogen

ENTRYPOINT ["./ogen"]
