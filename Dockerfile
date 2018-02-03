FROM golang:alpine as build
WORKDIR /go/src/github.com/scottrigby/trigger-cgp-cloudbuild
ADD . .
RUN apk --no-cache add git && \
    go get -u github.com/golang/dep/... && \
    dep ensure -v --vendor-only && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main . && cp main /tmp/ && \
    cp -r source /tmp/

FROM scratch
COPY --from=build /tmp/main .
COPY --from=build /tmp/source source
# @todo Keep the container running on scratch somehow without a shell wait loop.
CMD ["./main"]
