FROM golang:alpine as build

RUN apk --no-cache add make git gcc libtool musl-dev ca-certificates dumb-init 

WORKDIR /go/src/app
COPY . .

RUN make build
#RUN make test

FROM alpine:latest

ENTRYPOINT ./main
COPY --from=build /go/src/app/main .
