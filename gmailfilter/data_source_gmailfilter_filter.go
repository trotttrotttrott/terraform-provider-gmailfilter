package gmailfilter

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &FilterDataSource{}

func NewFilterDataSource() datasource.DataSource {
	return &FilterDataSource{}
}

type FilterDataSource struct {
	config *Config
}

type FilterDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Action   types.Object `tfsdk:"action"`
	Criteria types.Object `tfsdk:"criteria"`
}

func (d *FilterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filter"
}

func (d *FilterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Gmail filter",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the filter",
			},
			"action": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Action that the filter performs",
				Attributes: map[string]schema.Attribute{
					"add_label_ids": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "List of labels to add to the message",
					},
					"forward": schema.StringAttribute{
						Computed:    true,
						Description: "Email address that the message should be forwarded to",
					},
					"remove_label_ids": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "List of labels to remove from the message",
					},
				},
			},
			"criteria": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The criteria that a message should match to apply the filter",
				Attributes: map[string]schema.Attribute{
					"exclude_chats": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the response should exclude chats",
					},
					"from": schema.StringAttribute{
						Computed:    true,
						Description: "The sender's display name or email address",
					},
					"has_attachment": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the message has any attachment",
					},
					"negated_query": schema.StringAttribute{
						Computed:    true,
						Description: "Only return messages not matching the specified query",
					},
					"query": schema.StringAttribute{
						Computed:    true,
						Description: "Only return messages matching the specified query",
					},
					"size": schema.Int64Attribute{
						Computed:    true,
						Description: "The size of the entire RFC822 message in bytes",
					},
					"size_comparison": schema.StringAttribute{
						Computed:    true,
						Description: "How the message size should be compared",
					},
					"subject": schema.StringAttribute{
						Computed:    true,
						Description: "Case-insensitive phrase found in the message's subject",
					},
					"to": schema.StringAttribute{
						Computed:    true,
						Description: "The recipient's display name or email address",
					},
				},
			},
		},
	}
}

func (d *FilterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(*Config)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", "Expected *Config")
		return
	}
	d.config = config
}

func (d *FilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FilterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter, err := d.config.gmailService.Users.Settings.Filters.Get(gmailUser, data.ID.ValueString()).Do()
	if err != nil {
		resp.Diagnostics.AddError("Failed to read filter", err.Error())
		return
	}

	// Convert action to object
	actionObject, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"add_label_ids":    types.ListType{ElemType: types.StringType},
		"forward":          types.StringType,
		"remove_label_ids": types.ListType{ElemType: types.StringType},
	}, FilterActionModel{
		AddLabelIds:    convertStringSliceToList(ctx, filter.Action.AddLabelIds),
		Forward:        types.StringValue(filter.Action.Forward),
		RemoveLabelIds: convertStringSliceToList(ctx, filter.Action.RemoveLabelIds),
	})
	resp.Diagnostics.Append(diags...)

	// Convert criteria to object
	criteriaObject, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
		"exclude_chats":   types.BoolType,
		"from":            types.StringType,
		"has_attachment":  types.BoolType,
		"negated_query":   types.StringType,
		"query":           types.StringType,
		"size":            types.Int64Type,
		"size_comparison": types.StringType,
		"subject":         types.StringType,
		"to":              types.StringType,
	}, FilterCriteriaModel{
		ExcludeChats:   types.BoolValue(filter.Criteria.ExcludeChats),
		From:           types.StringValue(filter.Criteria.From),
		HasAttachment:  types.BoolValue(filter.Criteria.HasAttachment),
		NegatedQuery:   types.StringValue(filter.Criteria.NegatedQuery),
		Query:          types.StringValue(filter.Criteria.Query),
		Size:           types.Int64Value(filter.Criteria.Size),
		SizeComparison: types.StringValue(filter.Criteria.SizeComparison),
		Subject:        types.StringValue(filter.Criteria.Subject),
		To:             types.StringValue(filter.Criteria.To),
	})
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Action = actionObject
	data.Criteria = criteriaObject

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func convertStringSliceToList(ctx context.Context, slice []string) types.List {
	if slice == nil {
		return types.ListNull(types.StringType)
	}
	list, _ := types.ListValueFrom(ctx, types.StringType, slice)
	return list
}
