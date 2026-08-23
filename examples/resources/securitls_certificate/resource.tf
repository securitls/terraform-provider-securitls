# CREATE ROOT CA
resource "securitls_certificate" "root" {
  type                 = "root"
  common_name          = "Example Corp Root"
  expire_interval_days = 3650
}

# CREATE INTERMEDIATE CA
resource "securitls_certificate" "intermediate" {
  type                 = "intermediate"
  common_name          = "Example Corp Intermediate"
  expire_interval_days = 1825
  signer               = securitls_certificate.root.id
}

# CREATE LEAF CERT
resource "securitls_certificate" "leaf" {
  type                 = "leaf"
  common_name          = "Example Corp Leaf"
  expire_interval_days = 47
  signer               = securitls_certificate.intermediate.id
}