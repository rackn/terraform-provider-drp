#!/bin/bash

set -e

VERSION=$(echo $TRAVIS_TAG | sed 's/v//g')

BINARY="bin/darwin/amd64/terraform-provider-drp"
zip "${BINARY}_${VERSION}_darwin_amd64.zip" "${BINARY}"

BINARY="bin/linux/amd64/terraform-provider-drp"
zip "${BINARY}_${VERSION}_linux_amd64.zip" "${BINARY}"
