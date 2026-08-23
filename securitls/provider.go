package securitls

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &securitlsProvider{}
var _ provider.ProviderWithActions = &securitlsProvider{}

func New() provider.Provider { return &securitlsProvider{} }

type securitlsProvider struct{}

type securitlsProviderModel struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}

func (p *securitlsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "securitls"
}

func (p *securitlsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for the SecuriTLS PKI and certificate lifecycle management API.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "SecuriTLS API key. Can also be set with SECURITLS_API_KEY.",
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "SecuriTLS API base URL. Defaults to https://securitls.com/api and can be set with SECURITLS_BASE_URL.",
			},
		},
	}
}

func (p *securitlsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config securitlsProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(path.Root("api_key"), "Unknown SecuriTLS API key", "The API key must be known before provider configuration.")
		return
	}

	apiKey := os.Getenv("SECURITLS_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddAttributeError(path.Root("api_key"), "Missing SecuriTLS API key", "Set api_key or SECURITLS_API_KEY.")
		return
	}

	baseURL := os.Getenv("SECURITLS_BASE_URL")
	if baseURL == "" {
		baseURL = "https://securitls.com/api"
	}
	if !config.BaseURL.IsNull() && !config.BaseURL.IsUnknown() && config.BaseURL.ValueString() != "" {
		baseURL = config.BaseURL.ValueString()
	}

	client := NewClient(baseURL, apiKey)
	if err := client.authenticate(ctx); err != nil {
		resp.Diagnostics.AddError("Unable to authenticate to SecuriTLS", err.Error())
		return
	}
	resp.ResourceData = client
	resp.DataSourceData = client
	resp.ActionData = client
}

func (p *securitlsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrganizationResource,
		NewCredentialResource,
		NewStorageProviderResource,
		NewStorageKeyResource,
		NewDeviceResource,
		NewCertificateResource,
	}
}

func (p *securitlsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// NewOrganizationsDataSource,
		// NewCredentialsDataSource,
		// NewDevicesDataSource,
		// NewCertificatesDataSource,
		// NewCertificateDataSource,
		// NewStorageProvidersDataSource,
	}
}

func (p *securitlsProvider) Actions(_ context.Context) []func() action.Action {
	return []func() action.Action{
		NewDeployAttachmentAction,
		NewValidateAttachmentAction,
	}
}
