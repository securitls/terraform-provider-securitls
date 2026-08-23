package securitls

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/action"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ action.Action = &deployAttachmentAction{}
var _ action.ActionWithConfigure = &deployAttachmentAction{}

type deployAttachmentAction struct {
	client *Client
}

type deployAttachmentActionModel struct {
	DeviceID types.String `tfsdk:"device_id"`
	CertID   types.String `tfsdk:"cert_id"`
}

func NewDeployAttachmentAction() action.Action {
	return &deployAttachmentAction{}
}

func (a *deployAttachmentAction) Metadata(
	_ context.Context,
	req action.MetadataRequest,
	resp *action.MetadataResponse,
) {
	resp.TypeName = req.ProviderTypeName + "_deploy_attachment"
}

func (a *deployAttachmentAction) Schema(
	_ context.Context,
	_ action.SchemaRequest,
	resp *action.SchemaResponse,
) {
	resp.Schema = actionschema.Schema{
		Description: "Deploy a certificate attachment to a SecuriTLS device.",

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

func (a *deployAttachmentAction) Configure(
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

func (a *deployAttachmentAction) Invoke(
	ctx context.Context,
	req action.InvokeRequest,
	resp *action.InvokeResponse,
) {
	var config deployAttachmentActionModel

	resp.Diagnostics.Append(
		req.Config.Get(ctx, &config)...,
	)

	if resp.Diagnostics.HasError() {
		return
	}

	if a.client == nil {
		resp.Diagnostics.AddError(
			"SecuriTLS client not configured",
			"The provider client was not configured before invoking the deploy action.",
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
				"Deploying certificate %s to device %s",
				certID,
				deviceID,
			),
		})
	}

	status, err := a.client.Do(
		ctx,
		http.MethodPost,
		fmt.Sprintf(
			"/devices/%s/deploy/%s",
			deviceID,
			certID,
		),
		nil,
		nil,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to deploy certificate attachment",
			fmt.Sprintf(
				"SecuriTLS returned HTTP %d while deploying certificate %s to device %s: %s",
				status,
				certID,
				deviceID,
				err,
			),
		)
		return
	}

	if resp.SendProgress != nil {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf(
				"Certificate %s deployed to device %s",
				certID,
				deviceID,
			),
		})
	}
}
