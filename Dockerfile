ARG GO_VERSION=latest
ARG BASE_IMAGE=gcr.io/distroless/base-debian12

FROM golang:${GO_VERSION} AS builder
WORKDIR /go/src/app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/ogen ./cmd/ogen/main.go
RUN CGO_ENABLED=0 go build -o /go/bin/jschemagen ./cmd/jschemagen/main.go

FROM ${BASE_IMAGE}
# We need go in resulting image to run goimports.
COPY --from=builder /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"
# Copy built binary.
WORKDIR /
COPY --from=builder /go/bin/ogen /ogen
COPY --from=builder /go/bin/jschemagen /jschemagen
ENTRYPOINT ["/ogen"]
