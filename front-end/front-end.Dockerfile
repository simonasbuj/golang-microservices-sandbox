FROM golang:1.25.3-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o frontendApp ./cmd/web/main.go 

RUN chmod +x /app/frontendApp

# build tiny docker image
FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/frontendApp /app
COPY --from=builder /app/cmd/web/templates ./cmd/web/templates

CMD ["/app/frontendApp"]
