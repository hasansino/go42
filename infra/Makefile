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