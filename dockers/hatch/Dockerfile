FROM cr.shanhu.io/bld/base
LABEL author="Shanhu Tech Inc."

RUN <<EOF
#!/bin/sh

set -euo pipefail

apk update && apk add --no-cache musl-dev bash go
mkdir /hatch
EOF
COPY idle.go /hatch/idle.go
RUN cd /hatch && go build idle.go
WORKDIR /hatch
CMD ["/hatch/idle"]
