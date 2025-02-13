// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &StorytelFastlyProvider{}
var _ provider.ProviderWithFunctions = &StorytelFastlyProvider{}
var _ provider.ProviderWithEphemeralResources = &StorytelFastlyProvider{}

type StorytelFastlyProvider struct {
	version string
}

// StorytelFastlyProviderModel describes the provider data model.
type StorytelFastlyProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

func (p *StorytelFastlyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "storytelfastly"
	resp.Version = p.version
}

func (p *StorytelFastlyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Fastly API key",
				Required:            true,
			},
		},
	}
}

func (p *StorytelFastlyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data StorytelFastlyProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if data.APIKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_Key"),
			"Unknown Fastly SecretStore Entries API key",
			"Fastly SecretStore Entries client cannot be created because the API key is unknown.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := fastly.NewClient(data.APIKey.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("failed to initialize fastly client", "")
		return
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *StorytelFastlyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSecretStoreEntriesResource,
	}
}

func (p *StorytelFastlyProvider) EphemeralResources(ctx context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *StorytelFastlyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *StorytelFastlyProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &StorytelFastlyProvider{
			version: version,
		}
	}
}
