resource "securitls_storage_provider" "aws" {
  kind = "s3-self-managed"
  label = "example-aws-storage"
  aws_mode = "assume"
  aws_region = "us-east-1"
  aws_bucket_name = "securitls-s3bucket"
  aws_role_arn = "arn:aws:iam::012345678901:role/example-role"
  aws_external_id = "***"
}

resource "securitls_storage_provider" "sia" {
  kind = "sia-self-managed"
  label = "example-sia-storage"
  sia_bucket_name = "securitls-siabucket"
  sia_endpoint = "http://hostd.example.com:9980"
  sia_password = "***"
}