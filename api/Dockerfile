# Use a Go version that satisfies your go.mod requirement
FROM golang:1.23.3-alpine as builder

WORKDIR /app
COPY . .
RUN go build -o app

FROM alpine
WORKDIR /app
COPY --from=builder /app/app .
EXPOSE 8080
CMD ["./app"]

