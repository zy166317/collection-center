FROM golang:1.21 AS builder

WORKDIR /build

ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64
RUN	go env -w GOPROXY="https://goproxy.cn,direct";
#	go env -w GOPRIVATE="**.koblitzdigital.com**"; \
#	go env -w GONOPROXY="**.koblitzdigital.com**"; \
#    go env -w GONOSUMDB="**.koblitzdigital.com**"; \
#    echo "https://lib_read_only:lib_read_only@gitlab.koblitzdigital.com" >> ~/.git-credentials;\
#    git config --global credential.helper store
COPY go.mod .
COPY go.sum .
RUN go mod tidy

COPY . .
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -o collection-center .

FROM alpine:latest AS final
# TODO alpine 无法运行
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories \
    && apk --no-cache add tzdata ca-certificates libc6-compat libgcc libstdc++ \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone


WORKDIR /app
COPY --from=builder /build/collection-center /app/
COPY --from=builder /etc/passwd /etc/passwd

ENV GO_ENV=prod

CMD ["/app/collection-center"]
