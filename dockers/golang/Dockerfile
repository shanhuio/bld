FROM cr.shanhu.io/lib/base
MAINTAINER Shanhu Tech Inc.

RUN apk add --update git subversion mercurial
RUN apk add --update gcc g++ musl-dev make openssh
COPY go.tgz /usr/local
COPY ssh_known_hosts /etc/ssh/known_hosts
RUN tar -xzf /usr/local/go.tgz -C /usr/local
RUN rm /usr/local/go.tgz
ENV PATH /usr/local/go/bin:/usr/sbin:/usr/bin:/sbin:/bin
RUN mkdir /go
ENV GOPATH /go
WORKDIR /go

CMD ["/usr/local/go/bin/go", "version"]
