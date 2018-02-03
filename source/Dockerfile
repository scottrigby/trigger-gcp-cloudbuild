FROM golang:alpine as build
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main . && cp main /tmp/

FROM scratch
COPY --from=build /tmp/main .
CMD ["./main"]
