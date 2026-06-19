FROM cr.shanhu.io/bld/base
LABEL author="Shanhu Tech Inc."

RUN <<EOF
#!/bin/sh

set -euo pipefail

adduser -D -u 3000 app
mkdir -p /opt/app
cd /opt/app/ && mkdir bin var etc lib tmp
chown -R app /opt/app
EOF
WORKDIR /opt/app
