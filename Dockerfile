FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY logship /bin/logship
ENTRYPOINT ["/bin/logship"]
