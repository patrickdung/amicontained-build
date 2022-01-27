FROM golang:alpine as builder
#MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	bash \
	ca-certificates

COPY . /go/src/github.com/genuinetools/amicontained

	# && make static \
RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		git \
		gcc \
		libc-dev \
		libgcc \
		make \
	&& cd /go/src/github.com/genuinetools/amicontained \
	&& GOARCH=arm64 go build \
	&& mv amicontained-build /usr/bin/amicontained \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

FROM alpine:latest

COPY --from=builder /usr/bin/amicontained /usr/bin/amicontained
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENTRYPOINT [ "amicontained" ]
CMD [ "--help" ]
