FROM golang:1.9

WORKDIR /go/src/github.com/coldog/logship
COPY . .
ENTRYPOINT ["/go/src/github.com/coldog/logship/builder.sh"]
