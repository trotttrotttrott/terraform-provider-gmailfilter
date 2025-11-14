package gmailfilter

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{},

		DataSourcesMap: map[string]*schema.Resource{
			"gmailfilter_filter": dataSourceGmailfilterFilter(),
			"gmailfilter_label":  dataSourceGmailfilterLabel(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"gmailfilter_filter": resourceGmailfilterFilter(),
			"gmailfilter_label":  resourceGmailfilterLabel(),
		},
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		terraformVersion := provider.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(ctx, d, provider, terraformVersion)
	}

	return provider
}

func providerConfigure(ctx context.Context, d *schema.ResourceData, p *schema.Provider, terraformVersion string) (interface{}, diag.Diagnostics) {
	config := Config{
		terraformVersion: terraformVersion,
	}

	if err := config.LoadAndValidate(ctx); err != nil {
		return nil, diag.FromErr(err)
	}

	return &config, nil
}
