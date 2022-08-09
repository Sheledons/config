FROM gitlab.xunlei.cn/docker/xluser-centos:v1.2.0

ADD ./bin /eagle-eye/bin

WORKDIR /eagle-eye/bin

CMD ["./eagle-eye"]