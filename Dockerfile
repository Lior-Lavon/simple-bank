# Build stage
FROM golang:1.20-alpine3.17 As builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
# RUN env GOOS=linux GOARCH=arm64 go build -o main main.go

# Run stage
FROM alpine:3.13
WORKDIR /app
COPY --from=builder /app/main .
# copy config file
COPY --from=builder /app/app.env .
# copy start.sh from project to /app/start.sh
COPY start.sh .
# copy all migrate files from the db/migration folder to the image /app/migration folder
COPY db/migration ./db/migration

EXPOSE 8080
# run db migration & start main app
# the CMD param will be passed to the ENTRYPOINT script to to run the migration and start the app
CMD ["/app/main"]
ENTRYPOINT [ "/app/start.sh" ] 
# result with : ENTRYPOINT [ "/app/start.sh", "/app/main" ]
