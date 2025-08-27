# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy file go.mod and go.sum 
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build aplikasi
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# port yang digunakan oleh product service
EXPOSE 8081

# Jalankan / Run binary
CMD ["./main"]