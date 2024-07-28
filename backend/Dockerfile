# Build stage
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# Copy only necessary source files
COPY api/ ./api/
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY migrations/ ./migrations/

COPY main.go ./

ARG VERSION=dev

# Create a non-root user
RUN echo "nonroot:x:65532:65532:nonroot:/:" > /tmp/passwd && \
    echo "nonroot:x:65532:" > /tmp/group

# Build the application
RUN CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o smarty-pants-ai .

# Runtime stage
FROM scratch

# Copy the passwd and group files
COPY --from=builder /tmp/passwd /tmp/group /etc/

# Copy SSL certificates
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the built application
COPY --from=builder /backend/smarty-pants-ai /app/smarty-pants-ai

# Use the non-root user
USER nonroot:nonroot

# Set the working directory
WORKDIR /app

# Run the binary
CMD ["/app/smarty-pants-ai", "start"]