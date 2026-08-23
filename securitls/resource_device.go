package securitls

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &deviceResource{}

type deviceResource struct {
	client *Client
}

type deviceResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Credential  types.String `tfsdk:"credential"`
	Hostname    types.String `tfsdk:"hostname"`
	Port        types.Int64  `tfsdk:"port"`
	Username    types.String `tfsdk:"username"`
	CRLType     types.String `tfsdk:"crl_type"`
	CRLPath     types.String `tfsdk:"crl_path"`
	SatelliteID types.String `tfsdk:"satellite_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	Attachments types.Set    `tfsdk:"attachment"`
}

type deviceAttachmentModel struct {
	CertID                 types.String `tfsdk:"cert_id"`
	Path                   types.String `tfsdk:"path"`
	ChainMode              types.String `tfsdk:"chain_mode"`
	KeyMode                types.String `tfsdk:"key_mode"`
	KeyPath                types.String `tfsdk:"key_path"`
	CAPath                 types.String `tfsdk:"ca_path"`
	IncludeRoot            types.Bool   `tfsdk:"include_root"`
	EncryptionCredentialID types.String `tfsdk:"encryption_credential_id"`
}

func NewDeviceResource() resource.Resource {
	return &deviceResource{}
}

func (r *deviceResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *deviceResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "A SecuriTLS deployment device.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},

			"name": schema.StringAttribute{
				Required: true,
			},

			"credential": schema.StringAttribute{
				Required: true,
			},

			"hostname": schema.StringAttribute{
				Required: true,
			},

			"port": schema.Int64Attribute{
				Required: true,
			},

			"username": schema.StringAttribute{
				Required: true,
			},

			"crl_type": schema.StringAttribute{
				Required:    true,
				Description: "none, bundle, dir",
			},

			"crl_path": schema.StringAttribute{
				Optional: true,
			},

			"satellite_id": schema.StringAttribute{
				Optional: true,
			},

			"created_at": schema.StringAttribute{
				Computed: true,
			},
		},

		Blocks: map[string]schema.Block{
			"attachment": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cert_id": schema.StringAttribute{
							Required: true,
						},

						"path": schema.StringAttribute{
							Required: true,
						},

						"chain_mode": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},

						"key_mode": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},

						"key_path": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},

						"ca_path": schema.StringAttribute{
							Optional: true,
							Computed: true,
						},

						"include_root": schema.BoolAttribute{
							Optional: true,
							Computed: true,
						},

						"encryption_credential_id": schema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func (r *deviceResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	configureResource(req, &r.client)
}

//
// -----------------------------------------------------------------------------
// Attachment Terraform types
// -----------------------------------------------------------------------------
//

func deviceAttachmentObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"cert_id":                  types.StringType,
			"path":                     types.StringType,
			"chain_mode":               types.StringType,
			"key_mode":                 types.StringType,
			"key_path":                 types.StringType,
			"ca_path":                  types.StringType,
			"include_root":             types.BoolType,
			"encryption_credential_id": types.StringType,
		},
	}
}

func getDeviceAttachments(
	ctx context.Context,
	set types.Set,
) ([]deviceAttachmentModel, diag.Diagnostics) {
	var attachments []deviceAttachmentModel

	if set.IsNull() || set.IsUnknown() {
		return attachments, nil
	}

	diags := set.ElementsAs(
		ctx,
		&attachments,
		false,
	)

	return attachments, diags
}

//
// -----------------------------------------------------------------------------
// Device payload
// -----------------------------------------------------------------------------
//

func devicePayload(m deviceResourceModel) map[string]any {
	p := map[string]any{
		"name":       m.Name.ValueString(),
		"credential": m.Credential.ValueString(),
		"hostname":   m.Hostname.ValueString(),
		"port":       m.Port.ValueInt64(),
		"username":   m.Username.ValueString(),
		"crlType":    m.CRLType.ValueString(),
	}

	if !m.CRLPath.IsNull() &&
		!m.CRLPath.IsUnknown() &&
		m.CRLPath.ValueString() != "" {
		p["crlPath"] = m.CRLPath.ValueString()
	}

	if !m.SatelliteID.IsNull() &&
		!m.SatelliteID.IsUnknown() &&
		m.SatelliteID.ValueString() != "" {
		p["satelliteId"] = m.SatelliteID.ValueString()
	}

	return p
}

//
// -----------------------------------------------------------------------------
// Attachment payload
// -----------------------------------------------------------------------------
//

func attachmentPayload(
	attachment deviceAttachmentModel,
) map[string]any {
	p := map[string]any{
		"path": attachment.Path.ValueString(),
	}

	if !attachment.ChainMode.IsNull() &&
		!attachment.ChainMode.IsUnknown() {
		p["chainMode"] = attachment.ChainMode.ValueString()
	}

	if !attachment.KeyMode.IsNull() &&
		!attachment.KeyMode.IsUnknown() {
		p["keyMode"] = attachment.KeyMode.ValueString()
	}

	if !attachment.KeyPath.IsNull() &&
		!attachment.KeyPath.IsUnknown() {
		p["keyPath"] = attachment.KeyPath.ValueString()
	}

	if !attachment.CAPath.IsNull() &&
		!attachment.CAPath.IsUnknown() {
		p["caPath"] = attachment.CAPath.ValueString()
	}

	if !attachment.IncludeRoot.IsNull() &&
		!attachment.IncludeRoot.IsUnknown() {
		p["includeRoot"] = attachment.IncludeRoot.ValueBool()
	}

	return p
}

//
// -----------------------------------------------------------------------------
// Apply device API response
// -----------------------------------------------------------------------------
//

func applyDeviceResponse(
	m *deviceResourceModel,
	out map[string]any,
) {
	if v := stringFromMap(out, "_id", "id"); v != "" {
		m.ID = types.StringValue(v)
	}

	if v := stringFromMap(out, "name"); v != "" {
		m.Name = types.StringValue(v)
	}

	if v := stringFromMap(out, "credential"); v != "" {
		m.Credential = types.StringValue(v)
	}

	if v := stringFromMap(out, "hostname"); v != "" {
		m.Hostname = types.StringValue(v)
	}

	if v := stringFromMap(out, "username"); v != "" {
		m.Username = types.StringValue(v)
	}

	if v := stringFromMap(out, "crlType"); v != "" {
		m.CRLType = types.StringValue(v)
	}

	if v := stringFromMap(out, "crlPath"); v != "" {
		m.CRLPath = types.StringValue(v)
	} else {
		m.CRLPath = types.StringNull()
	}

	if v := stringFromMap(out, "satelliteId"); v != "" {
		m.SatelliteID = types.StringValue(v)
	} else {
		m.SatelliteID = types.StringNull()
	}

	if v := stringFromMap(out, "createdAt"); v != "" {
		m.CreatedAt = types.StringValue(v)
	}

	if p := floatFromMap(out, "port"); p != 0 {
		m.Port = types.Int64Value(int64(p))
	}
}

//
// -----------------------------------------------------------------------------
// Convert SecuriTLS attachments to Terraform state
// -----------------------------------------------------------------------------
//

func applyDeviceAttachments(
	ctx context.Context,
	m *deviceResourceModel,
	out map[string]any,
) diag.Diagnostics {
	var diags diag.Diagnostics

	raw, exists := out["attachments"]

	//
	// Known empty attachment collection.
	//
	if !exists || raw == nil {
		set, setDiags := types.SetValueFrom(
			ctx,
			deviceAttachmentObjectType(),
			[]deviceAttachmentModel{},
		)

		diags.Append(setDiags...)

		if !diags.HasError() {
			m.Attachments = set
		}

		return diags
	}

	rawAttachments, ok := raw.([]any)
	if !ok {
		diags.AddError(
			"Invalid SecuriTLS response",
			fmt.Sprintf(
				"Expected device attachments to be an array, received %T",
				raw,
			),
		)

		return diags
	}

	attachments := make(
		[]deviceAttachmentModel,
		0,
		len(rawAttachments),
	)

	for _, rawAttachment := range rawAttachments {
		a, ok := rawAttachment.(map[string]any)
		if !ok {
			diags.AddError(
				"Invalid SecuriTLS response",
				fmt.Sprintf(
					"Expected device attachment to be an object, received %T",
					rawAttachment,
				),
			)

			continue
		}

		attachment := deviceAttachmentModel{
			CertID:                 types.StringNull(),
			Path:                   types.StringNull(),
			ChainMode:              types.StringNull(),
			KeyMode:                types.StringNull(),
			KeyPath:                types.StringNull(),
			CAPath:                 types.StringNull(),
			IncludeRoot:            types.BoolNull(),
			EncryptionCredentialID: types.StringNull(),
		}

		if v := stringFromMap(
			a,
			"certId",
			"cert_id",
			"certificateId",
		); v != "" {
			attachment.CertID = types.StringValue(v)
		}

		if v := stringFromMap(a, "path"); v != "" {
			attachment.Path = types.StringValue(v)
		}

		if v := stringFromMap(
			a,
			"chainMode",
			"chain_mode",
		); v != "" {
			attachment.ChainMode = types.StringValue(v)
		}

		if v := stringFromMap(
			a,
			"keyMode",
			"key_mode",
		); v != "" {
			attachment.KeyMode = types.StringValue(v)
		}

		if v := stringFromMap(
			a,
			"keyPath",
			"key_path",
		); v != "" {
			attachment.KeyPath = types.StringValue(v)
		}

		if v := stringFromMap(
			a,
			"caPath",
			"ca_path",
		); v != "" {
			attachment.CAPath = types.StringValue(v)
		}

		if v, exists := a["includeRoot"]; exists {
			if b, ok := v.(bool); ok {
				attachment.IncludeRoot = types.BoolValue(b)
			}
		} else if v, exists := a["include_root"]; exists {
			if b, ok := v.(bool); ok {
				attachment.IncludeRoot = types.BoolValue(b)
			}
		}

		if v := stringFromMap(
			a,
			"encryption",
			"encryption_credential_id",
		); v != "" {
			attachment.EncryptionCredentialID = types.StringValue(v)
		}

		if attachment.CertID.IsNull() ||
			attachment.CertID.ValueString() == "" {
			diags.AddError(
				"Invalid SecuriTLS response",
				fmt.Sprintf(
					"Device attachment does not contain certId: %v",
					a,
				),
			)

			continue
		}

		if attachment.Path.IsNull() ||
			attachment.Path.ValueString() == "" {
			diags.AddError(
				"Invalid SecuriTLS response",
				fmt.Sprintf(
					"Device attachment does not contain path: %v",
					a,
				),
			)

			continue
		}

		attachments = append(
			attachments,
			attachment,
		)
	}

	if diags.HasError() {
		return diags
	}

	set, setDiags := types.SetValueFrom(
		ctx,
		deviceAttachmentObjectType(),
		attachments,
	)

	diags.Append(setDiags...)

	if !diags.HasError() {
		m.Attachments = set
	}

	return diags
}

//
// -----------------------------------------------------------------------------
// Read entire device from SecuriTLS
// -----------------------------------------------------------------------------
//

func (r *deviceResource) readDevice(
	ctx context.Context,
	deviceID string,
	m *deviceResourceModel,
) (int, diag.Diagnostics, error) {
	var diags diag.Diagnostics
	var out map[string]any

	status, err := r.client.Do(
		ctx,
		http.MethodGet,
		"/devices/"+deviceID,
		nil,
		&out,
	)

	if err != nil {
		return status, diags, err
	}

	applyDeviceResponse(
		m,
		out,
	)

	diags.Append(
		applyDeviceAttachments(
			ctx,
			m,
			out,
		)...,
	)

	return status, diags, nil
}

//
// -----------------------------------------------------------------------------
// Attachment API calls
// -----------------------------------------------------------------------------
//
// POST /devices/:deviceId/attach/:certId
//

func (r *deviceResource) attachCertificate(
	ctx context.Context,
	deviceID string,
	attachment deviceAttachmentModel,
) error {
	_, err := r.client.Do(
		ctx,
		http.MethodPost,
		fmt.Sprintf(
			"/devices/%s/attach/%s",
			deviceID,
			attachment.CertID.ValueString(),
		),
		attachmentPayload(attachment),
		nil,
	)

	return err
}

//
// PATCH /devices/:deviceId/attachments/:certId
//

func (r *deviceResource) updateCertificateAttachment(
	ctx context.Context,
	deviceID string,
	attachment deviceAttachmentModel,
) error {
	_, err := r.client.Do(
		ctx,
		http.MethodPatch,
		fmt.Sprintf(
			"/devices/%s/attachments/%s",
			deviceID,
			attachment.CertID.ValueString(),
		),
		attachmentPayload(attachment),
		nil,
	)

	return err
}

//
// POST /devices/:deviceId/detach/:certId
//

func (r *deviceResource) detachCertificate(
	ctx context.Context,
	deviceID string,
	certID string,
) error {
	status, err := r.client.Do(
		ctx,
		http.MethodPost,
		fmt.Sprintf(
			"/devices/%s/detach/%s",
			deviceID,
			certID,
		),
		nil,
		nil,
	)

	//
	// Already detached = desired state satisfied.
	//
	if status == http.StatusNotFound {
		return nil
	}

	return err
}

//
// -----------------------------------------------------------------------------
// Attachment comparison
// -----------------------------------------------------------------------------
//

func attachmentKey(
	a deviceAttachmentModel,
) string {
	return a.CertID.ValueString()
}

func attachmentMap(
	attachments []deviceAttachmentModel,
) map[string]deviceAttachmentModel {
	result := make(
		map[string]deviceAttachmentModel,
		len(attachments),
	)

	for _, attachment := range attachments {
		result[attachmentKey(attachment)] = attachment
	}

	return result
}

func stringValuesEqual(
	a types.String,
	b types.String,
) bool {
	if a.IsNull() != b.IsNull() {
		return false
	}

	if a.IsUnknown() != b.IsUnknown() {
		return false
	}

	if a.IsNull() || a.IsUnknown() {
		return true
	}

	return a.ValueString() == b.ValueString()
}

func boolValuesEqual(
	a types.Bool,
	b types.Bool,
) bool {
	if a.IsNull() != b.IsNull() {
		return false
	}

	if a.IsUnknown() != b.IsUnknown() {
		return false
	}

	if a.IsNull() || a.IsUnknown() {
		return true
	}

	return a.ValueBool() == b.ValueBool()
}

func attachmentsEqual(
	a deviceAttachmentModel,
	b deviceAttachmentModel,
) bool {
	return stringValuesEqual(a.CertID, b.CertID) &&
		stringValuesEqual(a.Path, b.Path) &&
		stringValuesEqual(a.ChainMode, b.ChainMode) &&
		stringValuesEqual(a.KeyMode, b.KeyMode) &&
		stringValuesEqual(a.KeyPath, b.KeyPath) &&
		stringValuesEqual(a.CAPath, b.CAPath) &&
		boolValuesEqual(a.IncludeRoot, b.IncludeRoot) &&
		stringValuesEqual(a.EncryptionCredentialID, b.EncryptionCredentialID)
}

//
// -----------------------------------------------------------------------------
// Reconcile attachments
// -----------------------------------------------------------------------------
//
// Old only:
//     POST /devices/:deviceId/detach/:certId
//
// New only:
//     POST /devices/:deviceId/attach/:certId
//
// Exists in both but settings changed:
//     PATCH /devices/:deviceId/attachments/:certId
//

func (r *deviceResource) reconcileAttachments(
	ctx context.Context,
	deviceID string,
	oldAttachments []deviceAttachmentModel,
	newAttachments []deviceAttachmentModel,
) error {
	oldMap := attachmentMap(oldAttachments)
	newMap := attachmentMap(newAttachments)

	//
	// Remove attachments that disappeared from configuration.
	//
	for certID, oldAttachment := range oldMap {
		if _, exists := newMap[certID]; exists {
			continue
		}

		if err := r.detachCertificate(
			ctx,
			deviceID,
			oldAttachment.CertID.ValueString(),
		); err != nil {
			return fmt.Errorf(
				"unable to detach certificate %s: %w",
				oldAttachment.CertID.ValueString(),
				err,
			)
		}
	}

	//
	// Add new attachments and PATCH changed attachments.
	//
	for certID, newAttachment := range newMap {
		oldAttachment, exists := oldMap[certID]

		if !exists {
			if err := r.attachCertificate(
				ctx,
				deviceID,
				newAttachment,
			); err != nil {
				return fmt.Errorf(
					"unable to attach certificate %s: %w",
					newAttachment.CertID.ValueString(),
					err,
				)
			}

			continue
		}

		if attachmentsEqual(
			oldAttachment,
			newAttachment,
		) {
			continue
		}

		if err := r.updateCertificateAttachment(
			ctx,
			deviceID,
			newAttachment,
		); err != nil {
			return fmt.Errorf(
				"unable to update certificate attachment %s: %w",
				newAttachment.CertID.ValueString(),
				err,
			)
		}
	}

	return nil
}

//
// -----------------------------------------------------------------------------
// CREATE
// -----------------------------------------------------------------------------
//

func (r *deviceResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan deviceResourceModel

	resp.Diagnostics.Append(
		req.Plan.Get(ctx, &plan)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	//
	// Create the device first.
	//
	var out map[string]any

	_, err := r.client.Do(
		ctx,
		http.MethodPost,
		"/devices",
		devicePayload(plan),
		&out,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create device",
			err.Error(),
		)

		return
	}

	applyDeviceResponse(
		&plan,
		out,
	)

	if plan.ID.IsNull() ||
		plan.ID.IsUnknown() ||
		plan.ID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid SecuriTLS response",
			fmt.Sprintf(
				"Device response has no id: %v",
				out,
			),
		)

		return
	}

	//
	// Attach every certificate requested in configuration.
	//
	attachments, diags := getDeviceAttachments(
		ctx,
		plan.Attachments,
	)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	for _, attachment := range attachments {
		if err := r.attachCertificate(
			ctx,
			plan.ID.ValueString(),
			attachment,
		); err != nil {
			resp.Diagnostics.AddError(
				"Unable to attach certificate",
				fmt.Sprintf(
					"Unable to attach certificate %s to device %s: %s",
					attachment.CertID.ValueString(),
					plan.ID.ValueString(),
					err,
				),
			)

			return
		}
	}

	//
	// Re-read after creation so Terraform state contains the canonical
	// values returned by SecuriTLS, including attachment defaults.
	//
	status, readDiags, err := r.readDevice(
		ctx,
		plan.ID.ValueString(),
		&plan,
	)

	resp.Diagnostics.Append(readDiags...)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read device after creation",
			fmt.Sprintf(
				"SecuriTLS returned HTTP %d: %s",
				status,
				err,
			),
		)

		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(
		resp.State.Set(ctx, &plan)...,
	)
}

//
// -----------------------------------------------------------------------------
// READ
// -----------------------------------------------------------------------------
//

func (r *deviceResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state deviceResourceModel

	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	status, diags, err := r.readDevice(
		ctx,
		state.ID.ValueString(),
		&state,
	)

	resp.Diagnostics.Append(diags...)

	if status == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read device",
			err.Error(),
		)

		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(
		resp.State.Set(ctx, &state)...,
	)
}

//
// -----------------------------------------------------------------------------
// UPDATE
// -----------------------------------------------------------------------------
//

func (r *deviceResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var state deviceResourceModel
	var plan deviceResourceModel

	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	resp.Diagnostics.Append(
		req.Plan.Get(ctx, &plan)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := state.ID.ValueString()

	//
	// Update normal device properties.
	//
	_, err := r.client.Do(
		ctx,
		http.MethodPatch,
		"/devices/"+deviceID,
		devicePayload(plan),
		nil,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update device",
			err.Error(),
		)

		return
	}

	//
	// Decode existing and planned attachment sets.
	//
	oldAttachments, diags := getDeviceAttachments(
		ctx,
		state.Attachments,
	)

	resp.Diagnostics.Append(diags...)

	newAttachments, diags := getDeviceAttachments(
		ctx,
		plan.Attachments,
	)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	//
	// Reconcile using the SecuriTLS-specific endpoints.
	//
	err = r.reconcileAttachments(
		ctx,
		deviceID,
		oldAttachments,
		newAttachments,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update device attachments",
			err.Error(),
		)

		return
	}

	//
	// Re-read canonical state from SecuriTLS.
	//
	//
	// Start with existing state so immutable computed values such as ID
	// and CreatedAt are never accidentally replaced with unknown plan values.
	//
	updated := state

	status, readDiags, err := r.readDevice(
		ctx,
		deviceID,
		&updated,
	)

	resp.Diagnostics.Append(readDiags...)

	if status == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read device after update",
			err.Error(),
		)

		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(
		resp.State.Set(ctx, &updated)...,
	)
}

//
// -----------------------------------------------------------------------------
// DELETE
// -----------------------------------------------------------------------------
//

func (r *deviceResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state deviceResourceModel

	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	//
	// Device deletion is expected to remove its attachment relationships
	// server-side. There is no reason to issue individual detach calls here.
	//
	status, err := r.client.Do(
		ctx,
		http.MethodDelete,
		"/devices/"+state.ID.ValueString(),
		nil,
		nil,
	)

	if err != nil && status != http.StatusNotFound {
		resp.Diagnostics.AddError(
			"Unable to delete device",
			err.Error(),
		)
	}
}
