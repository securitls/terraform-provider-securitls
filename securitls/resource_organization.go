package securitls

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &organizationResource{}

type organizationResource struct{ client *Client }

type organizationResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func NewOrganizationResource() resource.Resource { return &organizationResource{} }
func (r *organizationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}
func (r *organizationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A SecuriTLS organization value used in certificate distinguished names.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true},
			"name":       schema.StringAttribute{Required: true, Description: "Organization name (O=). SecuriTLS intentionally does not support renaming organizations.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"created_at": schema.StringAttribute{Computed: true},
		},
	}
}
func (r *organizationResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	configureResource(req, &r.client)
}
func (r *organizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out map[string]any
	_, err := r.client.Do(ctx, http.MethodPost, "/orgs", map[string]any{"name": plan.Name.ValueString()}, &out)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create organization", err.Error())
		return
	}
	id := stringFromMap(out, "orgId", "_id", "id")
	if id == "" {
		resp.Diagnostics.AddError("Invalid SecuriTLS response", fmt.Sprintf("organization create response did not contain an id: %v", out))
		return
	}
	plan.ID = types.StringValue(id)
	// GET is list-only, so fill computed fields by locating the new id.
	var orgs []map[string]any
	_, err = r.client.Do(ctx, http.MethodGet, "/orgs", nil, &orgs)
	if err == nil {
		for _, org := range orgs {
			if stringFromMap(org, "_id", "id") == id {
				plan.Name = types.StringValue(stringFromMap(org, "name"))
				if v := stringFromMap(org, "createdAt"); v != "" {
					plan.CreatedAt = types.StringValue(v)
				}
				break
			}
		}
	}
	if plan.CreatedAt.IsUnknown() {
		plan.CreatedAt = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *organizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var orgs []map[string]any
	_, err := r.client.Do(ctx, http.MethodGet, "/orgs", nil, &orgs)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read organizations", err.Error())
		return
	}
	for _, org := range orgs {
		if stringFromMap(org, "_id", "id") == state.ID.ValueString() {
			state.Name = types.StringValue(stringFromMap(org, "name"))
			if v := stringFromMap(org, "createdAt"); v != "" {
				state.CreatedAt = types.StringValue(v)
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *organizationResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}
func (r *organizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := r.client.Do(ctx, http.MethodDelete, "/orgs/"+state.ID.ValueString(), nil, nil)
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Unable to delete organization", err.Error())
	}
}
