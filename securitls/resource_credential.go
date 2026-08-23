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

var _ resource.Resource = &credentialResource{}

type credentialResource struct{ client *Client }

type credentialResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Secret    types.String `tfsdk:"secret"`
	Key       types.String `tfsdk:"key"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func NewCredentialResource() resource.Resource { return &credentialResource{} }
func (r *credentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_credential"
}
func (r *credentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A SecuriTLS deployment credential. type is immutable; secret/key values are write-only and remain in Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true},
			"name":       schema.StringAttribute{Required: true},
			"type":       schema.StringAttribute{Required: true, Description: "secret or rsa", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"secret":     schema.StringAttribute{Optional: true, Sensitive: true, Description: "Secret value when type=secret."},
			"key":        schema.StringAttribute{Optional: true, Sensitive: true, Description: "Private key value when type=rsa."},
			"created_at": schema.StringAttribute{Computed: true},
		},
	}
}
func (r *credentialResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	configureResource(req, &r.client)
}
func credentialPayload(m credentialResourceModel, includeType bool) (map[string]any, error) {
	p := map[string]any{"name": m.Name.ValueString()}
	if includeType {
		p["type"] = m.Type.ValueString()
	}
	switch m.Type.ValueString() {
	case "secret":
		if includeType && (m.Secret.IsNull() || m.Secret.ValueString() == "") {
			return nil, fmt.Errorf("secret is required when type=secret")
		}
		if !m.Secret.IsNull() && !m.Secret.IsUnknown() {
			p["secret"] = m.Secret.ValueString()
		}
	case "rsa":
		if includeType && (m.Key.IsNull() || m.Key.ValueString() == "") {
			return nil, fmt.Errorf("key is required when type=rsa")
		}
		if !m.Key.IsNull() && !m.Key.IsUnknown() {
			p["key"] = m.Key.ValueString()
		}
	default:
		return nil, fmt.Errorf("type must be secret or rsa")
	}
	return p, nil
}
func (r *credentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan credentialResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	payload, err := credentialPayload(plan, true)
	if err != nil {
		resp.Diagnostics.AddError("Invalid credential configuration", err.Error())
		return
	}
	var out map[string]any
	_, err = r.client.Do(ctx, http.MethodPost, "/creds", payload, &out)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create credential", err.Error())
		return
	}
	id := stringFromMap(out, "credId", "_id", "id")
	if id == "" {
		resp.Diagnostics.AddError("Invalid SecuriTLS response", fmt.Sprintf("credential create response did not contain an id: %v", out))
		return
	}
	plan.ID = types.StringValue(id)
	if plan.CreatedAt.IsUnknown() {
		plan.CreatedAt = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *credentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state credentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out struct {
		Creds []map[string]any `json:"creds"`
	}
	_, err := r.client.Do(ctx, http.MethodGet, "/creds", nil, &out)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read credentials", err.Error())
		return
	}
	for _, c := range out.Creds {
		if stringFromMap(c, "_id", "id") == state.ID.ValueString() {
			state.Name = types.StringValue(stringFromMap(c, "name"))
			state.Type = types.StringValue(stringFromMap(c, "type"))
			if v := stringFromMap(c, "createdAt"); v != "" {
				state.CreatedAt = types.StringValue(v)
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}
	resp.State.RemoveResource(ctx)
}
func (r *credentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state credentialResourceModel
	var plan credentialResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := credentialPayload(plan, false)
	if err != nil {
		resp.Diagnostics.AddError("Invalid credential configuration", err.Error())
		return
	}

	_, err = r.client.Do(ctx, http.MethodPatch, "/creds/"+state.ID.ValueString(), payload, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update credential", err.Error())
		return
	}

	state.Name = plan.Name
	state.Secret = plan.Secret
	state.Key = plan.Key

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *credentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state credentialResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := r.client.Do(ctx, http.MethodDelete, "/creds/"+state.ID.ValueString(), nil, nil)
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Unable to delete credential", err.Error())
	}
}
