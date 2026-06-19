FROM cr.shanhu.io/bld/golang
LABEL author="Shanhu Tech Inc."

RUN mkdir /usr/local/idle
COPY idle.go /usr/local/idle/idle.go
RUN <<EOF
#!/bin/sh

set -euo pipefail

cd /usr/local/idle && /usr/local/go/bin/go build idle.go
rm /usr/local/idle/idle.go
EOF

CMD ["/usr/local/idle/idle"]
