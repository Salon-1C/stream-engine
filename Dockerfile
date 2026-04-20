# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o server ./cmd/server

# ── Stage 2: runtime ──────────────────────────────────────────────────────────
FROM alpine:3.19

RUN addgroup -S stream && adduser -S -G stream stream

WORKDIR /app

COPY --from=build /app/server .
# Static web demo served by the Go server
COPY web/ ./web/
# mediamtx.yml is baked into the image so ECS doesn't need an EFS mount.
# The mediamtx container reads its config via MTX_* env vars injected by ECS,
# so this copy is only useful when running locally without the sidecar.
COPY configs/ ./configs/

RUN chown -R stream:stream /app
USER stream

EXPOSE 8080

ENTRYPOINT ["./server"]
