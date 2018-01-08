#!/usr/bin/env bash

set -e

. tools/version.sh
version="$Prepart$MajorV.$MinorV.$PatchV$Extra-$GITHASH"

TOKEN=R0cketSk8ts
for i in terraform ; do
    echo "Publishing $i to cloud"
    CONTENT=$i

    arches=("amd64")
    oses=("linux" "darwin" "windows")
    for arch in "${arches[@]}"; do
        for os in "${oses[@]}"; do
            path="$CONTENT/$version/$arch/$os"
            mkdir -p "rebar-catalog/$path"
            cp bin/$os/$arch/terraform-provider-drp "rebar-catalog/$path"
        done
    done
done

