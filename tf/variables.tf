variable "do_region" {
  type        = string
  description = "Region to deploy to"
}

variable "do_spaces_name" {
  type        = string
  description = "Name of your DigitalOcean Space/Bucket"
}

variable "spaces_access_id" {
  type        = string
  description = "Access ID for your DigitalOcean Space/Bucket"
}

variable "spaces_secret_key" {
  type        = string
  description = "Secret Key for your DigitalOcean Space/Bucket"
}
