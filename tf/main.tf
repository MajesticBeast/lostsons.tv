terraform {
  required_providers {
    digitalocean = {
      source  = "digitalocean/digitalocean"
      version = "2.30.0"
    }
  }
}

provider "digitalocean" {
  spaces_access_id  = var.spaces_access_id
  spaces_secret_key = var.spaces_secret_key
}

// create a digital ocean space (bucket)
resource "digitalocean_spaces_bucket" "bucket" {
  name   = var.do_spaces_name
  region = var.do_region
  acl    = "public-read"

  lifecycle_rule {
    enabled = true
    expiration {
      days = 1
    }
  }
}

