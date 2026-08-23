package securitls

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &storageKeyResource{}

type storageKeyResource struct{ client *Client }
type storageKeyResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Type         types.String `tfsdk:"type"`
	Name         types.String `tfsdk:"name"`
	RoleARN      types.String `tfsdk:"role_arn"`
	KeyARN       types.String `tfsdk:"key_arn"`
	ExternalID   types.String `tfsdk:"external_id"`
	VaultURL     types.String `tfsdk:"vault_url"`
	KeyName      types.String `tfsdk:"key_name"`
	KeyVersion   types.String `tfsdk:"key_version"`
	TenantID     types.String `tfsdk:"tenant_id"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

func NewStorageKeyResource() resource.Resource { return &storageKeyResource{} }
func (r *storageKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_key"
}
func (r *storageKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	resp.Schema = schema.Schema{
		Description: "A SecuriTLS BYOK storage encryption key using AWS KMS or Azure Key Vault. SecuriTLS does not currently expose an update endpoint, so all inputs are replacement-only.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"type":          schema.StringAttribute{Required: true, Description: "awskms or azure", PlanModifiers: replace},
			"name":          schema.StringAttribute{Required: true},
			"role_arn":      schema.StringAttribute{Optional: true, PlanModifiers: replace},
			"key_arn":       schema.StringAttribute{Optional: true, PlanModifiers: replace},
			"external_id":   schema.StringAttribute{Optional: true, Sensitive: true},
			"vault_url":     schema.StringAttribute{Optional: true, PlanModifiers: replace},
			"key_name":      schema.StringAttribute{Optional: true, PlanModifiers: replace},
			"key_version":   schema.StringAttribute{Optional: true, PlanModifiers: replace},
			"tenant_id":     schema.StringAttribute{Optional: true, PlanModifiers: replace},
			"client_id":     schema.StringAttribute{Optional: true, PlanModifiers: replace},
			"client_secret": schema.StringAttribute{Optional: true, Sensitive: true},
			"created_at":    schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		},
	}
}
func (r *storageKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	configureResource(req, &r.client)
}
func storageKeyPayload(m storageKeyResourceModel) map[string]any {
	p := map[string]any{"type": m.Type.ValueString(), "name": m.Name.ValueString()}
	setIfString(p, "roleArn", m.RoleARN)
	setIfString(p, "keyArn", m.KeyARN)
	setIfString(p, "externalId", m.ExternalID)
	setIfString(p, "vaultUrl", m.VaultURL)
	setIfString(p, "keyName", m.KeyName)
	setIfString(p, "keyVersion", m.KeyVersion)
	setIfString(p, "tenantId", m.TenantID)
	setIfString(p, "clientId", m.ClientID)
	setIfString(p, "clientSecret", m.ClientSecret)
	return p
}
func applyStorageKey(m *storageKeyResourceModel, out map[string]any) {
	if v := stringFromMap(out, "_id", "id"); v != "" {
		m.ID = types.StringValue(v)
	}
	if v := stringFromMap(out, "type"); v != "" {
		m.Type = types.StringValue(v)
	}
	if v := stringFromMap(out, "name"); v != "" {
		m.Name = types.StringValue(v)
	}
	if v := stringFromMap(out, "roleArn"); v != "" {
		m.RoleARN = types.StringValue(v)
	}
	if v := stringFromMap(out, "keyArn"); v != "" {
		m.KeyARN = types.StringValue(v)
	}
	if v := stringFromMap(out, "vaultUrl"); v != "" {
		m.VaultURL = types.StringValue(v)
	}
	if v := stringFromMap(out, "keyName"); v != "" {
		m.KeyName = types.StringValue(v)
	}
	if v := stringFromMap(out, "keyVersion"); v != "" {
		m.KeyVersion = types.StringValue(v)
	}
	if v := stringFromMap(out, "tenantId"); v != "" {
		m.TenantID = types.StringValue(v)
	}
	if v := stringFromMap(out, "clientId"); v != "" {
		m.ClientID = types.StringValue(v)
	}
	if v := stringFromMap(out, "createdAt"); v != "" {
		m.CreatedAt = types.StringValue(v)
	}
}
func (r *storageKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out map[string]any
	_, err := r.client.Do(ctx, http.MethodPost, "/storage/keys", storageKeyPayload(plan), &out)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create storage key", err.Error())
		return
	}
	applyStorageKey(&plan, out)
	if plan.CreatedAt.IsUnknown() {
		plan.CreatedAt = types.StringNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *storageKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out map[string]any
	status, err := r.client.Do(ctx, http.MethodGet, "/storage/keys/"+state.ID.ValueString(), nil, &out)
	if status == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read storage key", err.Error())
		return
	}
	applyStorageKey(&state, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *storageKeyResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var state storageKeyResourceModel
	var plan storageKeyResourceModel

	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	resp.Diagnostics.Append(
		req.Plan.Get(ctx, &plan)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{}

	//
	// Name
	//
	if !plan.Name.Equal(state.Name) {
		payload["name"] = plan.Name.ValueString()
	}

	//
	// AWS external ID
	//
	if !plan.ExternalID.Equal(state.ExternalID) {
		if plan.ExternalID.IsNull() || plan.ExternalID.IsUnknown() {
			payload["externalId"] = ""
		} else {
			payload["externalId"] = plan.ExternalID.ValueString()
		}
	}

	//
	// Azure client secret
	//
	if !plan.ClientSecret.Equal(state.ClientSecret) {
		if plan.ClientSecret.IsNull() || plan.ClientSecret.IsUnknown() {
			payload["clientSecret"] = ""
		} else {
			payload["clientSecret"] = plan.ClientSecret.ValueString()
		}
	}

	_, err := r.client.Do(
		ctx,
		http.MethodPatch,
		"/storage/keys/"+state.ID.ValueString(),
		payload,
		nil,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update storage key",
			err.Error(),
		)
		return
	}

	// Only name is mutable.
	state.Name = plan.Name
	state.ExternalID = plan.ExternalID
	state.ClientSecret = plan.ClientSecret

	resp.Diagnostics.Append(
		resp.State.Set(ctx, &state)...,
	)
}
func (r *storageKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := r.client.Do(ctx, http.MethodDelete, "/storage/keys/"+state.ID.ValueString(), nil, nil)
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Unable to delete storage key", err.Error())
	}
}
