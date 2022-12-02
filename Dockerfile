FROM golang:1.19.1-alpine3.16 as builder

RUN mkdir /build

ADD . /build/

RUN apk --no-cache add git alpine-sdk build-base gcc

WORKDIR /build


RUN go build -o cashu cmd/mint/mint.go
RUN go build -o feni cmd/cashu/feni.go

FROM alpine:latest

RUN adduser -S -D -H -h /app feni

COPY --from=builder /build/cashu/mint /app/cashu
COPY --from=builder /build/feni /app/feni

RUN chown -R feni /app

USER feni

EXPOSE 3338

WORKDIR /app

CMD ["./cashu"]
