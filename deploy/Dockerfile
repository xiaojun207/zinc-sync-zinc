#build源镜像
FROM golang:1.19-alpine as build
#作者
MAINTAINER xiaojun "xiaojun207@126.com"

#ENV GOPROXY https://goproxy.io
ENV GO111MODULE on

WORKDIR /go/release

ADD go.mod .
ADD go.sum .
RUN go mod download
RUN go mod tidy

WORKDIR /go/release

ADD . .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags '-w -s' -o App main.go
#RUN go build -a -ldflags "-linkmode external -extldflags '-static' -s -w " -o App main.go

#运行镜像
FROM alpine:latest AS production

RUN mkdir /app
WORKDIR /app

COPY --from=build /go/release/App /app/
COPY --from=build /go/release/deploy/entrypoint.sh /app/


ENTRYPOINT ["sh", "/app/entrypoint.sh"]





