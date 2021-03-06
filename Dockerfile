FROM alpine:3.7

RUN apk add --no-cache ca-certificates && update-ca-certificates
COPY bin/gateway /usr/bin/gateway
EXPOSE 8080

CMD /usr/bin/gateway