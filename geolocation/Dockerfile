FROM golang:1.9 AS builder

# Install dep
RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x /usr/local/bin/dep

WORKDIR /go/src/github.com/thirtyspokes/docker-elk/geolocation
ADD . /go/src/github.com/thirtyspokes/docker-elk/geolocation
RUN dep ensure -vendor-only

# Build the app and its dependencies
ADD . /go/src/github.com/thirtyspokes/docker-elk/geolocation
WORKDIR /go/src/github.com/thirtyspokes/docker-elk/geolocation

RUN dep ensure -vendor-only
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Mount the output binary to an empty alpine image
FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /go/src/github.com/thirtyspokes/docker-elk/geolocation/app .
EXPOSE 8080
CMD ["./app"]