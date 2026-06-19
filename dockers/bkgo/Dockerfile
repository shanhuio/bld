FROM cr.shanhu.io/bld/golang
LABEL author="Shanhu Tech Inc."

RUN <<EOF
#!/bin/sh

set -euo pipefail

apk add --update bash
adduser -D -u 1000 forge
chown -R forge /go
EOF

USER 1000
