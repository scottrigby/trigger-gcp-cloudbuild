FROM golang:alpine as build
WORKDIR /go/src/github.com/scottrigby/trigger-cgp-cloudbuild
ADD . .
RUN apk --no-cache add git ca-certificates && \
    go get -u github.com/golang/dep/... && \
    dep ensure -v --vendor-only && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main . && cp main /tmp/ && \
    cp -r source /tmp/ && \
    cp source.tgz /tmp/

FROM scratch
COPY --from=build /tmp/main .
COPY --from=build /tmp/source source
COPY --from=build /tmp/source.tgz .
# Fix error:
# > x509: failed to load system roots and no roots provided
# ref: https://medium.com/on-docker/use-multi-stage-builds-to-inject-ca-certs-ad1e8f01de1b
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
# @todo Keep the container running on scratch somehow without a shell wait loop.
CMD ["./main"]
