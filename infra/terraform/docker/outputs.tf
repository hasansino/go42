output "cluster_name" {
  description = "The name of the Kind cluster."
  value       = kind_cluster.this.name
}

output "kubeconfig" {
  description = "Kubeconfig content to connect to the Kind cluster. Use with caution and protect this output."
  value       = kind_cluster.this.kubeconfig
  sensitive   = true
}

output "kubeconfig_path_local" {
  description = "Path to the kubeconfig file if it was written locally."
  value       = var.kubeconfig_output_path != "" ? pathexpand(var.kubeconfig_output_path) : "Kubeconfig not written to local file (kubeconfig_output_path was empty)."
}

output "cluster_endpoint" {
  description = "The internal endpoint of the Kind cluster's API server (from within Docker network)."
  value       = kind_cluster.this.endpoint
  sensitive   = true
}
