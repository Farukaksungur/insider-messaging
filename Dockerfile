FROM golang:1.21-alpine AS build
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /insider ./cmd/app

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /insider /insider
EXPOSE 8080
ENTRYPOINT ["/insider"]
