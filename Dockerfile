# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS build
WORKDIR /go/src/app
ARG TARGETOS
ARG TARGETARCH

ARG BINARY_NAME

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-s -w" -o /go/bin/${BINARY_NAME} ./cmd/${BINARY_NAME}

# Run the binary
FROM --platform=$BUILDPLATFORM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=build /go/bin/${BINARY_NAME} /app/${BINARY_NAME}

CMD ["./${BINARY_NAME}"]
