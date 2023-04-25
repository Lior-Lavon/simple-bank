# Build stage
FROM golang:1.20-alpine3.17 As builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# Run stage
FROM alpine:3.13
WORKDIR /app
COPY --from=builder /app/main .
COPY --from=builder /app/app.env .

EXPOSE 8080
CMD ["/app/main"]
