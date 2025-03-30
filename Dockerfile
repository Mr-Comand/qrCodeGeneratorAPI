FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN go build -o qrcodeGenerator .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/qrcodeGenerator .

ENV LOG_LEVEL=DEBUG
ENV ANONYMIZE=true
ENV ALLOWED_CURRENCIES=EUR,USD,GBP,JPY,AUD,CAD
ENV DEFAULT_CURRENCY=EUR

EXPOSE 8080

CMD ["./qrcodeGenerator"]
