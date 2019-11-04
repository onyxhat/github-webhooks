FROM golang:1.13.4-alpine3.10

RUN apk add git

WORKDIR /go/src/app
COPY main.go .
RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8080
CMD ["app"]
