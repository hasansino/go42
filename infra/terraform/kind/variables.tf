variable "cluster_name" {
  description = "The name of the Kind cluster."
  type        = string
  default     = "go42"
}

variable "cluster_version" {
  description = "The Kubernetes version to use for the Kind cluster nodes. Example: 'v1.28.0'."
  type        = string
  default     = "v1.33.1"
}

variable "control_plane_nodes" {
  description = "Number of control-plane nodes in the Kind cluster."
  type        = number
  default     = 1
  validation {
    condition     = var.control_plane_nodes >= 1
    error_message = "The number of control-plane nodes must be at least one."
  }
}

variable "worker_nodes" {
  description = "Number of worker nodes in the Kind cluster."
  type        = number
  default     = 2
  validation {
    condition     = var.worker_nodes >= 0
    error_message = "The number of worker nodes must be zero or positive."
  }
}

variable "kubeconfig_output_path" {
  description = "Path to write the generated .kubeconfig."
  type        = string
  default     = ".kubeconfig"
}

variable "extra_port_mappings" {
  description = "A list of extra port mappings from the host to the control-plane node(s). Each mapping is an object with container_port and host_port."
  type = list(object({
    container_port = number
    host_port      = number
  }))
  default = []
}
