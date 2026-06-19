FROM cr.shanhu.io/bld/base
LABEL author="Shanhu Tech Inc."

ARG NPM_VERSION

RUN <<EOF
#!/bin/sh

set -euo pipefail

apk add --update nodejs npm make git openssh zip bash
npm install -g "npm@${NPM_VERSION}"
npm install -g typescript less
[[ "$(uname -m)" != amd64 ]] || npm install -g esbuild-linux-64
[[ "$(uname -m)" != aarch64 ]] || npm install -g esbuild-linux-arm64
mkdir /usr/local/idle
EOF
COPY idle.js /usr/local/idle/idle.js

COPY ssh_known_hosts /etc/ssh/known_hosts

RUN <<EOF
#!/bin/sh

set -euo pipefail

adduser -D -u 3000 node
mkdir -p /opt/node
chown -R node /opt/node
EOF
WORKDIR /opt/node
USER node

CMD ["/usr/bin/node", "/usr/local/idle/idle.js"]
