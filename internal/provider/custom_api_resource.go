package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"terraform-provider-customapi/go-customapi/client"
	clienttypes "terraform-provider-customapi/go-customapi/client/types"
)

type CustomAPIResource struct{
	client *client.CustomAPIClient
}

type CustomAPIResourceModel struct{
	ID          types.String `tfsdk:"id"`
	Endpoint    types.String `tfsdk:"endpoint"`
	Method      types.String `tfsdk:"method"`
	Body        types.String `tfsdk:"body"`
	OrgID       types.String `tfsdk:"org_id"`
	Headers     map[string]types.String `tfsdk:"headers"`
	QueryParams map[string]types.String `tfsdk:"query_params"`
	Response    types.String `tfsdk:"response"`
	StatusCode  types.Int64  `tfsdk:"status_code"`
	Success     types.Bool   `tfsdk:"success"`
	Error       types.String `tfsdk:"error"`
}

func NewCustomAPIResource() resource.Resource {
	return &CustomAPIResource{}
}

func (r *CustomAPIResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (r *CustomAPIResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.CustomAPIClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.CustomAPIClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *CustomAPIResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier",
			},
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "API endpoint",
			},
			"method": schema.StringAttribute{
				Required:    true,
				Description: "HTTP method (GET, POST, PUT, DELETE)",
			},
			"body": schema.StringAttribute{
				Optional:    true,
				Description: "Request body",
			},
			"org_id": schema.StringAttribute{
				Optional:    true,
				Description: "Organization ID",
			},
			"headers": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Custom headers",
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

func (r *CustomAPIResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CustomAPIResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client

	apiReq := r.buildAPIRequest(data)

	apiResp, err := apiClient.MakeRequest(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Request Failed",
			fmt.Sprintf("Failed to make API request: %v", err),
		)
		return
	}

	r.updateModelFromResponse(&data, apiResp)
	data.ID = types.StringValue(fmt.Sprintf("%s-%s", data.Endpoint.ValueString(), data.Method.ValueString()))

	tflog.Debug(ctx, "Resource created", map[string]interface{}{
		"endpoint":    data.Endpoint.ValueString(),
		"method":      data.Method.ValueString(),
		"status_code": apiResp.StatusCode,
		"success":     apiResp.Success,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomAPIResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CustomAPIResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client

	apiReq := r.buildAPIRequest(data)

	apiResp, err := apiClient.MakeRequest(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Request Failed",
			fmt.Sprintf("Failed to make API request: %v", err),
		)
		return
	}

	r.updateModelFromResponse(&data, apiResp)

	tflog.Debug(ctx, "Resource read", map[string]interface{}{
		"endpoint":    data.Endpoint.ValueString(),
		"method":      data.Method.ValueString(),
		"status_code": apiResp.StatusCode,
		"success":     apiResp.Success,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomAPIResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CustomAPIResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client

	apiReq := r.buildAPIRequest(data)

	apiResp, err := apiClient.MakeRequest(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Request Failed",
			fmt.Sprintf("Failed to make API request: %v", err),
		)
		return
	}

	r.updateModelFromResponse(&data, apiResp)

	tflog.Debug(ctx, "Resource updated", map[string]interface{}{
		"endpoint":    data.Endpoint.ValueString(),
		"method":      data.Method.ValueString(),
		"status_code": apiResp.StatusCode,
		"success":     apiResp.Success,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomAPIResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CustomAPIResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	apiClient := r.client

	apiReq := r.buildAPIRequest(data)

	apiResp, err := apiClient.MakeRequest(ctx, apiReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Request Failed",
			fmt.Sprintf("Failed to make API request: %v", err),
		)
		return
	}

	tflog.Debug(ctx, "Resource deleted", map[string]interface{}{
		"endpoint":    data.Endpoint.ValueString(),
		"method":      data.Method.ValueString(),
		"status_code": apiResp.StatusCode,
		"success":     apiResp.Success,
	})
}

func (r *CustomAPIResource) buildAPIRequest(data CustomAPIResourceModel) *clienttypes.CustomAPIRequest {
	headers := make(map[string]string)
	for key, value := range data.Headers {
		headers[key] = value.ValueString()
	}

	queryParams := make(map[string]string)
	for key, value := range data.QueryParams {
		queryParams[key] = value.ValueString()
	}

	apiReq := &clienttypes.CustomAPIRequest{
		Method:      data.Method.ValueString(),
		URL:         data.Endpoint.ValueString(),
		Headers:     headers,
		QueryParams: queryParams,
	}

	if !data.Body.IsNull() && !data.Body.IsUnknown() {
		apiReq.Body = []byte(data.Body.ValueString())
	}

	if !data.OrgID.IsNull() && !data.OrgID.IsUnknown() {
		apiReq.Headers["current-organization"] = data.OrgID.ValueString()
	}

	return apiReq
}

func (r *CustomAPIResource) updateModelFromResponse(data *CustomAPIResourceModel, apiResp *clienttypes.CustomAPIResponse) {
	data.Response = types.StringValue(string(apiResp.Body))
	data.StatusCode = types.Int64Value(int64(apiResp.StatusCode))
	data.Success = types.BoolValue(apiResp.Success)
	
	if !apiResp.Success {
		data.Error = types.StringValue(apiResp.Error)
	}
}
