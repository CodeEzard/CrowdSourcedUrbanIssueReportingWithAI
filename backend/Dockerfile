# Build a small Docker image for the backend
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
# Build the backend with verbose output so CI logs show which packages are compiled
# Stay in /src where go.mod is, and explicitly build the backend package
ENV CGO_ENABLED=0
ENV GOOS=linux
RUN go env && go list -m all
RUN go build -v -o /bin/backend ./backend

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
