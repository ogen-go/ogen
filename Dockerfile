ARG GO_VERSION=latest
# Use the "static" distroless variant rather than "base". The ogen and
# jschemagen binaries are built with CGO disabled, so they are fully static
# and need no glibc. The "base" variant additionally ships libssl3, which we
# never link against; it showed up in a vulnerability scan (CVE-2026-31789),
# and "static" omits it entirely, reducing the image's attack surface.
ARG BASE_IMAGE=gcr.io/distroless/static-debian13

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
