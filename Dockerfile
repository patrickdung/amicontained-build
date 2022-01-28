#FROM golang:alpine as builder
FROM docker.io/library/golang:alpine as builder

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	bash \
	ca-certificates

COPY . /go/src/github.com/genuinetools/amicontained

	#&& make static \
	#&& GOARCH=arm64 go build \
RUN set -eux \
	&& apk add --no-cache --virtual .build-deps \
		git \
		gcc \
		libc-dev \
		libgcc \
		make \
	&& cd /go/src/github.com/genuinetools/amicontained \
        && ls -lR /go \
	&& cp -p amicontained-build /usr/bin/amicontained \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

#FROM alpine:latest
FROM docker.io/library/alpine:latest

COPY --from=builder /usr/bin/amicontained /usr/bin/amicontained
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENTRYPOINT [ "amicontained" ]
#CMD [ "--help" ]
