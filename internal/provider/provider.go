package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure FireflyProvider satisfies various provider interfaces.
var _ provider.Provider = &FireflyProvider{}

// FireflyProvider defines the provider implementation.
type FireflyProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// FireflyProviderModel describes the provider data model.
type FireflyProviderModel struct {
	Endpoint    types.String `tfsdk:"endpoint"`
	AccessToken types.String `tfsdk:"access_token"`
}

func (p *FireflyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "firefly"
	resp.Version = p.version
}

func (p *FireflyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The URL of the Firefly instance, with an optional port, such as <http://firefly.local> or <http://firefly.local:8000>",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^https?://[^:/]+(:\d+)?$`),
						"must be a URL, like http://firefly.local or http://firefly.local:8000",
					),
				},
			},
			"access_token": schema.StringAttribute{
				MarkdownDescription: "A Personal Access Token generated on the Firefly web API. See [the docs](https://docs.firefly-iii.org/firefly-iii/api/#personal-access-token) for instructions.",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

type fireflyTransport struct {
	baseURL     string
	accessToken string
	rt          http.RoundTripper
}

func (t *fireflyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	parsed, err := url.Parse(t.baseURL)
	if err != nil {
		return nil, err
	}
	r.URL.Scheme = parsed.Scheme
	r.URL.Host = parsed.Host
	r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", t.accessToken))
	return t.rt.RoundTrip(r)
}

func (p *FireflyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data FireflyProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	// if data.Endpoint.IsNull() { /* ... */ }

	client := http.DefaultClient
	client.Transport = &fireflyTransport{
		baseURL:     data.Endpoint.ValueString(),
		accessToken: data.AccessToken.ValueString(),
		rt:          http.DefaultTransport,
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *FireflyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *FireflyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewSysInfoDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FireflyProvider{
			version: version,
		}
	}
}
