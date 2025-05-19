.PHONY: help
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = "|"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## helm-lint | lint helm files
# Dependencies:
#   * brew install helm
helm-lint:
	@helm lint helm/app/

## helm-tpl | preview k8 manifests
# Dependencies:
#   * brew install helm
helm-tpl:
	@helm template helm/app/

## helm-install | push chart to k8
# Dependencies:
#   * brew install helm
helm-install:
	@KUBECONFIG=./terraform/docker/.kubeconfig helm install go42-app helm/app/

## helm-install | push chart to k8
# Dependencies:
#   * brew install helm
helm-remove:
	@KUBECONFIG=./terraform/docker/.kubeconfig helm uninstall go42-app