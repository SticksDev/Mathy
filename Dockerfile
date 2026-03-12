FROM golang:1.26-alpine AS builder

WORKDIR /app

# CGO build tools
RUN apk add --no-cache \
    git \
    build-base

# copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# copy project files and build
COPY . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -x -v  -o bot .


# runner stage
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache \
    ca-certificates \
    libgcc \
    libstdc++

COPY --from=builder /app/bot .

RUN adduser -D botuser
USER botuser

ENTRYPOINT ["./bot"]