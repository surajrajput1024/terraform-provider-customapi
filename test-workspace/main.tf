terraform {
  required_providers {
    customapi = {
      source  = "hashicorp/customapi"
      version = "~> 1.0"
    }
  }
}

provider "customapi" {
  auth_token = ""
  base_url    = "https://qa-iot-api.pace.on-device.ai"
  org_id      = "90241446-c32b-49a5-a84b-a8e98401b52a"
}

data "customapi_data_source" "user_profile" {
  endpoint = "/api/users/profile/me"
}

output "user_profile" {
  value = data.customapi_data_source.user_profile.response
}
