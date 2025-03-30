# Build Stage
FROM golang:1.23.0 AS builder

# Set the working directory
WORKDIR /app

# Copy the source code
COPY . .
# Download the Go module dependencies
RUN go mod download
# Build the Go application for Linux
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o qrcodeGenerator main.go

# Final Stage
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/qrcodeGenerator .

# Expose the port your application runs on (adjust if necessary)
EXPOSE 8080
ENV LOG_LEVEL=DEBUG
ENV ANONYMIZE=true
ENV ALLOWED_CURRENCIES=EUR,USD,GBP,JPY,AUD,CAD
ENV DEFAULT_CURRENCY=EUR
# Command to run the executable
CMD ["./qrcodeGenerator"]