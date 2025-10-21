package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-customapi/go-customapi/client"
)

type CustomAPIProvider struct {
	version string
}

type CustomAPIProviderModel struct {
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	AuthToken   types.String `tfsdk:"auth_token"`
	Environment types.String `tfsdk:"environment"`
	BaseURL     types.String `tfsdk:"base_url"`
	OrgID       types.String `tfsdk:"org_id"`
}

func (p *CustomAPIProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "customapi"
	resp.Version = p.version
}

func (p *CustomAPIProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "Username for authentication",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password for authentication",
			},
			"auth_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Auth token for direct authentication",
			},
			"environment": schema.StringAttribute{
				Optional:    true,
				Description: "Environment (qa, staging, prod)",
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL for the API",
			},
			"org_id": schema.StringAttribute{
				Optional:    true,
				Description: "Organization ID",
			},
		},
	}
}

func (p *CustomAPIProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config CustomAPIProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Load environment configuration
	envConfig, err := client.LoadConfig()
	if err != nil {
		resp.Diagnostics.AddWarning(
			"Failed to load environment config",
			fmt.Sprintf("Using defaults: %v", err),
		)
	}

	// Use provider config values if provided, otherwise fall back to environment
	username := config.Username.ValueString()
	if username == "" {
		username = envConfig.Username
	}

	password := config.Password.ValueString()
	if password == "" {
		password = envConfig.Password
	}

	authToken := config.AuthToken.ValueString()
	if authToken == "" {
		authToken = envConfig.AuthToken
	}

	environment := config.Environment.ValueString()
	if environment == "" {
		environment = envConfig.Environment
	}

	baseURL := config.BaseURL.ValueString()
	if baseURL == "" {
		baseURL = envConfig.BaseURL
	}

	orgID := config.OrgID.ValueString()
	if orgID == "" {
		orgID = envConfig.DefaultOrgID
	}

	if authToken == "" && (username == "" || password == "") {
		resp.Diagnostics.AddError(
			"Missing Authentication",
			"Either auth_token or both username and password must be provided in provider config or environment variables",
		)
		return
	}

	authConfig := &client.AuthConfig{
		Username:    username,
		Password:    password,
		AuthToken:   authToken,
		Environment: environment,
		BaseURL:     envConfig.AuthURL,
	}

	apiClient := client.NewCustomAPIClient(authConfig, baseURL)

	ctx = tflog.SetField(ctx, "customapi_provider", "configured")
	tflog.Info(ctx, "CustomAPI provider configured", map[string]interface{}{
		"environment": environment,
		"base_url":    baseURL,
		"org_id":      orgID,
	})

	resp.ResourceData = apiClient
	resp.DataSourceData = apiClient
}

func (p *CustomAPIProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCustomAPIResource,
	}
}

func (p *CustomAPIProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCustomAPIDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CustomAPIProvider{
			version: version,
		}
	}
}
