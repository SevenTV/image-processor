data "kubernetes_namespace" "app" {
  metadata {
    name = var.namespace
  }
}

resource "kubernetes_secret" "app" {
  metadata {
    name      = "image-processor"
    namespace = var.namespace
  }

  data = {
    "config.yaml" = templatefile("${path.module}/config.template.yaml", {
      rmq_uri = local.infra.rabbitmq_uri
      worker_threads = var.production ? 3 : 1
      worker_jobs = var.production ? 2 : 1
      s3_region = local.infra.region
      s3_access_key = local.infra.s3_access_key.id
      s3_secret_key = local.infra.s3_access_key.secret
    })
  }
}

resource "random_id" "jwt-secret" {
  byte_length = 64
}

resource "kubernetes_deployment" "app" {
  metadata {
    name      = "image-processor"
    namespace = data.kubernetes_namespace.app.metadata[0].name
    labels = {
      app = "image-processor"
    }
  }

  lifecycle {
    replace_triggered_by = [kubernetes_secret.app]
  }

  timeouts {
    create = "4m"
    update = "2m"
    delete = "2m"
  }

  spec {
    selector {
      match_labels = {
        app = "image-processor"
      }
    }

    replicas = 1

    template {
      metadata {
        labels = {
          app = "image-processor"
        }
      }

      spec {
        container {
          name  = "image-processor"
          image = local.image_url

          port {
            name           = "metrics"
            container_port = 9100
            protocol       = "TCP"
          }

          port {
            name           = "health"
            container_port = 9000
            protocol       = "TCP"
          }

          env {
            name = "IMAGE_PROCESSOR_K8S_POD_NAME"
            value_from {
              field_ref {
                field_path = "metadata.name"
              }
            }
          }

          resources {
            requests = {
              cpu    = var.production ? "6000m" : "1500m"
              memory = var.production ? "7Gi" : "1Gi"
            }
            limits = {
              cpu    = var.production ? "6000m" : "1500m"
              memory = var.production ? "7Gi" : "1Gi"
            }
          }

          volume_mount {
            name       = "config"
            mount_path = "/app/config.yaml"
            sub_path   = "config.yaml"
          }

          volume_mount {
            name       = "tempfs"
            mount_path = "/tempfs"
          }

          liveness_probe {
            http_get {
              port = "health"
              path = "/"
            }
            initial_delay_seconds = 20
            timeout_seconds       = 5
            period_seconds        = 10
            success_threshold     = 1
            failure_threshold     = 4
          }

          readiness_probe {
            http_get {
              port = "health"
              path = "/"
            }
            initial_delay_seconds = 5
            timeout_seconds       = 5
            period_seconds        = 10
            success_threshold     = 1
            failure_threshold     = 4
          }

          image_pull_policy = var.image_pull_policy
        }

        volume {
          name = "config"
          secret {
            secret_name = kubernetes_secret.app.metadata[0].name
          }
        }

        volume {
          name = "tempfs"
          empty_dir {
            medium = "Memory"
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "app" {
  metadata {
    name      = "image-processor"
    namespace = data.kubernetes_namespace.app.metadata[0].name
    labels = {
      app = "image-processor"
    }
  }

  spec {
    selector = {
      app = "image-processor"
    }

    port {
      name        = "metrics"
      port        = 9100
      target_port = "metrics"
    }

    port {
      name        = "health"
      port        = 9000
      target_port = "health"
    }
  }
}

resource "kubectl_manifest" "app_monitor" {
  depends_on = [kubernetes_deployment.app]

  yaml_body = <<YAML
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: image-processor
  namespace: ${data.kubernetes_namespace.app.metadata[0].name}
  labels:
    app: image-processor
spec:
  selector:
    matchLabels:
      app: image-processor
  endpoints:
    - port: metrics
      interval: 10s
      scrapeTimeout: 8s
YAML
}
