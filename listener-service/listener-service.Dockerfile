FROM golang:1.25.3-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o listenerApp ./main.go 

RUN chmod +x /app/listenerApp

# build tiny docker image
FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/listenerApp /app

CMD ["/app/listenerApp"]
