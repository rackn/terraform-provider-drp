README for Hashicorp Registery
==============================

To use the terraform provider, we must create AND HOST a registery based on Hashicorp's very specific syntax and checksum requirements.  See https://www.terraform.io/docs/internals/provider-registry-protocol.html

This directory contains artificats use to populate that repository and maintain the required artifacts.

Currently, RackN is maintaining this repository under https://extras.rackn.io

Artifacts
=========

Please see the specific files referenced for formats.  This document covers the purpose, not the syntax.

.well-known/terraform.json
--------------------------

_required_

This file and path must be reachable from the TLD provided in the plan document.  It proscribes the deeper path into the registery.

Reference is `terraform.json`


rackn/drp/versions
------------------

_required_

This json file (no .json!) contains a list of the support versions of the provider.  This will be used to resolve deeper content for the provider based on the required version and architecture.

Information in this file MUST map to an arch json file at the paths inferred from the data.

Reference is `versions`

rackn/drp/[version]/[os]/[arch]
-------------------------------

For example, for version 2.0.0 on platform linux amd64, the required
path and file is `/rackn/drp/2.0.0/linux/amd64` where the json file is named `amd64`.


Reference is `arch.reference`


Build.sh
========

Build.sh is designed to create all the relevant files and update them to the s3 with public access and correct content type.  It assumes you've added credentials for aws cli since they are not coded into the file.

This includes:
* putting the binary in zip files for each architecture
* building the correct SHA256 subs and signatures for the zip file
* signing the signatures with GPG and attaching the public key (you need to setup gpg)

At the end, it invalidates the cloudfront cache.