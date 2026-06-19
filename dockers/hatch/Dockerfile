FROM cr.shanhu.io/bld/base
LABEL author="Shanhu Tech Inc."

COPY idle.go /tmp/idle.go

RUN <<EOF
#!/bin/sh

set -euo pipefail

apk update && apk add --no-cache musl-dev bash go
cd /tmp && go build -o /bin/idle idle.go
EOF

CMD ["/bin/idle"]
