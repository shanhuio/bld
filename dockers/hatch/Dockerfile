FROM cr.shanhu.io/lib/base
MAINTAINER Shanhu Tech Inc.

RUN apk update && apk add --no-cache musl-dev bash go
RUN mkdir /hatch
COPY idle.go /hatch/idle.go
RUN cd /hatch && go build idle.go
WORKDIR /hatch
CMD ["/hatch/idle"]
