action "securitls_deploy_attachment" "web01_cert" {
  config {
    device_id = securitls_device.web01.id
    cert_id   = securitls_certificate.leaf.id
  }
}
