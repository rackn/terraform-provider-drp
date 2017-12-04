#!/bin/bash

set -e

case $(uname -s) in
    Darwin)
        shasum="command shasum -a 256";;
    Linux)
        shasum="command sha256sum";;
    *)
        # Someday, support installing on Windows.  Service creation could be tricky.
        echo "No idea how to check sha256sums"
        exit 1;;
esac



go get -u github.com/digitalrebar/provision/cmds/drbundler
PATH=$PATH:$GOPATH/bin

. tools/version.sh

version="$Prepart$MajorV.$MinorV.$PatchV$Extra-$GITHASH"

for i in terraform ; do
    cd $i
    echo -n "$version" > ._Version.meta
    cd ..
    drbundler $i $i.yaml
    $shasum $i.yaml > $i.sha256
done

tmpdir="$(mktemp -d /tmp/rs-bundle-XXXXXXXX)"
cp -a bin "$tmpdir"
(
    cd "$tmpdir"
    $shasum $(find . -type f) >sha256sums
    zip -p -r terraform-provider-drp.zip *
)
cp "$tmpdir/terraform-provider-drp.zip" .
$shasum terraform-provider-drp.zip > terraform-provider-drp.sha256
rm -rf "$tmpdir"

