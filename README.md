# Terraform Provider for Gmail Filter

## Installation

This fork is not deployed to the terraform registry. For now you'd have to clone
this repo, build the provider locally (`go build`), and reference the provider
locally.

You can use a `.terraformrc` file like this:

```hcl
plugin_cache_dir   = "$HOME/.terraform.d/plugin-cache"
disable_checkpoint = true

provider_installation {
  dev_overrides {
    "gmailfilter" = "/path/to/terraform-provider-gmailfilter"
  }
}
```

See [docs](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers)
for more info on .terraformrc.

With the above in a `.terraformrc`, you can reference the provider like this:

```hcl
terraform {
  required_providers {
    gmailfilter = {
      source = "gmailfilter"
    }
  }
}
```

## Importing

```
terraform import gmailfilter_filter.name <filter-id>
terraform import gmailfilter_label.name <label-id>
```
