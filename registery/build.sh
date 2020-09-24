#!/usr/bin/env bash
# RackN Copyright 2019
# Build Terraform Registery

export PATH=$PATH:$PWD

BASE="https://extras.rackn.io/rackn/drp/"
OS="linux darwin windows"
ARCH="amd64"
VER="2.0.0"
NAME="terraform-provider-drp"
GPGOWNER="info@rackn.com"

REF=$(cat arch.reference.json)
VERSIONS=$(cat versions.reference.json)
v=$(jq ".versions[0][\"version\"] = \"$VER\"" <<< "$VERSIONS")

echo "upload well-known to s3"
cat terraform.json | jq . > /dev/null
aws s3 cp terraform.json s3://extras.rackn.io/.well-known/terraform.json --acl public-read --content-type application/json


rm ".gitignore"
touch ".gitignore"

for os in $OS; do
	for arch in $ARCH; do
		filename="${NAME}_${VER}_${os}_${arch}"
		echo "=== $filename ==="
		rm ${filename}
		rm ${filename}.zip
		rm ${filename}_SHA256SUMS
		rm ${filename}_SHA256SUMS.sig

		echo "  building zip"
		zip ${filename}.zip ../bin/${os}/${arch}/$NAME -9 -D -j
		zip ${filename}.zip ../bin/${os}/${arch}/$NAME.exe -9 -D -j
		echo "${filename}.zip" >> .gitignore

		echo "  writing sha256sum to ${os}_${arch}_SHA256SUMS"
		echo $(sha256sum  "${filename}.zip") > "${filename}_SHA256SUMS"
		echo "${filename}_SHA256SUMS" >> .gitignore

		echo "  gpg signature of SHA to ${filename}_SHA256SUMS.sig"
		gpg --detach-sign -r info@rackn.com --output ${filename}_SHA256SUMS.sig ${filename}_SHA256SUMS
		echo "${filename}_SHA256SUMS.sig" >> .gitignore

		echo "  get gpg info"
		gpg --armor --export $GPGOWNER > key.asc
		GPGKEYID=$(gpg --list-packets key.asc | awk '/keyid:/{ print $2 }' | head -n1)

		echo "  update versions file for $os $arch"
		v=$(jq ".versions[0][\"platforms\"] |= .+ [{\"os\":\"$os\", \"arch\": \"$arch\"}]" <<< "$v")

		echo "  writing output to ${filename}"
		o=$(jq ".os = \"$os\" \
			| .arch = \"$arch\" \
			| .filename = \"${filename}.zip\" \
			| .download_url = \"${BASE}${filename}.zip\" \
			| .shasums_url = \"${BASE}${filename}_SHA256SUMS\" \
			| .shasum = \"$(cat ${filename}_SHA256SUMS | awk '{print $1}')\" \
			| .shasums_signature_url = \"${BASE}${filename}_SHA256SUMS.sig\" \
			| .signing_keys.gpg_public_keys[0][\"key_id\"] = \"${GPGKEYID}\" \
			| .signing_keys.gpg_public_keys[0][\"ascii_armor\"] = \"$(awk '{printf "%s\\n", $0}' key.asc)\" \
			" <<< "$REF")
		echo $o > "${filename}"
		echo "${filename}" >> .gitignore
		cat ${filename} | jq . > /dev/null

		echo "upload reference and zip to s3"
		aws s3 cp "${filename}" s3://extras.rackn.io/rackn/drp/${VER}/download/${os}/${arch} --acl public-read --content-type application/json
		aws s3 cp "${filename}.zip" s3://extras.rackn.io/rackn/drp/${filename}.zip --acl public-read
		aws s3 cp "${filename}_SHA256SUMS" s3://extras.rackn.io/rackn/drp/${filename}_SHA256SUMS --acl public-read
		aws s3 cp "${filename}_SHA256SUMS.sig" s3://extras.rackn.io/rackn/drp/${filename}_SHA256SUMS.sig --acl public-read

		rm key.asc
	done
done

echo "update versions and upload to s3"
echo $v > "versions"
echo "versions" >> .gitignore
cat versions | jq . > /dev/null
aws s3 cp versions s3://extras.rackn.io/rackn/drp/versions --acl public-read --content-type application/json

echo "and finally, clear the cache"
aws cloudfront create-invalidation --distribution-id E3B5UZXIFAKUY0 --paths "/rackn/drp/*" "/.well-known/terraform.json"
