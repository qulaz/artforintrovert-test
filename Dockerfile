FROM golang:1.19.1-alpine AS builder

WORKDIR /app

ENV GO111MODULE="on"
ENV GOOS="linux"
ENV CGO_ENABLED=0

# System dependencies
RUN apk update \
    && apk add --no-cache \
    ca-certificates \
    git \
    && update-ca-certificates

# Install project deps
COPY ./go.mod ./go.sum /app/
RUN go mod download && go mod verify

COPY . /app

RUN go build -o server -a ./cmd/app/main.go
RUN go build -o populate -a ./cmd/populate/main.go

### Runtime
FROM alpine:latest as runtime

RUN apk update \
    && apk add --no-cache \
    ca-certificates \
    curl \
    tzdata \
    && update-ca-certificates

# Copy executable
COPY --from=builder /app/server /usr/local/bin/server
COPY --from=builder /app/populate /usr/local/bin/populate

CMD ["/usr/local/bin/server"]
