#!/usr/bin/env bash

set -e

[[ $GOPATH ]] || export GOPATH="$HOME/go"
fgrep -q "$GOPATH/bin" <<< "$PATH" || export PATH="$PATH:$GOPATH/bin"

go get -u github.com/stevenroose/remarshal

. tools/version.sh
version="$Prepart$MajorV.$MinorV.$PatchV$Extra-$GITHASH"

TOKEN=R0cketSk8ts
for i in terraform ; do
    echo "Publishing $i to cloud"
    CONTENT=$i
    remarshal -i $CONTENT.yaml -o $CONTENT.json -if yaml -of json
    curl -X PUT -T $CONTENT.json https://qww9e4paf1.execute-api.us-west-2.amazonaws.com/main/support/content/$CONTENT?token=$TOKEN
    echo

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

