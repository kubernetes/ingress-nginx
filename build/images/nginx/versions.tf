
terraform {
  required_version = ">= 0.12"

  required_providers {
    aws   = "~> 2.0"
    tls   = "~> 2.0"
    local = "~> 1.3"
    null  = "~> 2.1"
  }
}
