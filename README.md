Terraform Provider
==================

- Hashicorp Website: https://www.terraform.io
- RackN Website: https://rackn.com
- Digital Rebar Community:  http://rebar.digital

NOTE: For new users, you should use the release managed binaries from https://github.com/rackn/terraform-provider-drp/releases.

NOT Documentation!
------------------

This page is about building, NOT about using, the provider!  DRP Terraform Provider documentation is maintained with the project integrations documentation, please see https://provision.readthedocs.io/en/tip/doc/integrations/terraform.html

Build Requirements
------------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.11.x
-	[Go](https://golang.org/doc/install) 1.10 (to build the provider plugin)
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

Requirements for the Digital Rebar Provision (DRP) provider
-----------------------------------------------------------

DRP Terraform Provider documentation is maintained with the project integrations documentation, please see https://provision.readthedocs.io/en/tip/doc/integrations/terraform.html

The DRP Terraform Provider uses a pair of Machine Parameters to create an inventory pool. Only machines with these parameters will be available to the provider.

The terraform/managed parameter determines the basic inventory availability. This flag must be set to true for Terraform to find machines.

The terraform/allocated parameter determines when machines have been assigned to a Terraform plan. When true, the machine is now being managed by Terraform. When false, the machine is available for allocation.

The terraform/pool parameter allows operators to create groups of machines that can be managed separately.  It should have a default value of `default`.

Using the RackN terraform-ready stage will automatically set these three parameters.

The Terraform Provider can read additional fields when requesting inventory. In this way, users can request machines with specific characteristics.

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
