FROM cr.shanhu.io/bld/base
LABEL author="Shanhu Tech Inc."

RUN <<EOF
#!/bin/sh

set -euo pipefail

apk add --update git subversion mercurial
apk add --update gcc g++ musl-dev make openssh
EOF
COPY go.tgz /usr/local
COPY ssh_known_hosts /etc/ssh/known_hosts
RUN <<EOF
#!/bin/sh

set -euo pipefail

tar -xzf /usr/local/go.tgz -C /usr/local
rm /usr/local/go.tgz
EOF
ENV PATH=/usr/local/go/bin:/usr/sbin:/usr/bin:/sbin:/bin
RUN mkdir /go
ENV GOPATH=/go
WORKDIR /go

CMD ["/usr/local/go/bin/go", "version"]
