# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/finwa ./cmd/bot

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -H -u 10001 finwa

COPY --from=builder /out/finwa /usr/local/bin/finwa

WORKDIR /app
COPY migrations ./migrations

RUN mkdir -p /app/data && chown -R finwa:finwa /app
USER finwa

ENV TZ=Asia/Jakarta \
    APP_ENV=production

CMD ["finwa"]
