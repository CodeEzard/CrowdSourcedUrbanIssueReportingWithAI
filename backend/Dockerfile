# Build a small Docker image for the backend
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . ./
# Diagnostic: confirm main.go exists and has package main + func main
RUN echo "=== Checking /src/backend ===" && \
    ls -la /src/backend/ | grep -E "main|router" && \
    echo "=== Checking main.go content ===" && \
    head -30 /src/backend/main.go && \
    echo "=== Checking for other main packages ===" && \
    find /src -name "*.go" -type f -exec grep -l "^package main$" {} \;
# Build the backend - try without mod=readonly to allow Go full flexibility
RUN CGO_ENABLED=0 GOOS=linux go build -v -o /bin/backend ./backend

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/backend /usr/local/bin/backend
# Copy frontend assets into the image so the backend can serve them in production
COPY --from=build /src/frontend /app/frontend
WORKDIR /app
ENV PORT=8080
ENV FRONTEND_DIR=/app/frontend
EXPOSE 8080
CMD ["/usr/local/bin/backend"]
