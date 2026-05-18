REGISTRY ?= unified-dashboard
TAG ?= latest
NAMESPACE ?= unified-dashboard

.PHONY: docker-build docker-up docker-down k8s-apply k8s-route tekton-install tekton-run

docker-build:
	docker build -f deploy/docker/Dockerfile.backend -t $(REGISTRY)/unified-dashboard-backend:$(TAG) .
	docker build -f deploy/docker/Dockerfile.frontend -t $(REGISTRY)/unified-dashboard-frontend:$(TAG) .

docker-up:
	docker compose -f deploy/docker-compose.yml up --build -d

docker-down:
	docker compose -f deploy/docker-compose.yml down

k8s-apply:
	kubectl apply -k deploy/k8s/
	kubectl apply -f deploy/tekton/rbac.yaml

k8s-route:
	kubectl apply -f deploy/k8s/route-openshift.yaml -n $(NAMESPACE)

k8s-ingress:
	kubectl apply -f deploy/k8s/ingress.yaml -n $(NAMESPACE)

tekton-install:
	kubectl apply -f deploy/tekton/task-git-clone.yaml -n $(NAMESPACE)
	kubectl apply -f deploy/tekton/task-build-image.yaml -n $(NAMESPACE)
	kubectl apply -f deploy/tekton/pipeline.yaml -n $(NAMESPACE)

tekton-run:
	kubectl create -f deploy/tekton/pipelinerun.yaml -n $(NAMESPACE)
