#!/bin/bash

set -e

BINARY="bin/darwin/amd64/terraform-provider-drp"
zip "${BINARY}_${TRAVIS_TAG}_darwin_amd64.zip" "${BINARY}"

BINARY="bin/linux/amd64/terraform-provider-drp"
zip "${BINARY}_${TRAVIS_TAG}_linux_amd64.zip" "${BINARY}"
