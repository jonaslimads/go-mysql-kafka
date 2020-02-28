FROM buildpack-deps:buster-scm

# gcc for cgo
RUN apt-get update && apt-get install -y --no-install-recommends \
		g++ \
		gcc \
		libc6-dev \
		make \
		pkg-config \
		curl \
		git \
	&& rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION=1.14 \
    GOROOT=/var/mysql-cdc/goroot \
    GOPATH=/var/mysql-cdc/go \
    GOSRC=/var/mysql-cdc/src \
    GO111MODULE=on \
    ENABLE_HOT_RELOAD=false

ENV PATH="$GOPATH/bin:$GOROOT/bin:$PATH"

VOLUME $GOPATH

VOLUME $GOROOT

WORKDIR $GOSRC

COPY mysql-cdc/*.sh /tmp/

RUN mv -f /tmp/*.sh /usr/local/bin/ \
    && chmod +x /usr/local/bin/*

ENTRYPOINT "docker-entrypoint.sh"