terraform {
  required_providers {
    securitls = {
      source = "securitls/securitls"
    }
  }
}

provider "securitls" {
  api_key = "<YOUR API KEY>" # or set environment variable SECURITLS_API_KEY
}
