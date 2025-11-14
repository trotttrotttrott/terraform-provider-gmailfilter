package gmailfilter

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &LabelDataSource{}

func NewLabelDataSource() datasource.DataSource {
	return &LabelDataSource{}
}

type LabelDataSource struct {
	config *Config
}

type LabelDataSourceModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	BackgroundColor       types.String `tfsdk:"background_color"`
	TextColor             types.String `tfsdk:"text_color"`
	LabelListVisibility   types.String `tfsdk:"label_list_visibility"`
	MessageListVisibility types.String `tfsdk:"message_list_visibility"`
	MessagesTotal         types.Int64  `tfsdk:"messages_total"`
	MessagesUnread        types.Int64  `tfsdk:"messages_unread"`
	ThreadsTotal          types.Int64  `tfsdk:"threads_total"`
	ThreadsUnread         types.Int64  `tfsdk:"threads_unread"`
	Type                  types.String `tfsdk:"type"`
}

func (d *LabelDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label"
}

func (d *LabelDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Gmail label by name",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The immutable ID of the label",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the label",
			},
			"background_color": schema.StringAttribute{
				Computed:    true,
				Description: "The background color represented as hex string #RRGGBB",
			},
			"text_color": schema.StringAttribute{
				Computed:    true,
				Description: "The text color of the label, represented as hex string",
			},
			"label_list_visibility": schema.StringAttribute{
				Computed:    true,
				Description: "The visibility of the label in the label list in the Gmail web interface",
			},
			"message_list_visibility": schema.StringAttribute{
				Computed:    true,
				Description: "The visibility of messages with this label in the message list",
			},
			"messages_total": schema.Int64Attribute{
				Computed:    true,
				Description: "The total number of messages with the label",
			},
			"messages_unread": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of unread messages with the label",
			},
			"threads_total": schema.Int64Attribute{
				Computed:    true,
				Description: "The total number of threads with the label",
			},
			"threads_unread": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of unread threads with the label",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The owner type for the label (user or system)",
			},
		},
	}
}

func (d *LabelDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LabelDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LabelDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.config.gmailService.Users.Labels.List(gmailUser).Do()
	if err != nil {
		resp.Diagnostics.AddError("Failed to list labels", err.Error())
		return
	}

	// Find label by name
	for _, label := range res.Labels {
		if label.Name == data.Name.ValueString() {
			data.ID = types.StringValue(label.Id)
			data.Name = types.StringValue(label.Name)
			data.LabelListVisibility = types.StringValue(label.LabelListVisibility)
			data.MessageListVisibility = types.StringValue(label.MessageListVisibility)
			data.MessagesTotal = types.Int64Value(label.MessagesTotal)
			data.MessagesUnread = types.Int64Value(label.MessagesUnread)
			data.ThreadsTotal = types.Int64Value(label.ThreadsTotal)
			data.ThreadsUnread = types.Int64Value(label.ThreadsUnread)
			data.Type = types.StringValue(label.Type)

			if label.Color != nil {
				data.BackgroundColor = types.StringValue(label.Color.BackgroundColor)
				data.TextColor = types.StringValue(label.Color.TextColor)
			} else {
				data.BackgroundColor = types.StringNull()
				data.TextColor = types.StringNull()
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	resp.Diagnostics.AddError("Label not found", fmt.Sprintf("No label with name %q found", data.Name.ValueString()))
}
