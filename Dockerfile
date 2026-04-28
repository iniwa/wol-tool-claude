# --- Build Stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o wol-server .

# --- Runtime Stage ---
FROM alpine:3.19

# ping コマンド + リモートシャットダウン用 samba-client
RUN apk add --no-cache iputils samba-client tzdata \
 && adduser -D -u 1000 wol

WORKDIR /app
COPY --from=builder /app/wol-server .
COPY static/ ./static/

# データ保存ディレクトリ
RUN mkdir -p /data && chown -R wol:wol /data /app
VOLUME ["/data"]

USER wol

ENV PORT=8080
ENV DATA_PATH=/data/devices.json

EXPOSE 8080

CMD ["./wol-server"]
