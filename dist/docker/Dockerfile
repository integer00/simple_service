FROM golang:alpine as builder
WORKDIR /tmp
COPY main.go .
RUN go build -o service .

FROM alpine
WORKDIR /root/
COPY --from=builder /tmp/service .

EXPOSE 8080
CMD ["./service", "-listen", "0.0.0.0:8080"]