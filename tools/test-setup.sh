#!/bin/bash

set -x

srcdir="$(go mod download -json github.com/digitalrebar/provision/v4 |jq -r '.Dir')"
if ! [[ $srcdir ]]; then
    echo "Failed to fetch location of client code for running unit tests"
    exit 1
fi
srcv="$(go mod download -json github.com/digitalrebar/provision/v4 |jq -r '.Version')"
tmpdir="$(mktemp -d "$HOME/.provision-server-test-XXXXXXXX")"
if ! [[ -d $tmpdir ]]; then
    echo "Failed to create temporary local dir for running client unit tests"
    exit 1
fi
(
    export GOMOD_VER="$srcv"
    cd "$tmpdir"
    cp -a "$srcdir"/* .
    find -type d -exec chmod u+w '{}' ';'
    chmod 755 tools/*.sh
    tools/build-one.sh cmds/drpcli
) || exit 1

export PATH="$PWD/bin/$(go env GOOS)/$(go env GOARCH):$PWD/tools/build:$PATH"
mkdir -p tools/build
cp "$tmpdir/bin/$(go env GOOS)/$(go env GOARCH)/drpcli" tools/build

if ! which dr-provision ; then
  if [[ $Extra && $Extra = *beta* ]]; then
      drpcli catalog item download drp --version=tip
  else
      drpcli catalog item download drp
  fi
  unzip drp.zip "bin/$(go env GOOS)/$(go env GOARCH)/dr-provision"
  rm drp.zip
  mv "bin/$(go env GOOS)/$(go env GOARCH)/dr-provision" tools/build
  chmod +x tools/build/*
fi

rm -rf $tmpdir

if ! which dr-provision ; then
    echo "No dr-provision binary to run tests against"
    exit 1
fi

