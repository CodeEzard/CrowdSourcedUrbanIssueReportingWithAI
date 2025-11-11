# Build a small Docker image for the backend
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
WORKDIR /src/backend
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/backend ./

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
