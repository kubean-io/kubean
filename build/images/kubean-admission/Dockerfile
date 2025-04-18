FROM --platform=$BUILDPLATFORM golang:1.22.4 as build

WORKDIR /kubean

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

COPY . .

ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -mod vendor -o kubean-admission ./cmd/kubean-admission/main.go

FROM alpine:3.19

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

COPY --from=build /kubean/kubean-admission /bin/

CMD ["/bin/kubean-admission"]
