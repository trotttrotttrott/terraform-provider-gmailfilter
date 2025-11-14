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
		config := Config{}
		if err := config.LoadAndValidate(ctx); err != nil {
			return nil, diag.FromErr(err)
		}
		return &config, nil
	}

	return provider
}
