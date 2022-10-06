FROM golang:1.19.1-alpine3.16 as builder

RUN mkdir /build

ADD . /build/
RUN apk --no-cache add git alpine-sdk build-base gcc

WORKDIR /build

RUN go build -o cashu  .

FROM alpine:latest

RUN adduser -S -D -H -h /app feni

USER feni

COPY --from=builder /build/cashu /app/

EXPOSE 3338

WORKDIR /app

CMD ["./cashu"]
