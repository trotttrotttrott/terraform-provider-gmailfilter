package gmailfilter

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"google.golang.org/api/gmail/v1"
)

var _ resource.Resource = &FilterResource{}
var _ resource.ResourceWithImportState = &FilterResource{}
var _ resource.ResourceWithUpgradeState = &FilterResource{}

func NewFilterResource() resource.Resource {
	return &FilterResource{}
}

type FilterResource struct {
	config *Config
}

type FilterResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Action   types.Object `tfsdk:"action"`
	Criteria types.Object `tfsdk:"criteria"`
}

type FilterActionModel struct {
	AddLabelIds    types.List   `tfsdk:"add_label_ids"`
	Forward        types.String `tfsdk:"forward"`
	RemoveLabelIds types.List   `tfsdk:"remove_label_ids"`
}

type FilterCriteriaModel struct {
	ExcludeChats   types.Bool   `tfsdk:"exclude_chats"`
	From           types.String `tfsdk:"from"`
	HasAttachment  types.Bool   `tfsdk:"has_attachment"`
	NegatedQuery   types.String `tfsdk:"negated_query"`
	Query          types.String `tfsdk:"query"`
	Size           types.Int64  `tfsdk:"size"`
	SizeComparison types.String `tfsdk:"size_comparison"`
	Subject        types.String `tfsdk:"subject"`
	To             types.String `tfsdk:"to"`
}

func (r *FilterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filter"
}

func (r *FilterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     1,
		Description: "Manages a Gmail filter",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The server assigned ID of the filter",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"action": schema.SingleNestedBlock{
				Description: "Action that the filter performs. Changes to this block will require the filter to be recreated.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"add_label_ids": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "List of labels to add to the message",
					},
					"forward": schema.StringAttribute{
						Optional:    true,
						Description: "Email address that the message should be forwarded to",
					},
					"remove_label_ids": schema.ListAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Description: "List of labels to remove from the message",
					},
				},
			},
			"criteria": schema.SingleNestedBlock{
				Description: "The criteria that a message should match to apply the filter. Changes to this block will require the filter to be recreated.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"exclude_chats": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether the response should exclude chats",
					},
					"from": schema.StringAttribute{
						Optional:    true,
						Description: "The sender's display name or email address",
					},
					"has_attachment": schema.BoolAttribute{
						Optional:    true,
						Description: "Whether the message has any attachment",
					},
					"negated_query": schema.StringAttribute{
						Optional:    true,
						Description: "Only return messages not matching the specified query",
					},
					"query": schema.StringAttribute{
						Optional:    true,
						Description: "Only return messages matching the specified query",
					},
					"size": schema.Int64Attribute{
						Optional:    true,
						Description: "The size of the entire RFC822 message in bytes",
					},
					"size_comparison": schema.StringAttribute{
						Optional:    true,
						Description: "How the message size should be compared (larger/smaller/unspecified)",
					},
					"subject": schema.StringAttribute{
						Optional:    true,
						Description: "Case-insensitive phrase found in the message's subject",
					},
					"to": schema.StringAttribute{
						Optional:    true,
						Description: "The recipient's display name or email address",
					},
				},
			},
		},
	}
}

func (r *FilterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract action and criteria from plan
	var action FilterActionModel
	var criteria FilterCriteriaModel

	resp.Diagnostics.Append(data.Action.As(ctx, &action, basetypes.ObjectAsOptions{})...)
	resp.Diagnostics.Append(data.Criteria.As(ctx, &criteria, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert to Gmail API format
	filter := &gmail.Filter{
		Action:   convertActionToGmailAPI(ctx, action, &resp.Diagnostics),
		Criteria: convertCriteriaToGmailAPI(ctx, criteria, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.config.gmailService.Users.Settings.Filters.Create(gmailUser, filter).Do()
	if err != nil {
		resp.Diagnostics.AddError("Failed to create filter", err.Error())
		return
	}

	data.ID = types.StringValue(result.Id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.config.gmailService.Users.Settings.Filters.Get(gmailUser, data.ID.ValueString()).Do()
	if err != nil {
		if isNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read filter", err.Error())
		return
	}

	// Keep the existing values from state (Gmail API returns the same values)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FilterResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The RequiresReplace plan modifiers on action and criteria blocks ensure
	// that any changes to these blocks will trigger a Delete + Create cycle
	// instead of calling this Update method.
	// This method is kept for completeness and any potential future attributes
	// that might not require replacement.

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.config.gmailService.Users.Settings.Filters.Delete(gmailUser, data.ID.ValueString()).Do()
	if err != nil && !isNotFoundError(err) {
		resp.Diagnostics.AddError("Failed to delete filter", err.Error())
		return
	}
}

func (r *FilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *FilterResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &schema.Schema{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
					},
					"action": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"add_label_ids": schema.ListAttribute{
									ElementType: types.StringType,
									Optional:    true,
								},
								"forward": schema.StringAttribute{
									Optional: true,
								},
								"remove_label_ids": schema.ListAttribute{
									ElementType: types.StringType,
									Optional:    true,
								},
							},
						},
					},
					"criteria": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"exclude_chats": schema.BoolAttribute{
									Optional: true,
								},
								"from": schema.StringAttribute{
									Optional: true,
								},
								"has_attachment": schema.BoolAttribute{
									Optional: true,
								},
								"negated_query": schema.StringAttribute{
									Optional: true,
								},
								"query": schema.StringAttribute{
									Optional: true,
								},
								"size": schema.Int64Attribute{
									Optional: true,
								},
								"size_comparison": schema.StringAttribute{
									Optional: true,
								},
								"subject": schema.StringAttribute{
									Optional: true,
								},
								"to": schema.StringAttribute{
									Optional: true,
								},
							},
						},
					},
				},
			},
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				type OldFilterResourceModel struct {
					ID       types.String `tfsdk:"id"`
					Action   types.List   `tfsdk:"action"`
					Criteria types.List   `tfsdk:"criteria"`
				}

				var oldData OldFilterResourceModel
				resp.Diagnostics.Append(req.State.Get(ctx, &oldData)...)
				if resp.Diagnostics.HasError() {
					return
				}

				// Extract action and criteria from lists
				var actions []FilterActionModel
				var criterias []FilterCriteriaModel

				resp.Diagnostics.Append(oldData.Action.ElementsAs(ctx, &actions, false)...)
				resp.Diagnostics.Append(oldData.Criteria.ElementsAs(ctx, &criterias, false)...)
				if resp.Diagnostics.HasError() {
					return
				}

				// Convert first element of each list to object
				actionObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
					"add_label_ids":    types.ListType{ElemType: types.StringType},
					"forward":          types.StringType,
					"remove_label_ids": types.ListType{ElemType: types.StringType},
				}, actions[0])
				resp.Diagnostics.Append(diags...)

				criteriaObj, diags := types.ObjectValueFrom(ctx, map[string]attr.Type{
					"exclude_chats":   types.BoolType,
					"from":            types.StringType,
					"has_attachment":  types.BoolType,
					"negated_query":   types.StringType,
					"query":           types.StringType,
					"size":            types.Int64Type,
					"size_comparison": types.StringType,
					"subject":         types.StringType,
					"to":              types.StringType,
				}, criterias[0])
				resp.Diagnostics.Append(diags...)

				if resp.Diagnostics.HasError() {
					return
				}

				// Create new model with objects
				newData := FilterResourceModel{
					ID:       oldData.ID,
					Action:   actionObj,
					Criteria: criteriaObj,
				}

				resp.Diagnostics.Append(resp.State.Set(ctx, newData)...)
			},
		},
	}
}

// Helper functions
func convertActionToGmailAPI(ctx context.Context, action FilterActionModel, diags *diag.Diagnostics) *gmail.FilterAction {
	var addLabelIds []string
	var removeLabelIds []string

	if !action.AddLabelIds.IsNull() {
		diags.Append(action.AddLabelIds.ElementsAs(ctx, &addLabelIds, false)...)
	}
	if !action.RemoveLabelIds.IsNull() {
		diags.Append(action.RemoveLabelIds.ElementsAs(ctx, &removeLabelIds, false)...)
	}

	return &gmail.FilterAction{
		AddLabelIds:    addLabelIds,
		Forward:        action.Forward.ValueString(),
		RemoveLabelIds: removeLabelIds,
	}
}

func convertCriteriaToGmailAPI(ctx context.Context, criteria FilterCriteriaModel, diags *diag.Diagnostics) *gmail.FilterCriteria {
	return &gmail.FilterCriteria{
		ExcludeChats:   criteria.ExcludeChats.ValueBool(),
		From:           criteria.From.ValueString(),
		HasAttachment:  criteria.HasAttachment.ValueBool(),
		NegatedQuery:   criteria.NegatedQuery.ValueString(),
		Query:          criteria.Query.ValueString(),
		Size:           criteria.Size.ValueInt64(),
		SizeComparison: criteria.SizeComparison.ValueString(),
		Subject:        criteria.Subject.ValueString(),
		To:             criteria.To.ValueString(),
	}
}

func isNotFoundError(err error) bool {
	// Simple check - could be more sophisticated
	return err != nil && (err.Error() == "googleapi: Error 404: Not Found" ||
		fmt.Sprintf("%v", err) == "googleapi: Error 404: Not Found")
}
