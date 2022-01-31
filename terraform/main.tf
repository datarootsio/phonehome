terraform {
  backend "gcs" {
    bucket = "ph-terraform-state"
    prefix = "main"
  }
}


provider "google" {
  project     = var.gcp_project_id
  region      = "europe-west1"
}

provider "google-beta" {
  project     = var.gcp_project_id
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


resource "google_artifact_registry_repository" "repo_server" {
  provider = google-beta

  location = "europe-west1"
  repository_id = "core"
  description = "core ph repo"
  format = "DOCKER"
}


/* CLOUD RUN */

resource "google_cloud_run_service" "cloudrun_server" {
  name     = "server"
  location = "europe-west1"

  template {
    spec {
      containers {
        image = "europe-west1-docker.pkg.dev/phonehome-339613/core/server:${var.current_version}"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}


resource "google_cloud_run_service" "cloudrun_ui" {
  name     = "ui"
  location = "europe-west1"

  template {
    spec {
      containers {
        image = "europe-west1-docker.pkg.dev/phonehome-339613/core/ui:${var.current_version}"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}