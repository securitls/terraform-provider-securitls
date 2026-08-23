resource "securitls_device" "web01" {
  name       = "web01"
  credential = securitls_credential.ssh.id
  hostname   = "web01.internal.example.com"
  port       = 22
  username   = "deployuser"
  crl_type   = "none"

  attachment {
  	cert_id      = securitls_certificate.leaf.id
  	path         = "/etc/pki/tls/certs/myleaf.crt"
  }
}