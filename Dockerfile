ARG GO_VERSION=latest

FROM golang:$GO_VERSION AS builder
WORKDIR /go/src/app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/ogen ./cmd/ogen/main.go

FROM scratch
# We need go in resulting image to run goimports.
COPY --from=builder /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"
# Copy built binary.
WORKDIR /
COPY --from=builder /go/bin/ogen ./ogen
ENTRYPOINT ["./ogen"]
