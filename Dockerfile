FROM golang:1.22-alpine

# Install protoc and required packages
RUN apk add --no-cache protobuf protobuf-dev git

# Install protoc-gen-go and protoc-gen-go-grpc
RUN go get -u google.golang.org/protobuf/cmd/protoc-gen-go
RUN go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Generate Go code from protobuf
RUN protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           proto/notification.proto

# Build the Go app
RUN go build -o main .

# Start a new stage from scratch
FROM alpine:latest  

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Copy the .env file
COPY --from=builder /app/.env .

# Expose port 50052 to the outside world
EXPOSE 50052

# Command to run the executable
CMD ["./main"]