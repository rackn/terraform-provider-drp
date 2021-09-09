Terraform Provider for Digital Rebar v4.4+
==========================================

- Hashicorp Website: https://www.terraform.io
- RackN Website: https://rackn.com
- Digital Rebar (DRP) Community:  http://rebar.digital

NOTE: For new users, you should use the release managed binaries from https://gitlab.com/rackn/terraform-provider-drp/releases.

NOT Documentation!
------------------

This page is about building, NOT about using, the provider!  DRP Terraform Provider documentation is maintained with the project integrations documentation, please see https://provision.readthedocs.io/en/latest/doc/integrations/terraform.html

Build Requirements
------------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.13.x
-	[Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)
-	Digital Rebar terraform/[params] in system (can be imported from RackN content)


Building The Provider
---------------------

Clone repository to: `$GOPATH/src/gitlab.com/rackn/terraform-provider-drp`

```sh
$ mkdir -p $GOPATH/src/gitlab.com/rackn; cd $GOPATH/src/gitlab.com/rackn
$ git clone git@gitlab.com:rackn/terraform-provider-drp
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/gitlab.com/rackn/terraform-provider-drp
$ make build
```

Building The Provider (v0.13+)
------------------------------

v0.13+ requres use of the required_providers stanza for your your OS and architecture!  Then it will infer the cache path.  You must copy your build output to the correct cache path.


```sh
$ mkdir -p .terraform/plugins/rackn/drp/2.0/linux_amd64
$ ln -s bin/linux/amd64/terraform-provider-drp .terraform/plugins/extras.rackn.io/rackn/drp/2.0.0/linux_amd64
```

Tests
-----

At this time, no tests are available for the provider.


Requirements for the Digital Rebar Provision (DRP) provider
-----------------------------------------------------------

DRP Terraform Provider documentation is maintained with the project integrations documentation, please see https://provision.readthedocs.io/en/tip/doc/integrations/terraform.html

The DRP Terraform Provider uses the DRP v4.4+ Pooling API to allocate and release
machines from pools.

By design, the only limited state is exposed via this provider.  This prevents Terraform state from overriding or changing DRP machine information.

The Terraform Provider update interactions are limited to the allocation/release methods.

The Terraform Provider can read additional fields ("computed" valutes) when requesting inventory. In this way, users find additional characteristics; however, these are
added to the provider carefully.


To create and upload the 3rd party registery
=============================================

See registery/readme.md
