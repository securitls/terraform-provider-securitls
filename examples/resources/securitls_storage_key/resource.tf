resource "securitls_storage_key" "aws" {
  type = "awskms"
  name = "example-aws"
  role_arn = "arn:aws:iam::012345678901:role/example-role"
  key_arn = "arn:aws:kms:us-east-1:012345678901:key/00000000-0000-0000-0000-000000000000"
  external_id = "***"
}

resource "securitls_storage_key" "azure" {
  type = "azure"
  name = "example-azure"
  vault_url = "https://example.vault.azure.net"
  key_name = "example-azure-keyname"
  key_version = "00000000000000000000000000000000"
  tenant_id = "00000000-0000-0000-0000-000000000000"
  client_id = "00000000-0000-0000-0000-000000000000"
  client_secret = "***"
}
