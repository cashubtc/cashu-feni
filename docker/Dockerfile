FROM golang:1.19.1-alpine3.16 as builder

RUN mkdir /build

ADD . /build/

RUN apk --no-cache add git alpine-sdk build-base gcc

WORKDIR /build/cmd/cashu

RUN go build -o cashu  .

FROM alpine:latest

RUN adduser -S -D -H -h /app feni

COPY --from=builder /build/cmd/cashu/cashu /app/

RUN chown -R feni /app

USER feni

EXPOSE 3338

WORKDIR /app

CMD ["./cashu"]
