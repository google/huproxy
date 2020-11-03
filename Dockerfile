FROM golang:latest
WORKDIR /go/src/github.com/google/huproxy
COPY . .
RUN mkdir /app
RUN go get -d -v .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /app .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o /app ./huproxyclient

FROM alpine:latest
WORKDIR /
COPY --from=0 /app/ .
CMD ["/huproxy"]
