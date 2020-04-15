# Build intchain in a stock Go builder container
FROM golang:1.12-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . /intchain
RUN cd /intchain && make intchain

# Pull Intchain into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /intchain/bin/intchain /usr/local/bin/

EXPOSE 8556 8555 8554 8553 8552 8551 8550 8550/udp
ENTRYPOINT ["intchain"]
