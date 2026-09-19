package securitls

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &certificateResource{}
var _ resource.ResourceWithModifyPlan = &certificateResource{}
var _ resource.ResourceWithImportState = &certificateResource{}

type certificateResource struct {
	client *Client
}

type certificateResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Type               types.String `tfsdk:"type"`
	CommonName         types.String `tfsdk:"common_name"`
	ExpireDays         types.Int64  `tfsdk:"expire_interval_days"`
	Signer             types.String `tfsdk:"signer"`
	SatelliteID        types.String `tfsdk:"satellite_id"`
	Storage            types.String `tfsdk:"storage"`
	Country            types.String `tfsdk:"country"`
	Organization       types.String `tfsdk:"organization"`
	OrganizationalUnit types.String `tfsdk:"organizational_unit"`
	State              types.String `tfsdk:"state"`
	Locality           types.String `tfsdk:"locality"`
	Serial             types.String `tfsdk:"serial"`
	KeyUsage           types.Set    `tfsdk:"key_usage"`
	ExtendedKeyUsage   types.Set    `tfsdk:"extended_key_usage"`
	DNSSANs            types.Set    `tfsdk:"dns_sans"`
	KeyAlgorithm       types.String `tfsdk:"key_algorithm"`
	KeySizeBits        types.Int64  `tfsdk:"key_size_bits"`
	Curve              types.String `tfsdk:"curve"`
	ParameterSet       types.String `tfsdk:"parameter_set"`
	SignatureAlgorithm types.String `tfsdk:"signature_algorithm"`
	Status             types.String `tfsdk:"status"`
	NotBefore          types.String `tfsdk:"not_before"`
	NotAfter           types.String `tfsdk:"not_after"`
	CreatedAt          types.String `tfsdk:"created_at"`
	PEM                types.String `tfsdk:"pem"`
	AutoRenew          types.Bool   `tfsdk:"auto_renew"`
	AutoDeploy         types.Bool   `tfsdk:"auto_deploy"`
	CustodyModel       types.String `tfsdk:"custody_model"`
	KeyName            types.String `tfsdk:"key_name"`
	KeyRef             types.String `tfsdk:"key_ref"`

	RenewTrigger   types.String `tfsdk:"renew_trigger"`
	RekeyTrigger   types.String `tfsdk:"rekey_trigger"`
	ReissueTrigger types.String `tfsdk:"reissue_trigger"`
}

func NewCertificateResource() resource.Resource {
	return &certificateResource{}
}

func (r *certificateResource) Metadata(
	_ context.Context,
	req resource.MetadataRequest,
	resp *resource.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (r *certificateResource) ImportState(
	ctx context.Context,
	req resource.ImportStateRequest,
	resp *resource.ImportStateResponse,
) {
	resource.ImportStatePassthroughID(
		ctx,
		path.Root("id"),
		req,
		resp,
	)
}

func replaceString() []planmodifier.String {
	return []planmodifier.String{
		stringplanmodifier.RequiresReplace(),
	}
}

func (r *certificateResource) Schema(
	_ context.Context,
	_ resource.SchemaRequest,
	resp *resource.SchemaResponse,
) {
	resp.Schema = schema.Schema{
		Description: "A SecuriTLS root, intermediate, or leaf certificate. Certificate configuration changes cause SecuriTLS to reissue the certificate and advance this Terraform resource to the successor certificate.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},

			"type": schema.StringAttribute{
				Required:    true,
				Description: "root, intermediate, or leaf",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"common_name": schema.StringAttribute{
				Required: true,
			},

			"expire_interval_days": schema.Int64Attribute{
				Required: true,
			},

			"signer": schema.StringAttribute{
				Optional: true,
			},

			"satellite_id": schema.StringAttribute{
				Optional: true,
			},

			"storage": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},

			"country": schema.StringAttribute{
				Optional: true,
			},

			"organization": schema.StringAttribute{
				Optional: true,
			},

			"organizational_unit": schema.StringAttribute{
				Optional: true,
			},

			"state": schema.StringAttribute{
				Optional: true,
			},

			"locality": schema.StringAttribute{
				Optional: true,
			},

			"serial": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},

			"key_usage": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},

			"extended_key_usage": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},

			"dns_sans": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},

			"key_algorithm": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},

			"key_size_bits": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},

			"curve": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},

			"parameter_set": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},

			"signature_algorithm": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},

			"status": schema.StringAttribute{
				Computed: true,
			},

			"not_before": schema.StringAttribute{
				Computed: true,
			},

			"not_after": schema.StringAttribute{
				Computed: true,
			},

			"created_at": schema.StringAttribute{
				Computed: true,
			},

			"pem": schema.StringAttribute{
				Computed:    true,
				Description: "X.509 certificate in PEM format.",
			},

			"auto_renew": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Automatically renew the certificate according to SecuriTLS automation policy.",
			},

			"auto_deploy": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Automatically deploy the certificate after automated renewal.",
			},

			"custody_model": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Private key custody model: securitls, satellite, or hsm. Changing this value rekeys the certificate.",
			},

			"key_name": schema.StringAttribute{
				Computed:    true,
				Description: "Stored private key name for SecuriTLS-managed or Satellite-encrypted custody.",
			},

			"key_ref": schema.StringAttribute{
				Computed:    true,
				Description: "Non-exportable private key reference for HSM-backed custody.",
			},

			"renew_trigger": schema.StringAttribute{
				Optional:    true,
				Description: "Changing this value to a new non-null value triggers certificate renewal.",
			},

			"rekey_trigger": schema.StringAttribute{
				Optional:    true,
				Description: "Changing this value to a new non-null value triggers certificate rekeying.",
			},

			"reissue_trigger": schema.StringAttribute{
				Optional:    true,
				Description: "Changing this value to a new non-null value explicitly triggers certificate reissuance.",
			},
		},
	}
}

func (r *certificateResource) Configure(
	_ context.Context,
	req resource.ConfigureRequest,
	_ *resource.ConfigureResponse,
) {
	configureResource(req, &r.client)
}

func setStringsFromSet(
	ctx context.Context,
	s types.Set,
) ([]string, error) {
	if s.IsNull() || s.IsUnknown() {
		return nil, nil
	}

	var vals []string

	diags := s.ElementsAs(
		ctx,
		&vals,
		false,
	)

	if diags.HasError() {
		return nil, fmt.Errorf("failed to decode set")
	}

	return vals, nil
}

func certificatePayload(
	ctx context.Context,
	m certificateResourceModel,
) (map[string]any, error) {
	ku, err := setStringsFromSet(
		ctx,
		m.KeyUsage,
	)
	if err != nil {
		return nil, err
	}

	eku, err := setStringsFromSet(
		ctx,
		m.ExtendedKeyUsage,
	)
	if err != nil {
		return nil, err
	}

	sans, err := setStringsFromSet(
		ctx,
		m.DNSSANs,
	)
	if err != nil {
		return nil, err
	}

	p := map[string]any{
		"type":               m.Type.ValueString(),
		"commonName":         m.CommonName.ValueString(),
		"expireIntervalDays": m.ExpireDays.ValueInt64(),
	}

	if len(ku) > 0 {
		p["keyUsageList"] = ku
	}

	if len(eku) > 0 {
		p["extendedKeyUsageList"] = eku
	}

	if len(sans) > 0 {
		arr := make(
			[]map[string]string,
			0,
			len(sans),
		)

		for _, v := range sans {
			arr = append(
				arr,
				map[string]string{
					"type":  "dns",
					"value": v,
				},
			)
		}

		p["SANS"] = arr
	}

	setIfString(
		p,
		"signer",
		m.Signer,
	)

	setIfString(
		p,
		"satelliteId",
		m.SatelliteID,
	)

	setIfString(
		p,
		"custodyModel",
		m.CustodyModel,
	)

	if !m.CustodyModel.IsNull() && !m.CustodyModel.IsUnknown() {
		model := m.CustodyModel.ValueString()
		if model != "securitls" && model != "satellite" && model != "hsm" {
			return nil, fmt.Errorf("custody_model must be one of securitls, satellite, or hsm")
		}

		if (model == "satellite" || model == "hsm") &&
			(m.SatelliteID.IsNull() || m.SatelliteID.IsUnknown() || m.SatelliteID.ValueString() == "") {
			return nil, fmt.Errorf("satellite_id is required when custody_model is %s", model)
		}
	}

	setIfString(
		p,
		"storage",
		m.Storage,
	)

	setIfString(
		p,
		"countryName",
		m.Country,
	)

	setIfString(
		p,
		"organizationName",
		m.Organization,
	)

	setIfString(
		p,
		"organizationalUnitName",
		m.OrganizationalUnit,
	)

	setIfString(
		p,
		"state",
		m.State,
	)

	setIfString(
		p,
		"locality",
		m.Locality,
	)

	setIfString(
		p,
		"serial",
		m.Serial,
	)

	// The API applies its own automation defaults when the automation object is
	// omitted. Terraform may configure either automation field independently.
	automation := map[string]any{}

	if !m.AutoRenew.IsNull() && !m.AutoRenew.IsUnknown() {
		automation["autoRenew"] = m.AutoRenew.ValueBool()
	}

	if !m.AutoDeploy.IsNull() && !m.AutoDeploy.IsUnknown() {
		automation["autoDeploy"] = m.AutoDeploy.ValueBool()
	}

	if len(automation) > 0 {
		p["automation"] = automation
	}

	ks := map[string]any{}

	setIfString(
		ks,
		"algorithm",
		m.KeyAlgorithm,
	)

	if !m.KeySizeBits.IsNull() &&
		!m.KeySizeBits.IsUnknown() &&
		m.KeySizeBits.ValueInt64() > 0 {

		ks["keySizeBits"] = m.KeySizeBits.ValueInt64()
	}

	setIfString(
		ks,
		"curve",
		m.Curve,
	)

	setIfString(
		ks,
		"parameterSet",
		m.ParameterSet,
	)

	setIfString(
		ks,
		"signatureAlgorithm",
		m.SignatureAlgorithm,
	)

	if len(ks) > 0 {
		p["keySettings"] = ks
	}

	return p, nil
}

func stringsToSet(vals []string) types.Set {
	av := make(
		[]attr.Value,
		0,
		len(vals),
	)

	for _, v := range vals {
		av = append(
			av,
			types.StringValue(v),
		)
	}

	s, _ := types.SetValue(
		types.StringType,
		av,
	)

	return s
}

func applyCertificate(
	m *certificateResourceModel,
	out map[string]any,
) {
	if v := stringFromMap(
		out,
		"_id",
		"id",
	); v != "" {
		m.ID = types.StringValue(v)
	}

	if v := stringFromMap(
		out,
		"level",
	); v != "" {
		m.Type = types.StringValue(v)
	}

	if v := stringFromMap(
		out,
		"storage",
	); v != "" {
		m.Storage = types.StringValue(v)
	}

	if v := stringFromMap(
		out,
		"status",
	); v != "" {
		m.Status = types.StringValue(v)
	}

	if v := stringFromMap(
		out,
		"createdAt",
	); v != "" {
		m.CreatedAt = types.StringValue(v)
	}

	if automation := mapFromMap(out, "automation"); automation != nil {
		if v, ok := automation["autoRenew"].(bool); ok {
			m.AutoRenew = types.BoolValue(v)
		}

		if v, ok := automation["autoDeploy"].(bool); ok {
			m.AutoDeploy = types.BoolValue(v)
		}
	}

	if key := mapFromMap(out, "key"); key != nil {
		if v := stringFromMap(key, "custodyModel"); v != "" {
			m.CustodyModel = types.StringValue(v)
		}

		if v := stringFromMap(key, "keyName"); v != "" {
			m.KeyName = types.StringValue(v)
		} else {
			m.KeyName = types.StringNull()
		}

		if v := stringFromMap(key, "keyRef"); v != "" {
			m.KeyRef = types.StringValue(v)
		} else {
			m.KeyRef = types.StringNull()
		}
	}

	d := mapFromMap(
		out,
		"details",
	)

	if d == nil {
		return
	}

	if vals, ok := d["keyUsageList"].([]any); ok {
		items := make(
			[]string,
			0,
			len(vals),
		)

		for _, v := range vals {
			if s, ok := v.(string); ok {
				items = append(
					items,
					s,
				)
			}
		}

		m.KeyUsage = stringsToSet(items)
	}

	if vals, ok := d["extendedKeyUsageList"].([]any); ok {
		items := make(
			[]string,
			0,
			len(vals),
		)

		for _, v := range vals {
			if s, ok := v.(string); ok {
				items = append(
					items,
					s,
				)
			}
		}

		m.ExtendedKeyUsage = stringsToSet(items)
	}

	if vals, ok := d["SANS"].([]any); ok {
		items := make(
			[]string,
			0,
			len(vals),
		)

		for _, v := range vals {
			san, ok := v.(map[string]any)
			if !ok {
				continue
			}

			if stringFromMap(
				san,
				"type",
			) != "dns" {
				continue
			}

			if value := stringFromMap(
				san,
				"value",
			); value != "" {
				items = append(
					items,
					value,
				)
			}
		}

		m.DNSSANs = stringsToSet(items)
	}

	if v := stringFromMap(
		d,
		"commonName",
	); v != "" {
		m.CommonName = types.StringValue(v)
	}

	if v := stringFromMap(
		d,
		"signer",
	); v != "" {
		m.Signer = types.StringValue(v)
	}

	if v := stringFromMap(
		d,
		"satelliteId",
	); v != "" {
		m.SatelliteID = types.StringValue(v)
	}

	if v := stringFromMap(
		d,
		"countryName",
	); v != "" {
		m.Country = types.StringValue(v)
	}

	if v := stringFromMap(
		d,
		"organizationName",
	); v != "" {
		m.Organization = types.StringValue(v)
	}

	if v := stringFromMap(
		d,
		"organizationalUnitName",
	); v != "" {
		m.OrganizationalUnit = types.StringValue(v)
	}

	if v := stringFromMap(
		d,
		"state",
	); v != "" {
		m.State = types.StringValue(v)
	}

	if v := stringFromMap(
		d,
		"locality",
	); v != "" {
		m.Locality = types.StringValue(v)
	}

	if v := stringFromMap(
		d,
		"serial",
	); v != "" {
		m.Serial = types.StringValue(v)
	}

	if v := stringFromMap(
		d,
		"notBefore",
	); v != "" {
		m.NotBefore = types.StringValue(v)
	}

	if v := stringFromMap(
		d,
		"notAfter",
	); v != "" {
		m.NotAfter = types.StringValue(v)
	}

	if x := floatFromMap(
		d,
		"expireIntervalDays",
	); x > 0 {
		m.ExpireDays = types.Int64Value(
			int64(x),
		)
	}

	ks := mapFromMap(
		d,
		"keySettings",
	)

	if ks != nil {
		if v := stringFromMap(
			ks,
			"algorithm",
		); v != "" {
			m.KeyAlgorithm = types.StringValue(v)
		}

		if x := floatFromMap(
			ks,
			"keySizeBits",
		); x > 0 {
			m.KeySizeBits = types.Int64Value(
				int64(x),
			)
		}

		if v := stringFromMap(
			ks,
			"curve",
		); v != "" {
			m.Curve = types.StringValue(v)
		}

		if v := stringFromMap(
			ks,
			"parameterSet",
		); v != "" {
			m.ParameterSet = types.StringValue(v)
		}

		if v := stringFromMap(
			ks,
			"signatureAlgorithm",
		); v != "" {
			m.SignatureAlgorithm = types.StringValue(v)
		}
	}
}

func normalizeCertificateComputed(
	m *certificateResourceModel,
) {
	if m.Storage.IsUnknown() {
		m.Storage = types.StringNull()
	}

	if m.Serial.IsUnknown() {
		m.Serial = types.StringNull()
	}

	if m.KeyUsage.IsUnknown() {
		m.KeyUsage = types.SetNull(
			types.StringType,
		)
	}

	if m.ExtendedKeyUsage.IsUnknown() {
		m.ExtendedKeyUsage = types.SetNull(
			types.StringType,
		)
	}

	if m.DNSSANs.IsUnknown() {
		m.DNSSANs = types.SetNull(
			types.StringType,
		)
	}

	if m.KeyAlgorithm.IsUnknown() {
		m.KeyAlgorithm = types.StringNull()
	}

	if m.KeySizeBits.IsUnknown() {
		m.KeySizeBits = types.Int64Null()
	}

	if m.Curve.IsUnknown() {
		m.Curve = types.StringNull()
	}

	if m.ParameterSet.IsUnknown() {
		m.ParameterSet = types.StringNull()
	}

	if m.SignatureAlgorithm.IsUnknown() {
		m.SignatureAlgorithm = types.StringNull()
	}

	if m.Status.IsUnknown() {
		m.Status = types.StringNull()
	}

	if m.NotBefore.IsUnknown() {
		m.NotBefore = types.StringNull()
	}

	if m.NotAfter.IsUnknown() {
		m.NotAfter = types.StringNull()
	}

	if m.CreatedAt.IsUnknown() {
		m.CreatedAt = types.StringNull()
	}

	if m.PEM.IsUnknown() {
		m.PEM = types.StringNull()
	}

	if m.AutoRenew.IsUnknown() {
		m.AutoRenew = types.BoolNull()
	}

	if m.AutoDeploy.IsUnknown() {
		m.AutoDeploy = types.BoolNull()
	}

	if m.CustodyModel.IsUnknown() {
		m.CustodyModel = types.StringNull()
	}

	if m.KeyName.IsUnknown() {
		m.KeyName = types.StringNull()
	}

	if m.KeyRef.IsUnknown() {
		m.KeyRef = types.StringNull()
	}

	if m.RenewTrigger.IsUnknown() {
		m.RenewTrigger = types.StringNull()
	}

	if m.RekeyTrigger.IsUnknown() {
		m.RekeyTrigger = types.StringNull()
	}

	if m.ReissueTrigger.IsUnknown() {
		m.ReissueTrigger = types.StringNull()
	}
}

//
// -----------------------------------------------------------------------------
// Lifecycle trigger helpers
// -----------------------------------------------------------------------------
//

//
// Trigger semantics:
//
// null  -> "1"    = execute
// "1"   -> "2"    = execute
// "abc" -> "def"  = execute
// "1"   -> "1"    = no operation
// "1"   -> null   = no remote operation
//
// Removing a trigger from configuration is therefore a Terraform-state-only
// cleanup. To invoke the operation again, set the trigger to any new value.
//

func certificateTriggerChanged(
	oldValue types.String,
	newValue types.String,
) bool {
	if newValue.IsNull() || newValue.IsUnknown() {
		return false
	}

	if oldValue.IsNull() || oldValue.IsUnknown() {
		return true
	}

	return !oldValue.Equal(newValue)
}

//
// -----------------------------------------------------------------------------
// Certificate change detection
// -----------------------------------------------------------------------------
//
// Optional + Computed values can legitimately be UNKNOWN in the plan when
// the user did not configure them. UNKNOWN therefore does NOT by itself
// constitute a certificate change.
//

func certificateStringChanged(
	state types.String,
	plan types.String,
) bool {
	if plan.IsUnknown() {
		return false
	}

	return !state.Equal(plan)
}

func certificateInt64Changed(
	state types.Int64,
	plan types.Int64,
) bool {
	if plan.IsUnknown() {
		return false
	}

	return !state.Equal(plan)
}

func certificateBoolChanged(
	state types.Bool,
	plan types.Bool,
) bool {
	if plan.IsUnknown() {
		return false
	}

	return !state.Equal(plan)
}

func certificateSetChanged(
	state types.Set,
	plan types.Set,
) bool {
	if plan.IsUnknown() {
		return false
	}

	return !state.Equal(plan)
}

func certificateConfigChanged(
	state certificateResourceModel,
	plan certificateResourceModel,
) bool {
	//
	// Required values.
	//

	if certificateStringChanged(
		state.CommonName,
		plan.CommonName,
	) {
		return true
	}

	if certificateInt64Changed(
		state.ExpireDays,
		plan.ExpireDays,
	) {
		return true
	}

	//
	// Optional values.
	//

	if certificateStringChanged(
		state.Signer,
		plan.Signer,
	) {
		return true
	}

	if certificateStringChanged(
		state.Storage,
		plan.Storage,
	) {
		return true
	}

	if certificateStringChanged(
		state.Country,
		plan.Country,
	) {
		return true
	}

	if certificateStringChanged(
		state.Organization,
		plan.Organization,
	) {
		return true
	}

	if certificateStringChanged(
		state.OrganizationalUnit,
		plan.OrganizationalUnit,
	) {
		return true
	}

	if certificateStringChanged(
		state.State,
		plan.State,
	) {
		return true
	}

	if certificateStringChanged(
		state.Locality,
		plan.Locality,
	) {
		return true
	}

	if certificateStringChanged(
		state.Serial,
		plan.Serial,
	) {
		return true
	}

	if certificateSetChanged(
		state.KeyUsage,
		plan.KeyUsage,
	) {
		return true
	}

	if certificateSetChanged(
		state.ExtendedKeyUsage,
		plan.ExtendedKeyUsage,
	) {
		return true
	}

	if certificateSetChanged(
		state.DNSSANs,
		plan.DNSSANs,
	) {
		return true
	}

	if certificateStringChanged(
		state.KeyAlgorithm,
		plan.KeyAlgorithm,
	) {
		return true
	}

	if certificateInt64Changed(
		state.KeySizeBits,
		plan.KeySizeBits,
	) {
		return true
	}

	if certificateStringChanged(
		state.Curve,
		plan.Curve,
	) {
		return true
	}

	if certificateStringChanged(
		state.ParameterSet,
		plan.ParameterSet,
	) {
		return true
	}

	if certificateStringChanged(
		state.SignatureAlgorithm,
		plan.SignatureAlgorithm,
	) {
		return true
	}

	return false
}

func certificateCustodyChanged(
	state certificateResourceModel,
	plan certificateResourceModel,
) bool {
	return certificateStringChanged(state.CustodyModel, plan.CustodyModel) ||
		certificateStringChanged(state.SatelliteID, plan.SatelliteID)
}

func certificateAutomationChanged(
	state certificateResourceModel,
	plan certificateResourceModel,
) bool {
	return certificateBoolChanged(state.AutoRenew, plan.AutoRenew) ||
		certificateBoolChanged(state.AutoDeploy, plan.AutoDeploy)
}

//
// -----------------------------------------------------------------------------
// Execute renew/rekey/reissue
// -----------------------------------------------------------------------------
//
// Endpoints:
//
// POST /certificates/:certId/renew
// POST /certificates/:certId/rekey
// POST /certificates/:certId/reissue
//
// These operations create a successor certificate.
//
// The response is expected to contain the successor certificate object.
// applyCertificate() will therefore replace m.ID with the successor ID.
//

func (r *certificateResource) executeCertificateLifecycle(
	ctx context.Context,
	currentID string,
	operation string,
	plan *certificateResourceModel,
) error {
	payload, err := certificatePayload(
		ctx,
		*plan,
	)
	if err != nil {
		return fmt.Errorf(
			"failed to build certificate payload: %w",
			err,
		)
	}

	var endpoint string

	switch operation {
	case "renew":
		endpoint = fmt.Sprintf(
			"/certificates/%s/renew",
			currentID,
		)

	case "rekey":
		endpoint = fmt.Sprintf(
			"/certificates/%s/rekey",
			currentID,
		)

	case "reissue":
		endpoint = fmt.Sprintf(
			"/certificates/%s/reissue",
			currentID,
		)

	default:
		return fmt.Errorf(
			"unsupported certificate lifecycle operation %q",
			operation,
		)
	}

	var out map[string]any

	status, err := r.client.Do(
		ctx,
		http.MethodPost,
		endpoint,
		payload,
		&out,
	)

	if err != nil {
		return fmt.Errorf(
			"SecuriTLS returned HTTP %d: %w",
			status,
			err,
		)
	}

	//
	// The lifecycle endpoint should return the new/successor certificate.
	//
	applyCertificate(
		plan,
		out,
	)

	if plan.ID.IsNull() ||
		plan.ID.IsUnknown() ||
		plan.ID.ValueString() == "" {

		return fmt.Errorf(
			"%s response did not contain a successor certificate id: %v",
			operation,
			out,
		)
	}

	if plan.ID.ValueString() == currentID {
		return fmt.Errorf(
			"%s returned the existing certificate id %s instead of a successor certificate id",
			operation,
			currentID,
		)
	}

	//
	// Pull the PEM for the newly-created successor.
	//
	r.readPEM(
		ctx,
		plan,
	)

	normalizeCertificateComputed(
		plan,
	)

	return nil
}

//
// -----------------------------------------------------------------------------
// PATCH mutable certificate settings
// -----------------------------------------------------------------------------
//

func (r *certificateResource) patchCertificate(
	ctx context.Context,
	currentID string,
	state certificateResourceModel,
	plan *certificateResourceModel,
	patchAutomation bool,
	patchCustody bool,
) error {
	mergeCertificatePlanWithState(state, plan)

	payload := map[string]any{}

	if patchAutomation {
		if certificateBoolChanged(state.AutoRenew, plan.AutoRenew) &&
			!plan.AutoRenew.IsNull() && !plan.AutoRenew.IsUnknown() {
			payload["autoRenew"] = plan.AutoRenew.ValueBool()
		}

		if certificateBoolChanged(state.AutoDeploy, plan.AutoDeploy) &&
			!plan.AutoDeploy.IsNull() && !plan.AutoDeploy.IsUnknown() {
			payload["autoDeploy"] = plan.AutoDeploy.ValueBool()
		}
	}

	if patchCustody {
		if plan.CustodyModel.IsNull() || plan.CustodyModel.IsUnknown() || plan.CustodyModel.ValueString() == "" {
			return fmt.Errorf("custody_model must be known before changing certificate custody")
		}

		model := plan.CustodyModel.ValueString()
		if model != "securitls" && model != "satellite" && model != "hsm" {
			return fmt.Errorf("custody_model must be one of securitls, satellite, or hsm")
		}

		satellite := ""
		if !plan.SatelliteID.IsNull() && !plan.SatelliteID.IsUnknown() {
			satellite = plan.SatelliteID.ValueString()
		}

		if (model == "satellite" || model == "hsm") && satellite == "" {
			return fmt.Errorf("satellite_id is required when custody_model is %s", model)
		}

		if model == "securitls" {
			satellite = ""
		}

		payload["custodyModel"] = map[string]any{
			"model":     model,
			"satellite": satellite,
		}
	}

	var out map[string]any
	status, err := r.client.Do(
		ctx,
		http.MethodPatch,
		"/certificates/"+currentID,
		payload,
		&out,
	)
	if err != nil {
		return fmt.Errorf("SecuriTLS returned HTTP %d: %w", status, err)
	}

	applyCertificate(plan, out)

	if patchCustody {
		if plan.ID.IsNull() || plan.ID.IsUnknown() || plan.ID.ValueString() == "" {
			return fmt.Errorf("custody update response did not contain a successor certificate id: %v", out)
		}

		if plan.ID.ValueString() == currentID {
			return fmt.Errorf("custody update returned the existing certificate id %s instead of a rekeyed successor certificate id", currentID)
		}
	} else {
		// Automation updates keep the current certificate/key intact.
		plan.ID = types.StringValue(currentID)
	}

	r.readPEM(ctx, plan)
	normalizeCertificateComputed(plan)
	return nil
}

//
// -----------------------------------------------------------------------------
// CREATE
// -----------------------------------------------------------------------------
//

func (r *certificateResource) Create(
	ctx context.Context,
	req resource.CreateRequest,
	resp *resource.CreateResponse,
) {
	var plan certificateResourceModel

	resp.Diagnostics.Append(
		req.Plan.Get(
			ctx,
			&plan,
		)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	p, err := certificatePayload(
		ctx,
		plan,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid certificate configuration",
			err.Error(),
		)

		return
	}

	var out map[string]any

	_, err = r.client.Do(
		ctx,
		http.MethodPost,
		"/certificates",
		p,
		&out,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create certificate",
			err.Error(),
		)

		return
	}

	applyCertificate(
		&plan,
		out,
	)

	if plan.ID.IsNull() ||
		plan.ID.IsUnknown() ||
		plan.ID.ValueString() == "" {

		resp.Diagnostics.AddError(
			"Invalid SecuriTLS response",
			fmt.Sprintf(
				"certificate response has no id: %v",
				out,
			),
		)

		return
	}

	r.readPEM(
		ctx,
		&plan,
	)

	normalizeCertificateComputed(
		&plan,
	)

	resp.Diagnostics.Append(
		resp.State.Set(
			ctx,
			&plan,
		)...,
	)
}

//
// -----------------------------------------------------------------------------
// PEM
// -----------------------------------------------------------------------------
//

func (r *certificateResource) readPEM(
	ctx context.Context,
	m *certificateResourceModel,
) {
	var p map[string]any

	if _, err := r.client.Do(
		ctx,
		http.MethodGet,
		"/certificates/"+m.ID.ValueString()+"/pem",
		nil,
		&p,
	); err == nil {

		if v := stringFromMap(
			p,
			"x509",
		); v != "" {
			m.PEM = types.StringValue(v)
		}
	}
}

//
// -----------------------------------------------------------------------------
// READ
// -----------------------------------------------------------------------------
//

func (r *certificateResource) Read(
	ctx context.Context,
	req resource.ReadRequest,
	resp *resource.ReadResponse,
) {
	var state certificateResourceModel

	resp.Diagnostics.Append(
		req.State.Get(
			ctx,
			&state,
		)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	var out map[string]any

	status, err := r.client.Do(
		ctx,
		http.MethodGet,
		"/certificates/"+state.ID.ValueString(),
		nil,
		&out,
	)

	if status == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read certificate",
			err.Error(),
		)

		return
	}

	//
	// Preserve lifecycle triggers from Terraform state.
	// applyCertificate() does not modify them.
	//

	applyCertificate(
		&state,
		out,
	)

	r.readPEM(
		ctx,
		&state,
	)

	normalizeCertificateComputed(
		&state,
	)

	resp.Diagnostics.Append(
		resp.State.Set(
			ctx,
			&state,
		)...,
	)
}

//
// -----------------------------------------------------------------------------
// UPDATE
// -----------------------------------------------------------------------------
//

func (r *certificateResource) Update(
	ctx context.Context,
	req resource.UpdateRequest,
	resp *resource.UpdateResponse,
) {
	var state certificateResourceModel
	var plan certificateResourceModel

	resp.Diagnostics.Append(
		req.State.Get(
			ctx,
			&state,
		)...,
	)

	resp.Diagnostics.Append(
		req.Plan.Get(
			ctx,
			&plan,
		)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	currentID := state.ID.ValueString()

	if currentID == "" {
		resp.Diagnostics.AddError(
			"Invalid certificate state",
			"Certificate state does not contain an id.",
		)
		return
	}

	renewTriggered := certificateTriggerChanged(state.RenewTrigger, plan.RenewTrigger)
	rekeyTriggered := certificateTriggerChanged(state.RekeyTrigger, plan.RekeyTrigger)
	reissueTriggered := certificateTriggerChanged(state.ReissueTrigger, plan.ReissueTrigger)

	triggerCount := 0
	if renewTriggered {
		triggerCount++
	}
	if rekeyTriggered {
		triggerCount++
	}
	if reissueTriggered {
		triggerCount++
	}

	if triggerCount > 1 {
		resp.Diagnostics.AddError(
			"Conflicting certificate lifecycle operations",
			"Only one of renew_trigger, rekey_trigger, or reissue_trigger may change to a new value during a single apply.",
		)
		return
	}

	configChanged := certificateConfigChanged(state, plan)
	custodyChanged := certificateCustodyChanged(state, plan)
	automationChanged := certificateAutomationChanged(state, plan)

	// Custody changes use PATCH /certificates/:certId and internally rekey.
	// Do not combine them with another lifecycle operation or a reissue-causing
	// certificate configuration change, which would otherwise create two
	// successor certificates in a single apply.
	if custodyChanged && (renewTriggered || rekeyTriggered || reissueTriggered || configChanged) {
		resp.Diagnostics.AddError(
			"Cannot change custody with another certificate lifecycle operation",
			"Changing custody already rekeys the certificate. Apply the custody change separately from certificate configuration changes and lifecycle triggers.",
		)
		return
	}

	if configChanged && renewTriggered {
		resp.Diagnostics.AddError(
			"Cannot renew and change certificate configuration simultaneously",
			"Changing certificate properties causes a reissue. Apply the configuration change separately from renew_trigger.",
		)
		return
	}

	if configChanged && rekeyTriggered {
		resp.Diagnostics.AddError(
			"Cannot rekey and change certificate configuration simultaneously",
			"Changing certificate properties causes a reissue. Apply the configuration change separately from rekey_trigger.",
		)
		return
	}

	// Custody and automation are both mutable through PATCH. A custody change
	// returns the rekeyed successor certificate; an automation-only change keeps
	// the current certificate and key intact.
	if custodyChanged || (automationChanged && !renewTriggered && !rekeyTriggered && !reissueTriggered && !configChanged) {
		err := r.patchCertificate(
			ctx,
			currentID,
			state,
			&plan,
			automationChanged,
			custodyChanged,
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update certificate",
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(
			resp.State.Set(ctx, &plan)...,
		)
		return
	}

	var operation string
	switch {
	case renewTriggered:
		operation = "renew"
	case rekeyTriggered:
		operation = "rekey"
	case reissueTriggered:
		operation = "reissue"
	case configChanged:
		operation = "reissue"
	default:
		// No remote certificate operation. This includes removing a trigger
		// from configuration.
		state.RenewTrigger = plan.RenewTrigger
		state.RekeyTrigger = plan.RekeyTrigger
		state.ReissueTrigger = plan.ReissueTrigger

		resp.Diagnostics.Append(
			resp.State.Set(ctx, &state)...,
		)
		return
	}

	mergeCertificatePlanWithState(state, &plan)

	err := r.executeCertificateLifecycle(
		ctx,
		currentID,
		operation,
		&plan,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to %s certificate", operation),
			err.Error(),
		)
		return
	}

	// Automation is mutable independently of certificate/key lifecycle. If it
	// changed in the same apply, update the newly-created successor in place.
	if automationChanged {
		successorID := plan.ID.ValueString()
		err = r.patchCertificate(
			ctx,
			successorID,
			plan,
			&plan,
			true,
			false,
		)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to update certificate automation",
				err.Error(),
			)
			return
		}
	}

	normalizeCertificateComputed(&plan)
	resp.Diagnostics.Append(
		resp.State.Set(ctx, &plan)...,
	)
}

//
// -----------------------------------------------------------------------------
// DELETE
// -----------------------------------------------------------------------------
//

func (r *certificateResource) Delete(
	ctx context.Context,
	req resource.DeleteRequest,
	resp *resource.DeleteResponse,
) {
	var state certificateResourceModel

	resp.Diagnostics.Append(
		req.State.Get(
			ctx,
			&state,
		)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.Do(
		ctx,
		http.MethodDelete,
		"/certificates/"+state.ID.ValueString(),
		nil,
		nil,
	)

	if err != nil &&
		status != http.StatusNotFound {

		resp.Diagnostics.AddError(
			"Unable to delete certificate",
			err.Error(),
		)
	}
}

func (r *certificateResource) ModifyPlan(
	ctx context.Context,
	req resource.ModifyPlanRequest,
	resp *resource.ModifyPlanResponse,
) {
	//
	// Create.
	//
	if req.State.Raw.IsNull() {
		return
	}

	//
	// Delete.
	//
	if req.Plan.Raw.IsNull() {
		return
	}

	var state certificateResourceModel
	var plan certificateResourceModel
	var config certificateResourceModel

	resp.Diagnostics.Append(
		req.State.Get(
			ctx,
			&state,
		)...,
	)

	resp.Diagnostics.Append(
		req.Plan.Get(
			ctx,
			&plan,
		)...,
	)

	resp.Diagnostics.Append(
		req.Config.Get(
			ctx,
			&config,
		)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	renewTriggered := certificateTriggerChanged(
		state.RenewTrigger,
		plan.RenewTrigger,
	)

	rekeyTriggered := certificateTriggerChanged(
		state.RekeyTrigger,
		plan.RekeyTrigger,
	)

	reissueTriggered := certificateTriggerChanged(
		state.ReissueTrigger,
		plan.ReissueTrigger,
	)

	//
	// If Terraform configuration itself contains unresolved values,
	// we cannot yet know whether the certificate will need reissuing.
	//
	// Example:
	//
	// signer = securitls_certificate.root.id
	//
	// while root.id is actually changing.
	//
	// Do not make computed values known here, because Terraform may call
	// ModifyPlan again during apply after that dependency resolves.
	//
	if certificateConfigHasUnknowns(config) {
		return
	}

	configChanged := certificateConfigChanged(
		state,
		plan,
	)
	custodyChanged := certificateCustodyChanged(state, plan)

	//
	// A genuine certificate lifecycle operation will produce a successor
	// certificate. Leave computed values unknown.
	//
	if renewTriggered ||
		rekeyTriggered ||
		reissueTriggered ||
		configChanged ||
		custodyChanged {

		return
	}

	//
	// No SecuriTLS certificate operation will happen.
	//
	// Typical case:
	//
	//     reissue_trigger = "2" -> null
	//
	// This is purely a Terraform state cleanup; no SecuriTLS API call occurs.
	// Preserve the existing certificate data in the plan.
	//

	plan.ID = state.ID
	plan.Storage = state.Storage
	plan.Serial = state.Serial
	plan.KeyUsage = state.KeyUsage
	plan.ExtendedKeyUsage = state.ExtendedKeyUsage
	plan.DNSSANs = state.DNSSANs
	plan.KeyAlgorithm = state.KeyAlgorithm
	plan.KeySizeBits = state.KeySizeBits
	plan.Curve = state.Curve
	plan.ParameterSet = state.ParameterSet
	plan.SignatureAlgorithm = state.SignatureAlgorithm
	plan.Status = state.Status
	plan.NotBefore = state.NotBefore
	plan.NotAfter = state.NotAfter
	plan.CreatedAt = state.CreatedAt
	plan.PEM = state.PEM
	plan.KeyName = state.KeyName
	plan.KeyRef = state.KeyRef

	if plan.CustodyModel.IsUnknown() {
		plan.CustodyModel = state.CustodyModel
	}

	if plan.AutoRenew.IsUnknown() {
		plan.AutoRenew = state.AutoRenew
	}

	if plan.AutoDeploy.IsUnknown() {
		plan.AutoDeploy = state.AutoDeploy
	}

	resp.Diagnostics.Append(
		resp.Plan.Set(
			ctx,
			&plan,
		)...,
	)
}

func mergeCertificatePlanWithState(
	state certificateResourceModel,
	plan *certificateResourceModel,
) {
	if plan.Storage.IsUnknown() {
		plan.Storage = state.Storage
	}

	if plan.KeyUsage.IsUnknown() {
		plan.KeyUsage = state.KeyUsage
	}

	if plan.ExtendedKeyUsage.IsUnknown() {
		plan.ExtendedKeyUsage = state.ExtendedKeyUsage
	}

	if plan.DNSSANs.IsUnknown() {
		plan.DNSSANs = state.DNSSANs
	}

	if plan.KeyAlgorithm.IsUnknown() {
		plan.KeyAlgorithm = state.KeyAlgorithm
	}

	if plan.KeySizeBits.IsUnknown() {
		plan.KeySizeBits = state.KeySizeBits
	}

	if plan.Curve.IsUnknown() {
		plan.Curve = state.Curve
	}

	if plan.ParameterSet.IsUnknown() {
		plan.ParameterSet = state.ParameterSet
	}

	if plan.SignatureAlgorithm.IsUnknown() {
		plan.SignatureAlgorithm = state.SignatureAlgorithm
	}

	if plan.CustodyModel.IsUnknown() {
		plan.CustodyModel = state.CustodyModel
	}

	if plan.SatelliteID.IsUnknown() {
		plan.SatelliteID = state.SatelliteID
	}

	if plan.AutoRenew.IsUnknown() {
		plan.AutoRenew = state.AutoRenew
	}

	if plan.AutoDeploy.IsUnknown() {
		plan.AutoDeploy = state.AutoDeploy
	}

	if plan.KeyName.IsUnknown() {
		plan.KeyName = state.KeyName
	}

	if plan.KeyRef.IsUnknown() {
		plan.KeyRef = state.KeyRef
	}
}

func certificateConfigHasUnknowns(
	config certificateResourceModel,
) bool {
	return config.Type.IsUnknown() ||
		config.CommonName.IsUnknown() ||
		config.ExpireDays.IsUnknown() ||
		config.Signer.IsUnknown() ||
		config.SatelliteID.IsUnknown() ||
		config.Storage.IsUnknown() ||
		config.Country.IsUnknown() ||
		config.Organization.IsUnknown() ||
		config.OrganizationalUnit.IsUnknown() ||
		config.State.IsUnknown() ||
		config.Locality.IsUnknown() ||
		config.Serial.IsUnknown() ||
		config.KeyUsage.IsUnknown() ||
		config.ExtendedKeyUsage.IsUnknown() ||
		config.DNSSANs.IsUnknown() ||
		config.KeyAlgorithm.IsUnknown() ||
		config.KeySizeBits.IsUnknown() ||
		config.Curve.IsUnknown() ||
		config.ParameterSet.IsUnknown() ||
		config.SignatureAlgorithm.IsUnknown() ||
		config.CustodyModel.IsUnknown() ||
		config.AutoRenew.IsUnknown() ||
		config.AutoDeploy.IsUnknown()
}
