FROM golang:1.22.3-bookworm as build

# Should be 0 unless we import celo-blockchain or related libraries
ENV CGO_ENABLED=0

ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG BUILD=dev

RUN echo "Building on $BUILDPLATFORM, building for $TARGETPLATFORM"
WORKDIR /build

COPY . .
RUN go mod download
RUN go build -o celo-indexer -ldflags="-X main.build=${BUILD} -s -w" cmd/*

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /service

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/* .

EXPOSE 5001

CMD ["./celo-indexer"]