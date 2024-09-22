# Stage 1: Build the Go binary
FROM golang:1.23.1 AS builder
WORKDIR /go/src/github.com/worker
COPY main.go go.mod go.sum ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Stage 2: Create a lightweight image with the built binary
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/github.com/worker/app .
CMD ["./app"]
