# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /gateway ./cmd/gateway

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /

COPY --from=builder /gateway /gateway

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/gateway"]
