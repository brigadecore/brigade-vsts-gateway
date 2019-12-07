FROM quay.io/deis/lightweight-docker-go:v0.6.0 AS build
ENV CGO_ENABLED=0
WORKDIR /go/src/github.com/radu-matei/brigade-vsts-gateway
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY vendor/ vendor/
RUN go build -o bin/vsts-gateway ./cmd

# I know latest isn't a version... but this is a go app.
FROM alpine:latest AS final
COPY --from=build /go/src/github.com/radu-matei/brigade-vsts-gateway/bin/vsts-gateway /usr/bin/gateway
CMD ["/usr/bin/gateway"]
EXPOSE 8080
