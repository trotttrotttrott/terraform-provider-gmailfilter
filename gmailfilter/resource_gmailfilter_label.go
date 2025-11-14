package gmailfilter

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/api/gmail/v1"
)

var _ resource.Resource = &LabelResource{}
var _ resource.ResourceWithImportState = &LabelResource{}

func NewLabelResource() resource.Resource {
	return &LabelResource{}
}

type LabelResource struct {
	config *Config
}

type LabelResourceModel struct {
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

func (r *LabelResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_label"
}

func (r *LabelResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Gmail label",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The immutable ID of the label",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the label",
			},
			"background_color": schema.StringAttribute{
				Optional:    true,
				Description: "The background color represented as hex string #RRGGBB",
			},
			"text_color": schema.StringAttribute{
				Optional:    true,
				Description: "The text color of the label, represented as hex string",
			},
			"label_list_visibility": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The visibility of the label in the label list in the Gmail web interface",
			},
			"message_list_visibility": schema.StringAttribute{
				Optional:    true,
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

func (r *LabelResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(*Config)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *Config")
		return
	}
	r.config = config
}

func (r *LabelResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data LabelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	label := &gmail.Label{
		Name:                  data.Name.ValueString(),
		LabelListVisibility:   data.LabelListVisibility.ValueString(),
		MessageListVisibility: data.MessageListVisibility.ValueString(),
	}

	// Add color if both background and text colors are provided
	if !data.BackgroundColor.IsNull() && !data.TextColor.IsNull() {
		label.Color = &gmail.LabelColor{
			BackgroundColor: data.BackgroundColor.ValueString(),
			TextColor:       data.TextColor.ValueString(),
		}
	}

	result, err := r.config.gmailService.Users.Labels.Create(gmailUser, label).Do()
	if err != nil {
		resp.Diagnostics.AddError("Failed to create label", err.Error())
		return
	}

	// Update model with computed values
	data.ID = types.StringValue(result.Id)
	r.updateModelFromAPIResponse(&data, result)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LabelResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data LabelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	label, err := r.config.gmailService.Users.Labels.Get(gmailUser, data.ID.ValueString()).Do()
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read label", err.Error())
		return
	}

	r.updateModelFromAPIResponse(&data, label)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LabelResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data LabelResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	label := &gmail.Label{
		Name:                  data.Name.ValueString(),
		LabelListVisibility:   data.LabelListVisibility.ValueString(),
		MessageListVisibility: data.MessageListVisibility.ValueString(),
	}

	// Add color if both background and text colors are provided
	if !data.BackgroundColor.IsNull() && !data.TextColor.IsNull() {
		label.Color = &gmail.LabelColor{
			BackgroundColor: data.BackgroundColor.ValueString(),
			TextColor:       data.TextColor.ValueString(),
		}
	}

	result, err := r.config.gmailService.Users.Labels.Update(gmailUser, data.ID.ValueString(), label).Do()
	if err != nil {
		resp.Diagnostics.AddError("Failed to update label", err.Error())
		return
	}

	r.updateModelFromAPIResponse(&data, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *LabelResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data LabelResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.config.gmailService.Users.Labels.Delete(gmailUser, data.ID.ValueString()).Do()
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete label", err.Error())
		return
	}
}

func (r *LabelResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *LabelResource) updateModelFromAPIResponse(data *LabelResourceModel, label *gmail.Label) {
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
}
