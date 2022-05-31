FROM golang:latest as BUILDER

MAINTAINER zengchen1024<chenzeng765@gmail.com>

# build binary
WORKDIR /go/src/github.com/opensourceways/robot-gitee-review-trigger
COPY . .
RUN GO111MODULE=on CGO_ENABLED=0 go build -a -o robot-gitee-review-trigger .

# copy binary config and utils
FROM alpine:3.14
COPY  --from=BUILDER /go/src/github.com/opensourceways/robot-gitee-review-trigger/robot-gitee-review-trigger /opt/app/robot-gitee-review-trigger

ENTRYPOINT ["/opt/app/robot-gitee-review-trigger"]
