FROM --platform=$BUILDPLATFORM golang:1.26 AS builder
ARG  TARGETOS
ARG  TARGETARCH
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH}  go build -ldflags="-s -w" -o slipstream .

FROM gcr.io/distroless/static-debian13:latest-arm64
WORKDIR /app
COPY --from=builder /app/slipstream .
USER nonroot:nonroot
CMD ["./slipstream"]
