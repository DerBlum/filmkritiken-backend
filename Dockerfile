FROM golang:latest as build

WORKDIR /go/src/app
COPY . .

RUN make build
RUN make test

FROM alpine:3.13.5

ENTRYPOINT ./main
COPY --from=build /go/src/app/main .
