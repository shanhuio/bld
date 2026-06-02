#!/bin/bash

set -e

apk update && apk add --no-cache gcc openssl

export GOOS=linux
export GOARCH="$(go env GOARCH)"
export GOHOSTOS=linux
export GOHOSTARCH="$(go env GOARCH)"
export GOROOT_BOOTSTRAP=/usr/lib/go

tar -xzf /usr/local/go-src.tgz -C /usr/local

(cd /usr/local/go/src; /bin/bash make.bash)

rm -rf /usr/local/go/pkg/bootstrap /usr/local/go/pkg/obj

tar -czf /usr/local/go.tgz -C /usr/local go
