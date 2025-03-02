# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build
WORKDIR /go/src/app
ARG TARGETOS
ARG TARGETARCH

ARG BINARY_NAME

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /go/bin/app ./cmd/${BINARY_NAME}

# Run the binary
FROM --platform=$BUILDPLATFORM oven/bun:debian
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/
WORKDIR /app

COPY --from=build /go/bin/app /app/app

EXPOSE 8089

CMD ["/app/app"]
