# Build stage
FROM golang:1.20-alpine3.17 As builder
WORKDIR /app
COPY . .
RUN env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main main.go
# install curl 
RUN apk add curl
# download the migrate and extract the binary image and run it before starting the API server
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.2/migrate.linux-amd64.tar.gz | tar xvz

# Run stage
FROM alpine:3.13
WORKDIR /app
COPY --from=builder /app/main .
# copy from builder the download migrate binary to the final image 
COPY --from=builder /app/migrate ./migrate
# copy config file
COPY --from=builder /app/app.env .
# copy start.sh from project to /app/start.sh
COPY start.sh .
# copy all migrate files from the db/migration folder to the image /app/migration folder
COPY db/migration ./migration

EXPOSE 8080
# run db migration & start main app
# the CMD param will be passed to the ENTRYPOINT script to to run the migration and start the app
CMD ["/app/main"]
ENTRYPOINT [ "/app/start.sh" ] 
# result with : ENTRYPOINT [ "/app/start.sh", "/app/main" ]
