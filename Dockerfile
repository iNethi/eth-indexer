FROM golang:1.23.3-bookworm as build

ENV CGO_ENABLED=1

ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG BUILD=dev

RUN echo "Building on $BUILDPLATFORM, building for $TARGETPLATFORM"
WORKDIR /build

COPY . .
RUN go mod download
RUN go build -o inethi-indexer -ldflags="-X main.build=${BUILD} -s -w" cmd/service/*.go

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

CMD ["./inethi-indexer"]