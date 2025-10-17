FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY tracks /usr/local/bin/tracks

ENTRYPOINT ["tracks"]
CMD ["--help"]
