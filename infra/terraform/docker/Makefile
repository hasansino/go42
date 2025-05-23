.PHONY: help
help: Makefile
	@sed -n 's/^##//p' $< | awk 'BEGIN {FS = "|"}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

## apply | create terraform resources
# Dependencies:
#   * brew install terraform
apply:
	@terraform init && \
	terraform apply -auto-approve -target=kind_cluster.this && \
    terraform apply -auto-approve -target=helm_release.ingress_nginx_http && \
    terraform apply -auto-approve -target=helm_release.ingress_nginx_grpc && \
    terraform apply -auto-approve -target=helm_release.argocd && \
    terraform apply -auto-approve -target=kubernetes_manifest.application

## destroy | destroy terraform resources
# Dependencies:
#   * brew install terraform
destroy:
	@terraform destroy -auto-approve
