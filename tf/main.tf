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

// create a digital ocean app from a container in my digital ocean container repository
resource "digitalocean_app" "app" {
  name = var.do_app_name

  spec {
    services {
      name = var.do_app_name

      image {
        repository_type = "DOCR"
        repository      = var.do_repository_name
        deploy_on_push  = "enabled"
        tag             = "latest"
      }

      http_port = 80
    }
  }
}
