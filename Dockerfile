# One image for every component; docker-compose.yml runs each currency
# microservice and the gateway as separate containers.
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o /out/cryptomon ./cmd/cryptomon

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app
COPY --from=build /out/cryptomon /usr/local/bin/cryptomon
USER app
VOLUME /data
ENTRYPOINT ["cryptomon"]
