package securitls

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ action.Action = &validateAttachmentAction{}
var _ action.ActionWithConfigure = &validateAttachmentAction{}

type validateAttachmentAction struct {
	client *Client
}

type validateAttachmentActionModel struct {
	DeviceID types.String `tfsdk:"device_id"`
	CertID   types.String `tfsdk:"cert_id"`
}

func NewValidateAttachmentAction() action.Action {
	return &validateAttachmentAction{}
}

func (a *validateAttachmentAction) Metadata(
	_ context.Context,
	req action.MetadataRequest,
	resp *action.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_validate_attachment"
}

func (a *validateAttachmentAction) Schema(
	_ context.Context,
	_ action.SchemaRequest,
	resp *action.SchemaResponse,
) {
	resp.Schema = actionschema.Schema{
		Description: "Validate a certificate attachment on a SecuriTLS device.",

		Attributes: map[string]actionschema.Attribute{
			"device_id": actionschema.StringAttribute{
				Required:    true,
				Description: "SecuriTLS device ID.",
			},

			"cert_id": actionschema.StringAttribute{
				Required:    true,
				Description: "Certificate ID attached to the device.",
			},
		},
	}
}

func (a *validateAttachmentAction) Configure(
	_ context.Context,
	req action.ConfigureRequest,
	resp *action.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf(
				"Expected *Client, got %T",
				req.ProviderData,
			),
		)
		return
	}

	a.client = client
}

func (a *validateAttachmentAction) Invoke(
	ctx context.Context,
	req action.InvokeRequest,
	resp *action.InvokeResponse,
) {
	var config validateAttachmentActionModel

	resp.Diagnostics.Append(
		req.Config.Get(ctx, &config)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.client == nil {
		resp.Diagnostics.AddError(
			"SecuriTLS client not configured",
			"The provider client was not configured before invoking the validate action.",
		)
		return
	}

	deviceID := config.DeviceID.ValueString()
	certID := config.CertID.ValueString()

	if deviceID == "" {
		resp.Diagnostics.AddError(
			"Invalid device_id",
			"device_id cannot be empty.",
		)
		return
	}

	if certID == "" {
		resp.Diagnostics.AddError(
			"Invalid cert_id",
			"cert_id cannot be empty.",
		)
		return
	}

	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf(
				"Validating certificate %s on device %s",
				certID,
				deviceID,
			),
		})
	}

	var out map[string]any
	status, err := a.client.Do(
		ctx,
		http.MethodPost,
		fmt.Sprintf(
			"/devices/%s/validate/%s",
			deviceID,
			certID,
		),
		nil,
		&out,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to validate certificate attachment",
			fmt.Sprintf(
				"SecuriTLS returned HTTP %d while validating certificate %s on device %s: %s",
				status,
				certID,
				deviceID,
				err,
			),
		)
		return
	}

	b, err := json.Marshal(out)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to encode response",
			err.Error(),
		)
		return
	}

	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf(
				"Ran validate for certificate %s on device %s: %s",
				certID,
				deviceID,
				string(b),
			),
		})
	}
}
