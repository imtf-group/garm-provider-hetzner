FROM docker.io/golang:alpine

WORKDIR /root
USER root

RUN apk add musl-dev gcc libtool m4 curl autoconf g++ make libblkid util-linux-dev git linux-headers mingw-w64-gcc

RUN wget https://github.com/imtf-group/garm-provider-hetzner/releases/download/0.0.0/aarch64-linux-musl-cross.tgz -O /tmp/aarch64-linux-musl-cross.tgz && \
    tar --strip-components=1 -C /usr/local -xzf /tmp/aarch64-linux-musl-cross.tgz && \
    rm /tmp/aarch64-linux-musl-cross.tgz

ADD ./scripts/build-static.sh /build-static.sh
RUN chmod +x /build-static.sh

CMD ["/bin/sh"]
