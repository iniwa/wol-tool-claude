# --- Build Stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o wol-server .

# --- Runtime Stage ---
FROM alpine:3.19

# ping コマンドのためのパッケージ
RUN apk add --no-cache iputils

WORKDIR /app
COPY --from=builder /app/wol-server .
COPY static/ ./static/

# データ保存ディレクトリ
VOLUME ["/data"]

ENV PORT=8080
ENV DATA_PATH=/data/devices.json

EXPOSE 8080

CMD ["./wol-server"]
