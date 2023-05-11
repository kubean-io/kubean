FROM golang:1.20.4 as build

WORKDIR /kubean

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor -o kubean-operator ./cmd/kubean-operator/main.go

FROM alpine:3.15

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

COPY --from=build /kubean/kubean-operator /bin/

CMD ["/bin/kubean-operator"]
