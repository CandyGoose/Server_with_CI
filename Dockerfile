FROM golang:1.18-alpine

WORKDIR /app

COPY . .

RUN go build -o bin/app cmd/app/main.go

ENTRYPOINT ["./bin/app"]

EXPOSE 80/tcp
CMD ["app"]
