# Start from the official Golang image to build the binary
FROM golang:1.24.3-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Start a new stage from scratch for a lightweight image
FROM alpine:3.21.3

RUN apk update && apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Copy the model.conf and policy.csv files
COPY --from=builder /app/model.conf .
COPY --from=builder /app/policy.csv .

# Copy the template directory to the same directory as the Go app
COPY --from=builder /app/templates ./templates

# Expose port 3000 to the outside world
EXPOSE 3000

# Command to run the executable
CMD ["./main"] 