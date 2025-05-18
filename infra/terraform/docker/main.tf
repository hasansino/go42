
locals {}

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
        dynamic "extra_port_mappings" {
          for_each = var.extra_port_mappings // Assuming extra_port_mappings applies to control-plane
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
