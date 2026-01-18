# syntax=docker/dockerfile:1

FROM golang:1.24.4-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src

# Keep the build lean and reproducible.
RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
  go build -trimpath -ldflags="-s -w" -o /out/megabudget ./cmd/megabudget

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /out/megabudget /megabudget
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/megabudget"]
