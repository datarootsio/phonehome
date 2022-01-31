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
