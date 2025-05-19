cluster_name           = "go42"
cluster_version        = "v1.29.2"
control_plane_nodes    = 1
worker_nodes           = 2
kubeconfig_output_path = ".kubeconfig"
extra_port_mappings = [
  { host_port = 8080, container_port = 8080 },
  { host_port = 50051, container_port = 50051 }
]