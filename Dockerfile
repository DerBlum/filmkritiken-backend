FROM alpine:latest

WORKDIR /app

ENTRYPOINT /app/main
COPY output/main .

USER root
RUN chmod a+rw -R .
RUN chmod +x main
USER 1001
