FROM golang:1.14 AS builder
ENV GOPROXY https://goproxy.io
ENV CGO_ENABLED 0
WORKDIR /go/src/app
ADD . .
RUN go build -mod vendor -o /enforce-deployment-max-pods

FROM alpine:3.12
COPY --from=builder /enforce-deployment-max-pods /enforce-deployment-max-pods
CMD ["/enforce-deployment-max-pods"]