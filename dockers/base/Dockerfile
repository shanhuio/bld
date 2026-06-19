FROM cr.shanhu.io/bld/alpine
LABEL author="Shanhu Tech Inc."

RUN <<EOF
#!/bin/sh

set -euo pipefail

apk update
apk add ca-certificates && update-ca-certificates
EOF
