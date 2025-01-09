# Build stage
FROM golang:1.22.3 AS builder
WORKDIR /app
COPY build/main .

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/main .
RUN chmod +x /app/main
EXPOSE 8080
CMD ["./main"]