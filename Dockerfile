# Build Stage
FROM golang:1.24-alpine AS builder

# Add build argument for commit SHA
ARG COMMIT_SHA=unknown

WORKDIR /app

# Copy backend source
COPY backend/ ./backend/
COPY frontend/ ./frontend/

WORKDIR /app/backend

RUN go mod tidy

# Build the Go app with version info
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.commitSHA=${COMMIT_SHA}" -o /app/main .

# Final Stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

RUN mkdir -p backend data

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy frontend assets from the builder stage
COPY --from=builder /app/frontend ./frontend

EXPOSE 8080

CMD ["./main"]
