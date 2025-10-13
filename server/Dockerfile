# build stage
FROM golang:1.21-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /app/server

# copy module files first for caching
COPY server/go.mod server/go.sum* ./
RUN go mod download

# copy server source
COPY server/ ./

# build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /e5-server .

# final minimal image
FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /e5-server /e5-server
EXPOSE 8080
ENTRYPOINT ["/e5-server"]