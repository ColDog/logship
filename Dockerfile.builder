FROM golang:1.9-alpine

RUN apk add --no-cache ca-certificates
WORKDIR /go/src/github.com/coldog/logship
COPY . .
RUN go install github.com/coldog/logship
ENTRYPOINT ["/go/src/github.com/coldog/logship/builder.sh"]
