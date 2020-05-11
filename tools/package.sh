#!/bin/bash

set -e

BINARY="bin/darwin/amd64/terraform-provider-drp"
zip "${BINARY}_${TRAVIS_TAG}_darwin_amd64.zip" "${BINARY}"

BINARY="bin/linux/amd64/terraform-provider-drp"
zip "${BINARY}_${TRAVIS_TAG}_linux_amd64.zip" "${BINARY}"

##!/bin/bash
#
#set -e
#
#case $(uname -s) in
#    Darwin)
#        shasum="command shasum -a 256";;
#    Linux)
#        shasum="command sha256sum";;
#    *)
#        # Someday, support installing on Windows.  Service creation could be tricky.
#        echo "No idea how to check sha256sums"
#        exit 1;;
#esac
#
#. tools/version.sh
#
#version="$Prepart$MajorV.$MinorV.$PatchV$Extra-$GITHASH"
#
#tmpdir="$(mktemp -d /tmp/rs-bundle-XXXXXXXX)"
#cp -a bin "$tmpdir"
#(
#    cd "$tmpdir"
#    $shasum $(find . -type f) >sha256sums
#    zip -p -r terraform-provider-drp.zip *
#)
#cp "$tmpdir/terraform-provider-drp.zip" .
#$shasum terraform-provider-drp.zip > terraform-provider-drp.sha256
#rm -rf "$tmpdir"
#
