package gmailfilter

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &GmailFilterProvider{}

type GmailFilterProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GmailFilterProvider{
			version: version,
		}
	}
}

func (p *GmailFilterProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "gmailfilter"
	resp.Version = p.version
}

func (p *GmailFilterProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Gmail filters and labels using Application Default Credentials.",
	}
}

func (p *GmailFilterProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	config := &Config{}
	if err := config.LoadAndValidate(ctx); err != nil {
		resp.Diagnostics.AddError("Failed to configure provider", err.Error())
		return
	}
	resp.ResourceData = config
	resp.DataSourceData = config
}

func (p *GmailFilterProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFilterResource,
		NewLabelResource,
	}
}

func (p *GmailFilterProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewFilterDataSource,
		NewLabelDataSource,
	}
}
