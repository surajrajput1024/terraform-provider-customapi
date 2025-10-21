package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-customapi/go-customapi/client"
	clienttypes "terraform-provider-customapi/go-customapi/client/types"
)

type CustomAPIDataSource struct{
	client *client.CustomAPIClient
}

type CustomAPIDataSourceModel struct{
	Endpoint   types.String `tfsdk:"endpoint"`
	OrgID      types.String `tfsdk:"org_id"`
	QueryParams map[string]types.String `tfsdk:"query_params"`
	Response   types.String `tfsdk:"response"`
	StatusCode types.Int64  `tfsdk:"status_code"`
	Success    types.Bool   `tfsdk:"success"`
	Error      types.String `tfsdk:"error"`
}

func NewCustomAPIDataSource() datasource.DataSource {
	return &CustomAPIDataSource{}
}

func (d *CustomAPIDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_source"
}

func (d *CustomAPIDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.CustomAPIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.CustomAPIClient, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *CustomAPIDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "API endpoint to call",
			},
			"org_id": schema.StringAttribute{
				Optional:    true,
				Description: "Organization ID",
			},
			"query_params": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Query parameters",
			},
			"response": schema.StringAttribute{
				Computed:    true,
				Description: "API response body",
			},
			"status_code": schema.Int64Attribute{
				Computed:    true,
				Description: "HTTP status code",
			},
			"success": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the request was successful",
			},
			"error": schema.StringAttribute{
				Computed:    true,
				Description: "Error message if request failed",
			},
		},
	}
}

func (d *CustomAPIDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CustomAPIDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := d.client

	queryParams := make(map[string]string)
	for key, value := range data.QueryParams {
		queryParams[key] = value.ValueString()
	}

	apiReq := &clienttypes.CustomAPIRequest{
		Method:      "GET",
		URL:         data.Endpoint.ValueString(),
		QueryParams: queryParams,
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}

	if !data.OrgID.IsNull() {
		apiReq.Headers["current-organization"] = data.OrgID.ValueString()
	}

	apiResp, err := apiClient.MakeRequest(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Request Failed",
			fmt.Sprintf("Failed to make API request: %v", err),
		)
		return
	}

	data.Response = types.StringValue(string(apiResp.Body))
	data.StatusCode = types.Int64Value(int64(apiResp.StatusCode))
	data.Success = types.BoolValue(apiResp.Success)
	
	if !apiResp.Success {
		data.Error = types.StringValue(apiResp.Error)
	}

	tflog.Debug(ctx, "Data source read completed", map[string]interface{}{
		"endpoint":    data.Endpoint.ValueString(),
		"status_code": apiResp.StatusCode,
		"success":     apiResp.Success,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
