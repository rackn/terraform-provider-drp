#!/usr/bin/env bash

set -e

. tools/version.sh
version="$Prepart$MajorV.$MinorV.$PatchV$Extra"

TOKEN=R0cketSk8ts
for i in drp-tp ; do
    echo "Publishing $i to cloud"
    CONTENT=$i

    arches=("amd64")
    oses=("linux" "darwin" "windows")
    for arch in "${arches[@]}"; do
        for os in "${oses[@]}"; do
            suffix=""
            if [[ $os == windows ]] ; then
              suffix=".exe"
            fi
            path="$CONTENT/$version/"
            mkdir -p "rebar-catalog/$path"
            cp bin/$os/$arch/terraform-provider-drp${suffix} "rebar-catalog/$path/${os}_${arch}${suffix}"
        done
    done
done

