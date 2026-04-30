APP_NAME=pethealth
DEPLOYMENT=pethealth-deployment
PORT=8080

.PHONY: help deploy restart logs status pf-app pf-prom metrics

help: 
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'



deploy: ## Apply Kubernetes manifests (App + Prometheus)
	kubectl apply -f k8s/config.yaml
	kubectl apply -f k8s/deployment.yaml
	kubectl apply -f k8s/prometheus.yaml

restart: ## Restart pods to pull new images
	kubectl rollout restart deployment $(DEPLOYMENT)

status: ## Check pods, services and deployments
	kubectl get pods,svc,deploy

logs: ## Follow application logs
	kubectl logs -f deployment/$(DEPLOYMENT)

## --- Monitoring & Network ---

pf-app: ## Port-forward application to localhost:8080
	kubectl port-forward deployment/$(DEPLOYMENT) $(PORT):$(PORT)

pf-prom: ## Port-forward Prometheus UI to localhost:9090
	kubectl port-forward deployment/prometheus 9090:9090

metrics: ## Check raw metrics via curl
	curl http://localhost:$(PORT)/metrics

## --- Infrastructure ---

infra-up: ## Start PostgreSQL Shards and Kafka (Docker Compose)
	docker-compose up -d

infra-down: ## Stop local infrastructure
	docker-compose down