variable "pg_main_user" {
  type    = string
  default = "admin"
}

variable "pg_main_pass_secret" {
  type    = string
  default = "pg-main-pass"
}

variable "gcp_project_id" {
  type    = string
  default = "phonehome-339613"
}

variable "current_version" {
  default     = "develop"
  type        = string
  description = "A version identifying the infrastructure that will be deployed. Currently should point to the commit hash."
}

variable "phonehome_a_records" {
  default = {
    "@" = {
      ttl = "3600"
      ip  = "34.107.171.221"
    }
  }
}
