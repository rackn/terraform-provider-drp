#!/usr/bin/env bash

set -e

if ! which go &>/dev/null; then
        echo "Must have go installed"
        exit 255
fi

# Work out the GO version we are working with:
GO_VERSION=$(go version | awk '{ print $3 }' | sed 's/go//')
WANTED_VER=(1 12)
if ! [[ "$GO_VERSION" =~ ([0-9]+)\.([0-9]+) ]]; then
    echo "Cannot figure out what version of Go is installed"
    exit 1
elif ! (( ${BASH_REMATCH[1]} > ${WANTED_VER[0]} || ${BASH_REMATCH[2]} >= ${WANTED_VER[1]} )); then
    echo "Go Version needs to be ${WANTED_VER[0]}.${WANTED_VER[1]} or higher: currently $GO_VERSION"
    exit -1
fi

export GO111MODULE=on

. tools/version.sh

echo "Version = $Prepart$MajorV.$MinorV.$PatchV$Extra-$GITHASH"

VERFLAGS="-X main.RS_MAJOR_VERSION=$MajorV \
          -X main.RS_MINOR_VERSION=$MinorV \
          -X main.RS_PATCH_VERSION=$PatchV \
          -X main.RS_EXTRA=$Extra \
          -X main.RS_PREPART=$Prepart \
          -X main.BuildStamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` \
          -X main.GitHash=$GITHASH"

arches=("amd64")
oses=("linux" "darwin" "windows")
for arch in "${arches[@]}"; do
    for os in "${oses[@]}"; do
        (
            suffix=""
            if [[ $os == windows ]] ; then
              suffix=".exe"
            fi
            export GOOS="$os" GOARCH="$arch"
            echo "Building binaries for ${arch} ${os}"
            binpath="bin/$os/$arch"
            mkdir -p "$binpath"
            go build -ldflags "$VERFLAGS" -o "$binpath/terraform-provider-drp${suffix}"
        )
        done
done
echo "To run tests, run: tools/test.sh"
