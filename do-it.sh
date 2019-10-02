#!/bin/bash

set -x

export GO111MODULE=on
go get github.com/kardianos/govendor
./tools/build.sh
./tools/package.sh
./tools/publish.sh
./tools/test-setup.sh
export PATH="$PWD/tools/build:$PATH"
make test
make testacc
make vendor-status
make vet
