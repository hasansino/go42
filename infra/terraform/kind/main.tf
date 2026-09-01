# ---
locals {
  config_context = "kind-${var.cluster_name}"
}
# ---

resource "kind_cluster" "this" {
  wait_for_ready  = true
  name            = var.cluster_name
  node_image      = "kindest/node:${var.cluster_version}"
  kubeconfig_path = var.kubeconfig_output_path

  kind_config {
    kind        = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"

    dynamic "node" {
      for_each = range(var.control_plane_nodes)
      content {
        role = "control-plane"
        kubeadm_config_patches = [
          "kind: InitConfiguration\nnodeRegistration:\n  kubeletExtraArgs:\n    node-labels: \"ingress-ready=yes\"\n"
        ]
        dynamic "extra_port_mappings" {
          for_each = var.extra_port_mappings
          content {
            container_port = extra_port_mappings.value.container_port
            host_port      = extra_port_mappings.value.host_port
          }
        }
      }
    }

    dynamic "node" {
      for_each = range(var.worker_nodes)
      content {
        role = "worker"
      }
    }
  }
}

provider "kubernetes" {
  config_path    = kind_cluster.this.kubeconfig_path
  config_context = local.config_context
}

provider "helm" {
  kubernetes = {
    config_path    = kind_cluster.this.kubeconfig_path
    config_context = local.config_context
  }
}

resource "helm_release" "ingress_nginx_http" {
  depends_on = [kind_cluster.this]

  name             = "ingress-nginx-http"
  repository       = "https://kubernetes.github.io/ingress-nginx"
  chart            = "ingress-nginx"
  namespace        = "ingress-nginx-http"
  create_namespace = true
  version          = "4.12.2"
  wait             = false

  set = [
    { name = "controller.ingressClassResource.name", value = "nginx-http" },
    { name = "controller.ingressClassResource.controllerValue", value = "k8s.io/ingress-nginx-http" },
    { name = "controller.ingressClassByName", value = "true" },

    { name = "controller.kind", value = "DaemonSet" },
    { name = "controller.nodeSelector.ingress-ready", value = "yes" },
    { name = "controller.tolerations[0].key", value = "node-role.kubernetes.io/control-plane" },
    { name = "controller.tolerations[0].operator", value = "Exists" },
    { name = "controller.tolerations[0].effect", value = "NoSchedule" },

    { name = "controller.service.enabled", value = "false" },
    { name = "controller.hostPort.enabled", value = "true" },
    { name = "controller.service.ports.http", value = "nginx-http" },
    { name = "controller.service.targetPorts.http", value = "nginx-http" },
    { name = "controller.metrics.service.port", value = "10255" },
    { name = "controller.admissionWebhooks.port", value = "8443" },
    { name = "controller.admissionWebhooks.service.port", value = "9445" },

    { name = "controller.containerPort.http", value = "80" },
    { name = "controller.hostPort.ports.http", value = "80" },
    { name = "controller.containerPort.https", value = "443" },
    { name = "controller.hostPort.ports.https", value = "443" },
  ]
}

resource "helm_release" "ingress_nginx_grpc" {
  depends_on = [kind_cluster.this]

  name             = "ingress-nginx-grpc"
  repository       = "https://kubernetes.github.io/ingress-nginx"
  chart            = "ingress-nginx"
  namespace        = "ingress-nginx-grpc"
  create_namespace = true
  version          = "4.12.2"
  wait             = false

  set = [
    { name = "controller.ingressClassResource.name", value = "nginx-grpc" },
    { name = "controller.ingressClassResource.controllerValue", value = "k8s.io/ingress-nginx-grpc" },
    { name = "controller.ingressClassByName", value = "true" },

    { name = "controller.kind", value = "DaemonSet" },
    { name = "controller.nodeSelector.ingress-ready", value = "yes" },
    { name = "controller.tolerations[0].key", value = "node-role.kubernetes.io/control-plane" },
    { name = "controller.tolerations[0].operator", value = "Exists" },
    { name = "controller.tolerations[0].effect", value = "NoSchedule" },

    { name = "controller.service.enabled", value = "false" },
    { name = "controller.hostPort.enabled", value = "true" },
    { name = "controller.service.ports.http", value = "nginx-grpc" },
    { name = "controller.service.targetPorts.http", value = "nginx-grpc" },
    { name = "controller.metrics.service.port", value = "10265" },
    { name = "controller.admissionWebhooks.port", value = "8453" },
    { name = "controller.admissionWebhooks.service.port", value = "9455" },

    { name = "controller.containerPort.http", value = "80" },
    { name = "controller.hostPort.ports.http", value = "50051" },
    { name = "controller.containerPort.https", value = "8443" },
    { name = "controller.hostPort.ports.https", value = "8443" },
  ]
}

resource "helm_release" "argocd" {
  depends_on = [kind_cluster.this]

  name             = "argocd"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  version          = "8.0.9"
  namespace        = "argocd"
  create_namespace = true

  values = [
    <<-EOT
    server:
      extraArgs:
        - --insecure
      service:
        type: ClusterIP
    controller:
      resources:
        limits:
          cpu: 500m
          memory: 512Mi
        requests:
          cpu: 100m
          memory: 128Mi
    repoServer:
      resources:
        limits:
          cpu: 300m
          memory: 256Mi
        requests:
          cpu: 50m
          memory: 64Mi
    dex:
      enabled: false
    redis:
      resources:
        limits:
          cpu: 200m
          memory: 128Mi
        requests:
          cpu: 50m
          memory: 64Mi
    EOT
  ]
}

resource "kubernetes_manifest" "application" {
  depends_on = [
    kind_cluster.this,
    helm_release.argocd
  ]

  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "go42-app"
      namespace = "argocd"
    }
    spec = {
      project = "default"
      source = {
        repoURL        = "https://github.com/hasansino/go42.git"
        targetRevision = "HEAD"
        path           = "infra/helm/app"
        helm = {
          valueFiles = ["values.yaml"]
        }
      }
      destination = {
        server    = "https://kubernetes.default.svc"
        namespace = "default"
      }
      syncPolicy = {
        automated = {
          prune    = true
          selfHeal = true
        }
        syncOptions = [
          "CreateNamespace=true"
        ]
      }
    }
  }
}
