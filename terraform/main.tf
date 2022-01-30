terraform {
  backend "gcs" {
    bucket = "ph-terraform-state"
    prefix = "main"
  }
}


provider "google" {
  credentials = file("../.sa.creds.json")
  project     = "phonehome-339613"
  region      = "europe-west1"
}

resource "google_sql_database_instance" "instance" {
  name             = "phonehome-db"
  region           = "europe-west1"
  database_version = "POSTGRES_14"
  settings {
    tier = "db-f1-micro"
    backup_configuration {
      enabled    = true
      start_time = "00:00"
      backup_retention_settings {
        retained_backups = 30
      }
    }
  }

  deletion_protection = "true"
}

resource "google_sql_database" "dwh_main" {
  name     = "phonehome"
  instance = google_sql_database_instance.instance.name
}

data "google_secret_manager_secret_version" "pg_password" {
  secret = var.pg_main_pass_secret
}

resource "google_sql_user" "user" {
  name     = var.pg_main_user
  password = data.google_secret_manager_secret_version.pg_password.secret_data
  instance = resource.google_sql_database_instance.instance.name
}
