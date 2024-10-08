FROM golang:1.23.0-bookworm as build

ENV CGO_ENABLED=1

ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG BUILD=dev

RUN echo "Building on $BUILDPLATFORM, building for $TARGETPLATFORM"
WORKDIR /code

COPY . .
RUN go mod download
RUN go build -o eth-indexer-bootstrap -ldflags="-X main.build=${BUILD} -s -w" cmd/bootstrap/main.go
RUN go build -o /build/eth-indexer -ldflags="-X main.build=${BUILD} -s -w" cmd/service/*.go

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /service

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/* .
COPY migrations migrations/
COPY config.toml .
COPY queries.sql .
COPY LICENSE .

EXPOSE 5002

CMD ["./eth-indexer"]