# ---- Build stage ----
FROM golang:1.24-alpine AS build

WORKDIR /app

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /sabacc-server .

# ---- Runtime stage ----
FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=build /sabacc-server /sabacc-server

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/sabacc-server"]
