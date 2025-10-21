# Terraform Provider for Custom API

A custom Terraform provider that enables you to interact with REST APIs through Terraform data sources and resources.

## Features

- **Data Sources**: Read data from REST APIs
- **Resources**: Full CRUD operations for API resources
- **Authentication**: Support for Bearer token and username/password authentication
- **Environment Configuration**: All configuration via environment variables
- **Flexible API Calls**: Support for any REST endpoint with custom headers and query parameters

## Installation

### Prerequisites

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/dl/) >= 1.19 (for building from source)

### Building from Source

1. Clone the repository:
```bash
git clone <repository-url>
cd terraform-provider-customapi
```

2. Build the provider:
```bash
go build -o terraform-provider-customapi main.go
```

3. Install the provider:
```bash
# Create plugin directory
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hashicorp/customapi/1.0.0/darwin_arm64

# Copy the binary
cp terraform-provider-customapi ~/.terraform.d/plugins/registry.terraform.io/hashicorp/customapi/1.0.0/darwin_arm64/
```

## Configuration

### Environment Variables

Create a `.env` file in your project root or set the following environment variables:

```bash
# API Configuration
CUSTOMAPI_BASE_URL=https://your-api.example.com
CUSTOMAPI_AUTH_URL=https://your-auth.example.com
CUSTOMAPI_ENVIRONMENT=production
CUSTOMAPI_ORG_ID=your-org-id

# Authentication (choose one method)
# Method 1: Direct token
CUSTOMAPI_AUTH_TOKEN=your-jwt-token

# Method 2: Username/Password (for OAuth2)
CUSTOMAPI_USERNAME=your-username
CUSTOMAPI_PASSWORD=your-password
CUSTOMAPI_CLIENT_ID=your-client-id
CUSTOMAPI_AUDIENCE=your-audience
```

### Provider Configuration

```hcl
terraform {
  required_providers {
    customapi = {
      source  = "hashicorp/customapi"
      version = "~> 1.0"
    }
  }
}

provider "customapi" {
  # Optional: Override environment variables
  auth_token = "your-jwt-token"  # Optional if set in env
  base_url   = "https://your-api.example.com"  # Optional if set in env
  org_id     = "your-org-id"  # Optional if set in env
}
```

## Usage

### Data Source

Read data from any API endpoint:

```hcl
data "customapi_data_source" "user_profile" {
  endpoint = "/api/users/profile/me"
}

output "user_profile" {
  value = data.customapi_data_source.user_profile.response
}
```

### Resource

Manage API resources with full CRUD operations:

```hcl
resource "customapi_resource" "user" {
  endpoint = "/api/users"
  method   = "POST"
  body = jsonencode({
    name  = "John Doe"
    email = "john@example.com"
  })
}
```

## Development

### Running the Provider in Debug Mode

1. Start the provider server:
```bash
go run main.go --debug
```

2. Set the reattach environment variable (output from step 1):
```bash
export TF_REATTACH_PROVIDERS='{"registry.terraform.io/hashicorp/customapi":{"Protocol":"grpc","ProtocolVersion":6,"Pid":XXXXX,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/..."}}}'
```

3. Run Terraform (skip `terraform init` when using TF_REATTACH_PROVIDERS):
```bash
cd test-workspace
terraform plan
```

### Testing

1. Navigate to the test workspace:
```bash
cd test-workspace
```

2. Update the `main.tf` with your API details:
```hcl
provider "customapi" {
  auth_token = "your-jwt-token"
  base_url   = "https://your-api.example.com"
  org_id     = "your-org-id"
}
```

3. Run Terraform:
```bash
terraform plan
terraform apply
```

## API Response Format

The provider expects API responses in the following format:

```json
{
  "success": true,
  "data": {
    // Your actual data here
  },
  "message": "Success message"
}
```

## Error Handling

The provider handles various error scenarios:
- Authentication failures
- Network timeouts
- Invalid JSON responses
- HTTP error status codes

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:
1. Check the [Issues](https://github.com/your-repo/issues) page
2. Create a new issue with detailed information
3. Include logs and configuration details

## Changelog

### v1.0.0
- Initial release
- Data source support
- Resource CRUD operations
- Bearer token authentication
- Username/password authentication
- Environment variable configuration
