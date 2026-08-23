# Terraform Provider for SecuriTLS

A Terraform Plugin Framework provider for the SecuriTLS API

## Provider configuration

```
terraform {
  required_providers {
    securitls = {
      source = "securitls/securitls"
    }
  }
}

provider "securitls" {
  # Prefer SECURITLS_API_KEY in CI.
  api_key = var.securitls_api_key
  # base_url = "https://securitls.com/api"
}
```

The provider authenticates through `POST /api/authenticate`, caches the returned one-hour JWT, and transparently re-authenticates after a 401.

## Resources

- `securitls_organization`
- `securitls_credential`
- `securitls_storage_provider`
- `securitls_storage_key`
- `securitls_device`
- `securitls_certificate`

## Actions

- `securitls_deploy_attachment`
- `securitls_validate_attachment`

## Certificate example

```
resource "securitls_certificate" "root" {
  type                 = "root"
  common_name          = "Example Root CA"
  expire_interval_days = 3650

  key_usage = [
    "keyCertSign",
    "cRLSign",
    "digitalSignature",
  ]

  key_algorithm       = "rsa"
  key_size_bits       = 4096
  signature_algorithm = "sha384"
}

resource "securitls_certificate" "leaf" {
  type                 = "leaf"
  common_name          = "app.example.com"
  expire_interval_days = 47
  signer               = securitls_certificate.root.id

  key_usage          = ["digitalSignature"]
  extended_key_usage = ["serverAuth"]
  dns_sans            = ["app.example.com", "www.app.example.com"]

  key_algorithm       = "ec"
  curve               = "prime256v1"
  signature_algorithm = "sha256"
}

output "leaf_pem" {
  value = securitls_certificate.leaf.pem
}
```

### Lifecycle semantics

Certificate resources are modeled as a stable Terraform resource address that may advance to a new SecuriTLS certificate record over time.

Changing certificate configuration such as the common name, signer, SANs, key settings, expiration interval, or other certificate properties causes SecuriTLS to **reissue** the certificate. The Terraform resource remains the same logical resource, but its computed `id` is updated to the successor certificate ID returned by SecuriTLS.

This allows dependent resources to follow certificate lineage naturally. For example:

```
resource "securitls_certificate" "leaf" {
  type                 = "leaf"
  common_name          = "app.example.com"
  expire_interval_days = 47
  signer               = securitls_certificate.root.id
}
```

If `securitls_certificate.root` is reissued and receives a new certificate ID, Terraform sees the changed `signer` value on the leaf and reissues the leaf as well.

The certificate type itself is replacement-only. Changing between `root`, `intermediate`, and `leaf` causes Terraform to replace the resource rather than perform a certificate lifecycle operation.

SecuriTLS also exposes explicit lifecycle triggers:

```
resource "securitls_certificate" "root" {
  type                 = "root"
  common_name          = "Example Root CA"
  expire_interval_days = 3650

  renew_trigger = "1"
}
```

The available triggers are:

```text
renew_trigger
rekey_trigger
reissue_trigger
```

A trigger executes its corresponding lifecycle operation whenever its value changes to a new non-null string.

For example:

```
renew_trigger = "1"
```

can later be changed to:

```
renew_trigger = "2"
```

to perform another renewal.

Trigger values are opaque to SecuriTLS and only tracked in Terraform state. They may be counters, timestamps, release identifiers, or any other string whose change represents a new requested operation:

```
rekey_trigger   = "rotation-3"
reissue_trigger = "2026-08-22"
```

Trigger behavior is:

```text
null  -> "1"       execute operation
"1"   -> "2"       execute operation
"abc" -> "def"     execute operation
"1"   -> "1"       no operation
"1"   -> null      no operation
```

Removing a trigger only removes the trigger value from Terraform state; it does not perform another certificate lifecycle operation.

Only one explicit lifecycle operation may be triggered during a single apply. `renew_trigger`, `rekey_trigger`, and `reissue_trigger` should not be changed simultaneously.

Certificate configuration changes already imply a reissue. Therefore:

- configuration change by itself → reissue
- `reissue_trigger` change by itself → reissue
- configuration change together with `reissue_trigger` → one reissue
- configuration change together with `renew_trigger` → rejected
- configuration change together with `rekey_trigger` → rejected

Renew, rekey, and reissue create a successor certificate in SecuriTLS. After the operation succeeds, the provider updates the resource's computed `id`, certificate metadata, and `pem` to represent that successor certificate. The Terraform resource address itself does not change.

For example:

```text
securitls_certificate.root
```

remains the same Terraform resource even though:

```text
id = old-certificate-id
```

may become:

```text
id = successor-certificate-id
```

after a lifecycle operation.

This keeps Terraform dependencies aligned with the currently active certificate without requiring destroy/create replacement semantics for normal certificate lifecycle operations.

## Local development

```bash
go mod tidy
go test ./...
go build ./...
```
