---
layout: "drp"
page_title: "Provider: Drp"
sidebar_current: "docs-drp-index"
description: |-
  The Amazon Web Services (DRP) provider is used to interact with the many resources supported by DRP. The provider needs to be configured with the proper credentials before it can be used.
---

# DRP Provider

The Amazon Web Services (DRP) provider is used to interact with the
many resources supported by DRP. The provider needs to be configured
with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Configure the DRP Provider
provider "aws" {
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  region     = "us-east-1"
}

# Create a web server
resource "aws_instance" "web" {
  # ...
}
```

