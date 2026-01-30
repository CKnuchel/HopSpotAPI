# Build Stage
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

# Diese Args werden von Docker automatisch gesetzt
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Mod Download (Cacheing)
COPY go.mod go.sum ./
RUN go mod download

# Copy Source Code
COPY . .

# Build Application für Ziel-Plattform
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" \
    -o server ./cmd/server

# Runtime Stage
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /app/server /app/server

USER nonroot:nonroot

ENTRYPOINT ["/app/server"]