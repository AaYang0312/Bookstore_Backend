# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG TARGETOS=linux
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/bookstore-manager ./cmd/bookstore-manager

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S bookstore && adduser -S -G bookstore bookstore

WORKDIR /app
COPY --from=builder /out/bookstore-manager ./bookstore-manager
COPY conf ./conf

USER bookstore
EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=5 \
  CMD wget -qO- http://127.0.0.1:8080/api/v1/category/list >/dev/null || exit 1

ENTRYPOINT ["./bookstore-manager"]
