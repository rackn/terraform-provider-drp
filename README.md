Terraform Provider for Digital Rebar v4.4+
==========================================

- Hashicorp Website: https://www.terraform.io
- RackN Website: https://rackn.com
- Digital Rebar (DRP) Community:  http://rebar.digital

NOTE: For new users, you should use the release managed binaries from https://github.com/rackn/terraform-provider-drp/releases.

NOT Documentation!
------------------

This page is about building, NOT about using, the provider!  DRP Terraform Provider documentation is maintained with the project integrations documentation, please see https://provision.readthedocs.io/en/latest/doc/content-packages/terraform.html

Build Requirements
------------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.12.x
-	[Go](https://golang.org/doc/install) 1.13 (to build the provider plugin)
-	Digital Rebar terraform/[params] in system (can be imported from RackN content)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/rackn/terraform-provider-drp`

```sh
$ mkdir -p $GOPATH/src/github.com/rackn; cd $GOPATH/src/github.com/rackn
$ git clone git@github.com:rackn/terraform-provider-drp
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/rackn/terraform-provider-drp
$ make build
```

Running The Provider (v0.13+)
---------------------

v0.13+ requres use of the required_providers stanza
Update for your OS and architecture!

```sh
$ mkdir -p .terraform/plugins/rackn/drp/2.0/linux_amd64
$ ln -s bin/linux/amd64/terraform-provider-drp .terraform/plugins/extras.rackn.io/rackn/drp/2.0.0/linux_amd64
```

Requirements for the Digital Rebar Provision (DRP) provider
-----------------------------------------------------------

DRP Terraform Provider documentation is maintained with the project integrations documentation, please see https://provision.readthedocs.io/en/tip/doc/integrations/terraform.html

The DRP Terraform Provider uses the DRP v4.4+ Pooling API to allocate and release
machines from pools.

By design, the only limited state is exposed via this provider.  This prevents Terraform state from overriding or changing DRP machine information.

The Terraform Provider update interactions are limited to the allocation/release methods.

The Terraform Provider can read additional fields ("computed" valutes) when requesting inventory. In this way, users find additional characteristics; however, these are
added to the provider carefully.

Developing the Provider
-----------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.9+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make bin
...
$ $GOPATH/bin/terraform-provider-drp
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.
*Note:* In this case, acceptances run locally without external resources.

```sh
$ ulimit -n 2560 # for MACs
$ make testacc
```


To create and upload the 3rd party registery
=============================================

See registery/readme.md
