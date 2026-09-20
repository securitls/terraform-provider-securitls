package securitls

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ resource.Resource = &deviceResource{}

type deviceResource struct {
	client *Client
}

type deviceResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Credential    types.String `tfsdk:"credential"`
	Hostname      types.String `tfsdk:"hostname"`
	Port          types.Int64  `tfsdk:"port"`
	Username      types.String `tfsdk:"username"`
	CRLType       types.String `tfsdk:"crl_type"`
	CRLPath       types.String `tfsdk:"crl_path"`
	SatelliteID   types.String `tfsdk:"satellite_id"`
	CreatedAt     types.String `tfsdk:"created_at"`
	Attachments   types.List   `tfsdk:"attachment"`
	CAAttachments types.List   `tfsdk:"ca_attachment"`
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
	FileValidation         types.Object `tfsdk:"file_validation"`
	TLSValidation          types.Object `tfsdk:"tls_validation"`
}

type deviceCAAttachmentModel struct {
	CertID               types.String `tfsdk:"cert_id"`
	Path                 types.String `tfsdk:"path"`
	CAUpdateTrustPreset  types.String `tfsdk:"ca_update_trust_preset"`
	CAUpdateTrustCommand types.String `tfsdk:"ca_update_trust_command"`
	FileValidation       types.Object `tfsdk:"file_validation"`
}

type deviceFileValidationModel struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

type deviceTLSValidationModel struct {
	Enabled    types.Bool   `tfsdk:"enabled"`
	Port       types.Int64  `tfsdk:"port"`
	ServerName types.String `tfsdk:"server_name"`
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},

		Blocks: map[string]schema.Block{
			"attachment": schema.ListNestedBlock{
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
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},

						"key_mode": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},

						"key_path": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},

						"ca_path": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},

						"include_root": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},

						"encryption_credential_id": schema.StringAttribute{
							Optional: true,
						},

						"file_validation": schema.SingleNestedAttribute{
							Optional:    true,
							Computed:    true,
							Description: "File validation settings. If omitted, enabled defaults to true.",
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									deviceFileValidationObjectType().AttrTypes,
									map[string]attr.Value{
										"enabled": types.BoolValue(true),
									},
								),
							),
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									Optional:    true,
									Computed:    true,
									Default:     booldefault.StaticBool(true),
									Description: "Whether file validation is enabled. Defaults to true.",
								},
							},
						},

						"tls_validation": schema.SingleNestedAttribute{
							Optional:    true,
							Computed:    true,
							Description: "TLS validation settings. If omitted, enabled defaults to false.",
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									deviceTLSValidationObjectType().AttrTypes,
									map[string]attr.Value{
										"enabled":     types.BoolValue(false),
										"port":        types.Int64Null(),
										"server_name": types.StringNull(),
									},
								),
							),
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									Optional:    true,
									Computed:    true,
									Default:     booldefault.StaticBool(false),
									Description: "Whether TLS validation is enabled. Defaults to false.",
								},

								"port": schema.Int64Attribute{
									Optional:    true,
									Description: "TLS validation port. Required when TLS validation is enabled.",
								},

								"server_name": schema.StringAttribute{
									Optional:    true,
									Description: "TLS server name used for certificate verification. Required when TLS validation is enabled.",
								},
							},
						},
					},
				},
			},

			"ca_attachment": schema.ListNestedBlock{
				Description: "A CA trust attachment. Only root and intermediate certificates are valid.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cert_id": schema.StringAttribute{
							Required: true,
						},

						"path": schema.StringAttribute{
							Required: true,
						},

						"ca_update_trust_preset": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("none"),
							Description: "Trust-store update preset. Defaults to none.",
						},

						"ca_update_trust_command": schema.StringAttribute{
							Optional:    true,
							Description: "Custom trust-store update command when ca_update_trust_preset is custom.",
						},

						"file_validation": schema.SingleNestedAttribute{
							Optional:    true,
							Computed:    true,
							Description: "File validation settings. If omitted, enabled defaults to true.",
							Default: objectdefault.StaticValue(
								types.ObjectValueMust(
									deviceFileValidationObjectType().AttrTypes,
									map[string]attr.Value{
										"enabled": types.BoolValue(true),
									},
								),
							),
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									Optional:    true,
									Computed:    true,
									Default:     booldefault.StaticBool(true),
									Description: "Whether file validation is enabled. Defaults to true.",
								},
							},
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

func deviceFileValidationObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"enabled": types.BoolType,
		},
	}
}

func deviceTLSValidationObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"enabled":     types.BoolType,
			"port":        types.Int64Type,
			"server_name": types.StringType,
		},
	}
}

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
			"file_validation":          deviceFileValidationObjectType(),
			"tls_validation":           deviceTLSValidationObjectType(),
		},
	}
}

func deviceCAAttachmentObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"cert_id":                 types.StringType,
			"path":                    types.StringType,
			"ca_update_trust_preset":  types.StringType,
			"ca_update_trust_command": types.StringType,
			"file_validation":         deviceFileValidationObjectType(),
		},
	}
}

func getDeviceAttachments(
	ctx context.Context,
	list types.List,
) ([]deviceAttachmentModel, diag.Diagnostics) {
	var attachments []deviceAttachmentModel

	if list.IsNull() || list.IsUnknown() {
		return attachments, nil
	}

	diags := list.ElementsAs(
		ctx,
		&attachments,
		false,
	)

	return attachments, diags
}

func getDeviceCAAttachments(
	ctx context.Context,
	list types.List,
) ([]deviceCAAttachmentModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	if list.IsNull() || list.IsUnknown() {
		return []deviceCAAttachmentModel{}, diags
	}

	var attachments []deviceCAAttachmentModel

	diags.Append(
		list.ElementsAs(
			ctx,
			&attachments,
			false,
		)...,
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
	ctx context.Context,
	attachment deviceAttachmentModel,
) (map[string]any, error) {
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

	// SecuriTLS attachment defaults:
	//   file validation = enabled
	//   TLS validation  = disabled
	//
	// Optional Terraform blocks override these defaults when present.
	fileValidation := map[string]any{
		"enabled": true,
	}

	tlsValidation := map[string]any{
		"enabled": false,
	}

	if !attachment.FileValidation.IsNull() &&
		!attachment.FileValidation.IsUnknown() {

		var file deviceFileValidationModel

		diags := attachment.FileValidation.As(
			ctx,
			&file,
			basetypes.ObjectAsOptions{},
		)

		if diags.HasError() {
			return nil, fmt.Errorf(
				"failed to decode file_validation: %v",
				diags,
			)
		}

		enabled := true
		if !file.Enabled.IsNull() && !file.Enabled.IsUnknown() {
			enabled = file.Enabled.ValueBool()
		}

		fileValidation["enabled"] = enabled
	}

	if !attachment.TLSValidation.IsNull() &&
		!attachment.TLSValidation.IsUnknown() {

		var tls deviceTLSValidationModel

		diags := attachment.TLSValidation.As(
			ctx,
			&tls,
			basetypes.ObjectAsOptions{},
		)

		if diags.HasError() {
			return nil, fmt.Errorf(
				"failed to decode tls_validation: %v",
				diags,
			)
		}

		enabled := false
		if !tls.Enabled.IsNull() && !tls.Enabled.IsUnknown() {
			enabled = tls.Enabled.ValueBool()
		}

		tlsValidation["enabled"] = enabled

		if enabled {
			if tls.Port.IsNull() ||
				tls.Port.IsUnknown() ||
				tls.Port.ValueInt64() <= 0 {
				return nil, fmt.Errorf(
					"enabled tls_validation requires a positive port",
				)
			}

			if tls.ServerName.IsNull() ||
				tls.ServerName.IsUnknown() ||
				tls.ServerName.ValueString() == "" {
				return nil, fmt.Errorf(
					"enabled tls_validation requires server_name",
				)
			}

			tlsValidation["port"] = tls.Port.ValueInt64()
			tlsValidation["serverName"] = tls.ServerName.ValueString()
		}
	}

	p["validation"] = map[string]any{
		"file": fileValidation,
		"tls":  tlsValidation,
	}

	return p, nil
}

func caAttachmentPayload(
	ctx context.Context,
	attachment deviceCAAttachmentModel,
) (map[string]any, error) {
	p := map[string]any{
		"path": attachment.Path.ValueString(),
	}

	preset := "none"
	if !attachment.CAUpdateTrustPreset.IsNull() &&
		!attachment.CAUpdateTrustPreset.IsUnknown() &&
		attachment.CAUpdateTrustPreset.ValueString() != "" {
		preset = attachment.CAUpdateTrustPreset.ValueString()
	}

	p["caUpdateTrustPreset"] = preset

	if preset == "custom" {
		if attachment.CAUpdateTrustCommand.IsNull() ||
			attachment.CAUpdateTrustCommand.IsUnknown() ||
			attachment.CAUpdateTrustCommand.ValueString() == "" {
			return nil, fmt.Errorf(
				"ca_update_trust_preset custom requires ca_update_trust_command",
			)
		}

		p["caUpdateTrustCmd"] = attachment.CAUpdateTrustCommand.ValueString()
	}

	fileEnabled := true

	if !attachment.FileValidation.IsNull() &&
		!attachment.FileValidation.IsUnknown() {

		var file deviceFileValidationModel

		diags := attachment.FileValidation.As(
			ctx,
			&file,
			basetypes.ObjectAsOptions{},
		)

		if diags.HasError() {
			return nil, fmt.Errorf(
				"failed to decode ca_attachment.file_validation: %v",
				diags,
			)
		}

		if !file.Enabled.IsNull() && !file.Enabled.IsUnknown() {
			fileEnabled = file.Enabled.ValueBool()
		}
	}

	p["validation"] = map[string]any{
		"file": map[string]any{
			"enabled": fileEnabled,
		},
	}

	return p, nil
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

	if !exists || raw == nil {
		leafList, leafDiags := types.ListValueFrom(
			ctx,
			deviceAttachmentObjectType(),
			[]deviceAttachmentModel{},
		)
		diags.Append(leafDiags...)

		caList, caDiags := types.ListValueFrom(
			ctx,
			deviceCAAttachmentObjectType(),
			[]deviceCAAttachmentModel{},
		)
		diags.Append(caDiags...)

		if !diags.HasError() {
			m.Attachments = leafList
			m.CAAttachments = caList
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

	leafAttachments := make([]deviceAttachmentModel, 0)
	caAttachments := make([]deviceCAAttachmentModel, 0)

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

		if stringFromMap(a, "type") == "ca" {
			attachment := deviceCAAttachmentModel{
				CertID:               types.StringNull(),
				Path:                 types.StringNull(),
				CAUpdateTrustPreset:  types.StringValue("none"),
				CAUpdateTrustCommand: types.StringNull(),
				FileValidation:       types.ObjectNull(deviceFileValidationObjectType().AttrTypes),
			}

			if v := stringFromMap(a, "certId", "cert_id", "certificateId"); v != "" {
				attachment.CertID = types.StringValue(v)
			}

			if v := stringFromMap(a, "path"); v != "" {
				attachment.Path = types.StringValue(v)
			}

			if v := stringFromMap(a, "caUpdateTrustPreset", "ca_update_trust_preset"); v != "" {
				attachment.CAUpdateTrustPreset = types.StringValue(v)
			}

			if v := stringFromMap(a, "caUpdateTrustCmd", "ca_update_trust_command"); v != "" {
				attachment.CAUpdateTrustCommand = types.StringValue(v)
			}

			fileEnabled := true
			if validation := mapFromMap(a, "validation"); validation != nil {
				if fileValidation := mapFromMap(validation, "file"); fileValidation != nil {
					if enabled, ok := fileValidation["enabled"].(bool); ok {
						fileEnabled = enabled
					}
				}
			}

			fileObj, fileObjDiags := types.ObjectValueFrom(
				ctx,
				deviceFileValidationObjectType().AttrTypes,
				deviceFileValidationModel{
					Enabled: types.BoolValue(fileEnabled),
				},
			)
			diags.Append(fileObjDiags...)
			if !fileObjDiags.HasError() {
				attachment.FileValidation = fileObj
			}

			if attachment.CertID.IsNull() || attachment.CertID.ValueString() == "" {
				diags.AddError(
					"Invalid SecuriTLS response",
					fmt.Sprintf(
						"CA attachment does not contain certId: %v",
						a,
					),
				)
				continue
			}

			if attachment.Path.IsNull() || attachment.Path.ValueString() == "" {
				diags.AddError(
					"Invalid SecuriTLS response",
					fmt.Sprintf(
						"CA attachment does not contain path: %v",
						a,
					),
				)
				continue
			}

			caAttachments = append(caAttachments, attachment)
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
			FileValidation:         types.ObjectNull(deviceFileValidationObjectType().AttrTypes),
			TLSValidation:          types.ObjectNull(deviceTLSValidationObjectType().AttrTypes),
		}

		if v := stringFromMap(a, "certId", "cert_id", "certificateId"); v != "" {
			attachment.CertID = types.StringValue(v)
		}

		if v := stringFromMap(a, "path"); v != "" {
			attachment.Path = types.StringValue(v)
		}

		if v := stringFromMap(a, "chainMode", "chain_mode"); v != "" {
			attachment.ChainMode = types.StringValue(v)
		}

		if v := stringFromMap(a, "keyMode", "key_mode"); v != "" {
			attachment.KeyMode = types.StringValue(v)
		}

		if v := stringFromMap(a, "keyPath", "key_path"); v != "" {
			attachment.KeyPath = types.StringValue(v)
		}

		if v := stringFromMap(a, "caPath", "ca_path"); v != "" {
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

		if v := stringFromMap(a, "encryption", "encryption_credential_id"); v != "" {
			attachment.EncryptionCredentialID = types.StringValue(v)
		}

		fileEnabled := true
		tlsEnabled := false
		tlsPort := types.Int64Null()
		tlsServerName := types.StringNull()

		if validation := mapFromMap(a, "validation"); validation != nil {
			if fileValidation := mapFromMap(validation, "file"); fileValidation != nil {
				if enabled, ok := fileValidation["enabled"].(bool); ok {
					fileEnabled = enabled
				}
			}

			if tlsValidation := mapFromMap(validation, "tls"); tlsValidation != nil {
				if enabled, ok := tlsValidation["enabled"].(bool); ok {
					tlsEnabled = enabled
				}

				if port := floatFromMap(tlsValidation, "port"); port > 0 {
					tlsPort = types.Int64Value(int64(port))
				}

				if serverName := stringFromMap(
					tlsValidation,
					"serverName",
					"server_name",
				); serverName != "" {
					tlsServerName = types.StringValue(serverName)
				}
			}
		}

		fileObj, fileObjDiags := types.ObjectValueFrom(
			ctx,
			deviceFileValidationObjectType().AttrTypes,
			deviceFileValidationModel{
				Enabled: types.BoolValue(fileEnabled),
			},
		)
		diags.Append(fileObjDiags...)
		if !fileObjDiags.HasError() {
			attachment.FileValidation = fileObj
		}

		tlsObj, tlsObjDiags := types.ObjectValueFrom(
			ctx,
			deviceTLSValidationObjectType().AttrTypes,
			deviceTLSValidationModel{
				Enabled:    types.BoolValue(tlsEnabled),
				Port:       tlsPort,
				ServerName: tlsServerName,
			},
		)
		diags.Append(tlsObjDiags...)
		if !tlsObjDiags.HasError() {
			attachment.TLSValidation = tlsObj
		}

		if attachment.CertID.IsNull() || attachment.CertID.ValueString() == "" {
			diags.AddError(
				"Invalid SecuriTLS response",
				fmt.Sprintf(
					"Device attachment does not contain certId: %v",
					a,
				),
			)
			continue
		}

		if attachment.Path.IsNull() || attachment.Path.ValueString() == "" {
			diags.AddError(
				"Invalid SecuriTLS response",
				fmt.Sprintf(
					"Device attachment does not contain path: %v",
					a,
				),
			)
			continue
		}

		leafAttachments = append(leafAttachments, attachment)
	}

	if diags.HasError() {
		return diags
	}

	leafList, leafDiags := types.ListValueFrom(
		ctx,
		deviceAttachmentObjectType(),
		leafAttachments,
	)
	diags.Append(leafDiags...)

	caList, caDiags := types.ListValueFrom(
		ctx,
		deviceCAAttachmentObjectType(),
		caAttachments,
	)
	diags.Append(caDiags...)

	if !diags.HasError() {
		m.Attachments = leafList
		m.CAAttachments = caList
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
	payload, err := attachmentPayload(
		ctx,
		attachment,
	)

	if err != nil {
		return err
	}

	_, err = r.client.Do(
		ctx,
		http.MethodPost,
		fmt.Sprintf(
			"/devices/%s/attach/%s",
			deviceID,
			attachment.CertID.ValueString(),
		),
		payload,
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
	payload, err := attachmentPayload(
		ctx,
		attachment,
	)

	if err != nil {
		return err
	}

	_, err = r.client.Do(
		ctx,
		http.MethodPatch,
		fmt.Sprintf(
			"/devices/%s/attachments/%s",
			deviceID,
			attachment.CertID.ValueString(),
		),
		payload,
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

func (r *deviceResource) attachCA(
	ctx context.Context,
	deviceID string,
	attachment deviceCAAttachmentModel,
) error {
	payload, err := caAttachmentPayload(
		ctx,
		attachment,
	)
	if err != nil {
		return err
	}

	// Uses the same attach endpoint as leaf attachments; the backend dispatches
	// based on the verified certificate level.
	_, err = r.client.Do(
		ctx,
		http.MethodPost,
		fmt.Sprintf(
			"/devices/%s/attach/%s",
			deviceID,
			attachment.CertID.ValueString(),
		),
		payload,
		nil,
	)

	return err
}

func (r *deviceResource) updateCAAttachment(
	ctx context.Context,
	deviceID string,
	attachment deviceCAAttachmentModel,
) error {
	payload, err := caAttachmentPayload(
		ctx,
		attachment,
	)
	if err != nil {
		return err
	}

	// Uses the same attachment update endpoint; the backend dispatches to the
	// CA update handler for root/intermediate certificates.
	_, err = r.client.Do(
		ctx,
		http.MethodPatch,
		fmt.Sprintf(
			"/devices/%s/attachments/%s",
			deviceID,
			attachment.CertID.ValueString(),
		),
		payload,
		nil,
	)

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

func objectValuesEqual(
	a types.Object,
	b types.Object,
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

	return a.Equal(b)
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
		stringValuesEqual(a.EncryptionCredentialID, b.EncryptionCredentialID) &&
		objectValuesEqual(a.FileValidation, b.FileValidation) &&
		objectValuesEqual(a.TLSValidation, b.TLSValidation)
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

func caAttachmentMap(
	attachments []deviceCAAttachmentModel,
) map[string]deviceCAAttachmentModel {
	result := make(
		map[string]deviceCAAttachmentModel,
		len(attachments),
	)

	for _, attachment := range attachments {
		result[attachment.CertID.ValueString()] = attachment
	}

	return result
}

func caAttachmentsEqual(
	a deviceCAAttachmentModel,
	b deviceCAAttachmentModel,
) bool {
	return stringValuesEqual(a.CertID, b.CertID) &&
		stringValuesEqual(a.Path, b.Path) &&
		stringValuesEqual(a.CAUpdateTrustPreset, b.CAUpdateTrustPreset) &&
		stringValuesEqual(a.CAUpdateTrustCommand, b.CAUpdateTrustCommand) &&
		objectValuesEqual(a.FileValidation, b.FileValidation)
}

func (r *deviceResource) reconcileCAAttachments(
	ctx context.Context,
	deviceID string,
	oldAttachments []deviceCAAttachmentModel,
	newAttachments []deviceCAAttachmentModel,
) error {
	oldMap := caAttachmentMap(oldAttachments)
	newMap := caAttachmentMap(newAttachments)

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
				"unable to detach CA certificate %s: %w",
				oldAttachment.CertID.ValueString(),
				err,
			)
		}
	}

	for certID, newAttachment := range newMap {
		oldAttachment, exists := oldMap[certID]

		if !exists {
			if err := r.attachCA(
				ctx,
				deviceID,
				newAttachment,
			); err != nil {
				return fmt.Errorf(
					"unable to attach CA certificate %s: %w",
					newAttachment.CertID.ValueString(),
					err,
				)
			}
			continue
		}

		if caAttachmentsEqual(oldAttachment, newAttachment) {
			continue
		}

		if err := r.updateCAAttachment(
			ctx,
			deviceID,
			newAttachment,
		); err != nil {
			return fmt.Errorf(
				"unable to update CA attachment %s: %w",
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

	caAttachments, diags := getDeviceCAAttachments(
		ctx,
		plan.CAAttachments,
	)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	for _, attachment := range caAttachments {
		if err := r.attachCA(
			ctx,
			plan.ID.ValueString(),
			attachment,
		); err != nil {
			resp.Diagnostics.AddError(
				"Unable to attach CA certificate",
				fmt.Sprintf(
					"Unable to attach CA certificate %s to device %s: %s",
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

	oldCAAttachments, diags := getDeviceCAAttachments(
		ctx,
		state.CAAttachments,
	)
	resp.Diagnostics.Append(diags...)

	newCAAttachments, diags := getDeviceCAAttachments(
		ctx,
		plan.CAAttachments,
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

	err = r.reconcileCAAttachments(
		ctx,
		deviceID,
		oldCAAttachments,
		newCAAttachments,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to update device CA attachments",
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
