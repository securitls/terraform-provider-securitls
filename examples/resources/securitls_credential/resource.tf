
resource "securitls_credential" "secret" {
  name   = "secret example"
  type   = "secret"
  secret = "***"
}

resource "securitls_credential" "rsa" {
  name   = "rsa example"
  type   = "rsa"
  key = <<EOT
-----BEGIN RSA PRIVATE KEY-----
***
-----END RSA PRIVATE KEY-----
EOT
}