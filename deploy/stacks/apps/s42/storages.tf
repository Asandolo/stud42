resource "kubernetes_manifest" "rabbitmq" {
  field_manager {
    force_conflicts = true
  }

  manifest = {
    apiVersion = "rabbitmq.com/v1beta1"
    kind       = "RabbitmqCluster"
    metadata = {
      name      = "rabbitmq"
      namespace = var.namespace
    }


    spec = {
      replicas = 1
      # Custom image available at build/packages/rabbitmq used to enable the custom 
      # plugins on rabbitmq (rabbitmq_delayed_message_exchange)
      image = "ghcr.io/42atomys/s42-rabbitmq:3.10.2-management"
      imagePullSecrets = [
        {
          name = "ghcr-creds"
        }
      ]

      affinity = {
        nodeAffinity = {
          requiredDuringSchedulingIgnoredDuringExecution = {
            nodeSelectorTerms = [
              {
                matchExpressions = [
                  {
                    key      = "nodepool"
                    operator = "In"
                    values   = [local.nodepoolSelector["storages"]["nodepool"]]
                  }
                ]
              }
            ]
          }
        }
      }

      resources = {
        requests = {
          cpu    = "50m"
          memory = "300Mi"
        }

        limits = {
          memory = "2Gi"
        }
      }

      persistence = var.hasPersistentStorage ? {
        storageClassName = "csi-cinder-high-speed"
        storage          = var.namespace == "production" ? "5Gi" : "1Gi"
      } : null

      rabbitmq = {
        additionalPlugins = [
          "rabbitmq_delayed_message_exchange",
          "rabbitmq_prometheus",
          "rabbitmq_tracing"
        ]
      }
    }
  }
}

resource "kubernetes_pod_disruption_budget" "rabbitmq" {
  depends_on = [
    kubernetes_manifest.rabbitmq
  ]

  metadata {
    name      = "rabbitmq"
    namespace = var.namespace
  }

  spec {
    max_unavailable = 0
    selector {
      match_labels = {
        "app.kubernetes.io/name" = "rabbitmq"
      }
    }
  }
}

###############################
## Postgres database
###############################
resource "random_password" "postgres" {
  length  = 64
  special = true
}

module "postgres" {
  source = "../../../modules/service"
  kind   = "StatefulSet"

  appName         = "postgres"
  appVersion      = "14.1"
  name            = "postgres"
  namespace       = var.namespace
  image           = "ghcr.io/42atomys/s42-postgres:14.2-alpine3.15"
  imagePullPolicy = "IfNotPresent"

  nodeSelector = local.nodepoolSelector["storages"]

  revisionHistoryLimit = 1
  replicas             = 1
  maxUnavailable       = 0

  podSecurityContext = {
    fsGroup      = 70
    runAsGroup   = 70
    runAsNonRoot = true
  }

  containerSecurityContext = {
    runAsGroup   = 70
    runAsNonRoot = true
    runAsUser    = 70
  }

  resources = {
    requests = {
      cpu    = var.namespace == "production" ? "100m" : "10m"
      memory = var.namespace == "production" ? "1200Mi" : "50Mi"
    }
    limits = {
      memory = "2Gi"
    }
  }

  env = {
    PGDATA      = "/var/lib/postgresql/data/pgdata"
    POSTGRES_DB = "s42"
  }

  envFromSecret = {
    POSTGRES_PASSWORD = {
      key  = "POSTGRES_PASSWORD"
      name = "postgres-credentials"
    }
  }

  ports = {
    postgres = {
      containerPort = 5432
      istioProtocol = "tcp"
    }
  }

  volumeMounts = [
    {
      mountPath  = "/var/lib/postgresql/data"
      volumeName = "data"
    },
    {
      mountPath  = "/docker-entrypoint-initdb.d"
      volumeName = "init-scripts"
      readOnly   = true
    }
  ]

  volumesFromPVC = var.hasPersistentStorage ? {
    data = {
      claimName = "postgres-data"
      readOnly  = false
    }
  } : {}

  volumesFromEmptyDir = !var.hasPersistentStorage ? {
    data = {}
  } : {}

  volumesFromConfig = {
    init-scripts = {
      configMapName = "postgres-config"
    }
  }

  configMaps = {
    config = {
      data = {
        "init-database.sh" = file("${path.root}/configs/postgres/init-database.sh")
      }
      immutable = false
    }
  }

  secrets = {
    credentials = {
      data = {
        "POSTGRES_PASSWORD"         = random_password.postgres.result
        "POSTGRES_PASSWORD_ENCODED" = urlencode(random_password.postgres.result)
      }
      immutable = false
    }
  }

  persistentVolumeClaims = var.hasPersistentStorage ? {
    data = {
      accessModes      = ["ReadWriteOnce"]
      storage          = var.namespace == "production" ? "10Gi" : "1Gi"
      storageClassName = "csi-cinder-high-speed"
    }
  } : {}
}
