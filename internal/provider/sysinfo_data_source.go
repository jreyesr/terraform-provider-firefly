package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &SysInfoDataSource{}
	_ datasource.DataSourceWithConfigure = &SysInfoDataSource{}
)

func NewSysInfoDataSource() datasource.DataSource {
	return &SysInfoDataSource{}
}

// SysInfoDataSource defines the data source implementation.
type SysInfoDataSource struct {
	client *http.Client
}

// SysInfoDataSourceModel describes the data source data model.
// No parameters required, the system info is global to the provider instance
type SysInfoDataSourceModel struct {
	Version    types.String `tfsdk:"version"`
	APIVersion types.String `tfsdk:"api_version"`
	PHPVersion types.String `tfsdk:"php_version"`
	OS         types.String `tfsdk:"os"`
	DBDriver   types.String `tfsdk:"driver"`
}

func (d *SysInfoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sysinfo"
}

func (d *SysInfoDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Exposes general system information and versions of the supporting software",

		Attributes: map[string]schema.Attribute{
			"version":     schema.StringAttribute{Computed: true},
			"api_version": schema.StringAttribute{Computed: true},
			"php_version": schema.StringAttribute{Computed: true},
			"os":          schema.StringAttribute{Computed: true},
			"driver":      schema.StringAttribute{Computed: true},
		},
	}
}

func (d *SysInfoDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *SysInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SysInfoDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	httpReq, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/about", "https://demo.firefly-iii.org"), nil)
	httpResp, err := d.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	var respData struct {
		Data SysInfoDataSourceModel `json:"data"`
	}
	json.NewDecoder(httpResp.Body).Decode(&respData)

	data.Version = types.StringValue(respData.Data.Version.String())
	data.APIVersion = types.StringValue(respData.Data.APIVersion.String())
	data.PHPVersion = types.StringValue(respData.Data.PHPVersion.String())
	data.OS = types.StringValue(respData.Data.OS.String())
	data.DBDriver = types.StringValue(respData.Data.DBDriver.String())

	// Write logs using the tflog package
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
