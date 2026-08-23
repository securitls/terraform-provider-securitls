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

var _ resource.Resource = &storageProviderResource{}

type storageProviderResource struct{ client *Client }
type storageProviderResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Kind            types.String `tfsdk:"kind"`
	Label           types.String `tfsdk:"label"`
	AWSMode         types.String `tfsdk:"aws_mode"`
	AWSRegion       types.String `tfsdk:"aws_region"`
	AWSBucketName   types.String `tfsdk:"aws_bucket_name"`
	AWSAccessKeyID  types.String `tfsdk:"aws_access_key_id"`
	AWSSecretKey    types.String `tfsdk:"aws_secret_access_key"`
	AWSSessionToken types.String `tfsdk:"aws_session_token"`
	AWSRoleARN      types.String `tfsdk:"aws_role_arn"`
	AWSExternalID   types.String `tfsdk:"aws_external_id"`
	SIABucketName   types.String `tfsdk:"sia_bucket_name"`
	SIAEndpoint     types.String `tfsdk:"sia_endpoint"`
	SIAPassword     types.String `tfsdk:"sia_password"`
}

func NewStorageProviderResource() resource.Resource { return &storageProviderResource{} }
func (r *storageProviderResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_provider"
}
func (r *storageProviderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{Description: "A SecuriTLS certificate storage provider. Sensitive credentials are write-only and preserved in Terraform state.", Attributes: map[string]schema.Attribute{
		"id":       schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
		"kind":     schema.StringAttribute{Required: true, Description: "s3-self-managed, sia-self-managed", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"label":    schema.StringAttribute{Optional: true},
		"aws_mode": schema.StringAttribute{Optional: true}, "aws_region": schema.StringAttribute{Optional: true}, "aws_bucket_name": schema.StringAttribute{Optional: true}, "aws_access_key_id": schema.StringAttribute{Optional: true, Sensitive: true}, "aws_secret_access_key": schema.StringAttribute{Optional: true, Sensitive: true}, "aws_session_token": schema.StringAttribute{Optional: true, Sensitive: true}, "aws_role_arn": schema.StringAttribute{Optional: true}, "aws_external_id": schema.StringAttribute{Optional: true, Sensitive: true},
		"sia_bucket_name": schema.StringAttribute{Optional: true}, "sia_endpoint": schema.StringAttribute{Optional: true}, "sia_password": schema.StringAttribute{Optional: true, Sensitive: true},
	}}
}
func (r *storageProviderResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	configureResource(req, &r.client)
}
func setIfString(p map[string]any, key string, v types.String) {
	if !v.IsNull() && !v.IsUnknown() && v.ValueString() != "" {
		p[key] = v.ValueString()
	}
}
func storageProviderPayload(m storageProviderResourceModel) map[string]any {
	p := map[string]any{"kind": m.Kind.ValueString()}
	setIfString(p, "label", m.Label)
	if m.Kind.ValueString() == "s3-self-managed" {
		aws := map[string]any{}
		setIfString(aws, "mode", m.AWSMode)
		setIfString(aws, "region", m.AWSRegion)
		setIfString(aws, "bucketName", m.AWSBucketName)
		setIfString(aws, "accessKeyId", m.AWSAccessKeyID)
		setIfString(aws, "secretAccessKey", m.AWSSecretKey)
		setIfString(aws, "sessionToken", m.AWSSessionToken)
		setIfString(aws, "roleArn", m.AWSRoleARN)
		setIfString(aws, "externalId", m.AWSExternalID)
		p["aws"] = aws
	}
	if m.Kind.ValueString() == "sia-self-managed" {
		sia := map[string]any{}
		setIfString(sia, "bucketName", m.SIABucketName)
		setIfString(sia, "endpoint", m.SIAEndpoint)
		setIfString(sia, "password", m.SIAPassword)
		p["sia"] = sia
	}
	return p
}
func applyStorageProvider(m *storageProviderResourceModel, out map[string]any) {
	if v := stringFromMap(out, "_id", "id"); v != "" {
		m.ID = types.StringValue(v)
	}
	if v := stringFromMap(out, "kind"); v != "" {
		m.Kind = types.StringValue(v)
	}
	if v := stringFromMap(out, "label"); v != "" {
		m.Label = types.StringValue(v)
	}
	if aws := mapFromMap(out, "aws"); aws != nil {
		if v := stringFromMap(aws, "mode"); v != "" {
			m.AWSMode = types.StringValue(v)
		}
		if v := stringFromMap(aws, "region"); v != "" {
			m.AWSRegion = types.StringValue(v)
		}
		if v := stringFromMap(aws, "bucketName"); v != "" {
			m.AWSBucketName = types.StringValue(v)
		}
		if v := stringFromMap(aws, "accessKeyId"); v != "" {
			m.AWSAccessKeyID = types.StringValue(v)
		}
		if v := stringFromMap(aws, "roleArn"); v != "" {
			m.AWSRoleARN = types.StringValue(v)
		}
	}
	if sia := mapFromMap(out, "sia"); sia != nil {
		if v := stringFromMap(sia, "bucketName"); v != "" {
			m.SIABucketName = types.StringValue(v)
		}
		if v := stringFromMap(sia, "endpoint"); v != "" {
			m.SIAEndpoint = types.StringValue(v)
		}
	}
}
func (r *storageProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageProviderResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out map[string]any
	_, err := r.client.Do(ctx, http.MethodPost, "/storage/providers", storageProviderPayload(plan), &out)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create storage provider", err.Error())
		return
	}
	applyStorageProvider(&plan, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *storageProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var out map[string]any
	status, err := r.client.Do(ctx, http.MethodGet, "/storage/providers/"+state.ID.ValueString(), nil, &out)
	if status == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read storage provider", err.Error())
		return
	}
	applyStorageProvider(&state, out)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *storageProviderResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var plan storageProviderResourceModel
	var state storageProviderResourceModel

	resp.Diagnostics.Append(
		req.Plan.Get(ctx, &plan)...,
	)

	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Do(
		ctx,
		http.MethodPatch,
		"/storage/providers/"+state.ID.ValueString(),
		storageProviderPayload(plan),
		nil,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update storage provider",
			err.Error(),
		)
		return
	}

	// ID is immutable for an in-place update.
	plan.ID = state.ID

	resp.Diagnostics.Append(
		resp.State.Set(ctx, &plan)...,
	)
}
func (r *storageProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageProviderResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	status, err := r.client.Do(ctx, http.MethodDelete, "/storage/providers/"+state.ID.ValueString(), nil, nil)
	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError("Unable to delete storage provider", err.Error())
	}
}
