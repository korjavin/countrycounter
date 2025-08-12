# Build Stage
FROM golang:1.19-alpine AS builder

WORKDIR /app

# Copy backend source
COPY backend/ ./backend/
COPY frontend/ ./frontend/

# Build the Go app
# The go build command will automatically download dependencies.
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./backend

# Final Stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy frontend assets from the builder stage
COPY --from=builder /app/frontend ./frontend

EXPOSE 8080

CMD ["./main"]
