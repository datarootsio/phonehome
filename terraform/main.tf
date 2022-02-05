terraform {
  backend "gcs" {
    bucket = "ph-terraform-state"
    prefix = "main"
  }
}


provider "google" {
  project = var.gcp_project_id
  region  = "europe-west1"
}

provider "google-beta" {
  project = var.gcp_project_id
  region  = "europe-west1"
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

resource "google_sql_database" "db_main" {
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

  location      = "europe-west1"
  repository_id = "core"
  description   = "core ph repo"
  format        = "DOCKER"
}


/* CLOUD RUN */

resource "google_cloud_run_service" "cloudrun_server" {
  name     = "server"
  location = "europe-west1"

  template {
    spec {
      containers {
        image = "europe-west1-docker.pkg.dev/phonehome-339613/core/server:${var.current_version}"
        ports {
          name           = "http1"
          container_port = 8888
        }
        env {
          name  = "PG_SOCKET_DIR"
          value = "/cloudsql"
        }

        env {
          name  = "CHECK_REPO_EXISTENCE"
          value = "true"
        }

        env {
          name  = "PG_INSTANCE_CONNECTION_NAME"
          value = google_sql_database_instance.instance.connection_name
        }
        env {
          name  = "PG_PORT"
          value = 5432
        }
        env {
          name  = "PG_DATABASE"
          value = google_sql_database.db_main.name
        }
        env {
          name  = "PG_USER"
          value = var.pg_main_user
        }
        env {
          name  = "PG_PASS"
          value = data.google_secret_manager_secret_version.pg_password.secret_data
        }

      }
    }
    metadata {
      annotations = {
        "run.googleapis.com/cloudsql-instances" = google_sql_database_instance.instance.connection_name
      }
    }

  }



  traffic {
    percent         = 100
    latest_revision = true
  }

}


# resource "google_cloud_run_service" "cloudrun_ui" {
#   name     = "ui"
#   location = "europe-west1"

#   template {
#     spec {
#       containers {
#         image = "europe-west1-docker.pkg.dev/phonehome-339613/core/ui:${var.current_version}"
#         ports {
#           name           = "http1"
#           container_port = 80
#         }
#       }
#     }
#   }

#   traffic {
#     percent         = 100
#     latest_revision = true
#   }
# }


data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

resource "google_cloud_run_service_iam_policy" "noauth_server" {
  location = google_cloud_run_service.cloudrun_server.location
  project  = google_cloud_run_service.cloudrun_server.project
  service  = google_cloud_run_service.cloudrun_server.name

  policy_data = data.google_iam_policy.noauth.policy_data
}

# resource "google_cloud_run_service_iam_policy" "noauth_ui" {
#   location = google_cloud_run_service.cloudrun_ui.location
#   project  = google_cloud_run_service.cloudrun_ui.project
#   service  = google_cloud_run_service.cloudrun_ui.name

#   policy_data = data.google_iam_policy.noauth.policy_data
# }


/* load balancer config server */

resource "google_compute_region_network_endpoint_group" "server_neg" {
  name                  = "server-neg"
  network_endpoint_type = "SERVERLESS"
  region                = "europe-west1"
  cloud_run {
    service = google_cloud_run_service.cloudrun_server.name
  }
}

module "lb_http_server" {
  source  = "GoogleCloudPlatform/lb-http/google//modules/serverless_negs"
  version = "~> 4.5"

  project = var.gcp_project_id
  name    = "default-lb-server"

  managed_ssl_certificate_domains = ["api.phonehome.dev"]
  ssl                             = true
  https_redirect                  = true

  backends = {
    server = {
      description             = null
      enable_cdn              = false
      custom_request_headers  = null
      custom_response_headers = null
      security_policy         = null

      log_config = {
        enable      = true
        sample_rate = 1.0
      }

      groups = [
        {
          group = google_compute_region_network_endpoint_group.server_neg.id
        }
      ]

      iap_config = {
        enable               = false
        oauth2_client_id     = null
        oauth2_client_secret = null
      }

    }
  }
}

/* dns stuff */

resource "google_dns_managed_zone" "ph_root_zone" {
  name        = "phonehome-dev"
  dns_name    = "phonehome.dev."
  description = "phonehome.devs main zone"
}

resource "google_dns_record_set" "phonehome_zone_a_record_server" {
  name = "api.${google_dns_managed_zone.ph_root_zone.dns_name}"
  type = "A"
  ttl  = "3600"

  managed_zone = google_dns_managed_zone.ph_root_zone.name

  rrdatas = [module.lb_http_server.external_ip]
}

/* static site phonehome.dev */
resource "google_storage_bucket" "website" {
  provider = google
  name     = "phonehome-website"
  location = "EU"
}

# bucket
resource "google_storage_default_object_access_control" "website_read" {
  bucket = google_storage_bucket.website.name
  role   = "READER"
  entity = "allUsers"
}

# Reserve an external IP for website
resource "google_compute_global_address" "website" {
  provider = google
  name     = "website-lb-ip"
}

# Add the IP to the DNS
resource "google_dns_record_set" "website" {
  provider     = google
  name         = google_dns_managed_zone.ph_root_zone.dns_name
  type         = "A"
  ttl          = 300
  managed_zone = google_dns_managed_zone.ph_root_zone.name
  rrdatas      = [google_compute_global_address.website.address]
}

# Add the bucket as a CDN backend
resource "google_compute_backend_bucket" "website" {
  provider    = google
  name        = "website-backend"
  description = "Contains files needed by the website"
  bucket_name = google_storage_bucket.website.name
  enable_cdn  = true
}

# https certs
resource "google_compute_managed_ssl_certificate" "website" {
  provider = google-beta
  name     = "website-cert"
  managed {
    domains = [google_dns_record_set.website.name]
  }
}

# url map
resource "google_compute_url_map" "website" {
  provider        = google
  name            = "website-url-map"
  default_service = google_compute_backend_bucket.website.self_link
}

# proxy
resource "google_compute_target_https_proxy" "website" {
  provider         = google
  name             = "website-target-proxy"
  url_map          = google_compute_url_map.website.self_link
  ssl_certificates = [google_compute_managed_ssl_certificate.website.self_link]
}

# fw to https
resource "google_compute_global_forwarding_rule" "default" {
  provider              = google
  name                  = "website-forwarding-rule"
  load_balancing_scheme = "EXTERNAL"
  ip_address            = google_compute_global_address.website.address
  ip_protocol           = "TCP"
  port_range            = "443"
  target                = google_compute_target_https_proxy.website.self_link
}

/* -- static site phonehome.dev */