FROM golang:1.19.1 as builder

RUN mkdir /build

ADD . /build/

WORKDIR /build

RUN go build -o cashu-feni .

FROM alpine

RUN adduser -S -D -H -h /app feni

USER feni

COPY --from=builder /build/cashu-feni /app/

WORKDIR /app

CMD ["./cashu-feni"]
