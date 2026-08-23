package securitls

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func configureResource(req resource.ConfigureRequest, client **Client) {
	if req.ProviderData == nil {
		return
	}
	*client = req.ProviderData.(*Client)
}

func configureDataSource(req datasource.ConfigureRequest, client **Client) {
	if req.ProviderData == nil {
		return
	}
	*client = req.ProviderData.(*Client)
}
