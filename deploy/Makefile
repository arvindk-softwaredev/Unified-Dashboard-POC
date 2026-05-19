# Makefile lives in deploy/; repo root is one level up.
REPO_ROOT := $(abspath ..)

REGISTRY ?= quay.io/rh-ee-apalit
BACKEND_IMAGE ?= tektoncd-unified-dashboard-backend
FRONTEND_IMAGE ?= tektoncd-unified-dashboard-frontend
TAG ?= latest
# OpenShift clusters are usually amd64; set empty to build for host arch (local compose).
DOCKER_PLATFORM ?= linux/amd64

.PHONY: docker-build docker-up docker-down

docker-build:
	docker build $(if $(DOCKER_PLATFORM),--platform $(DOCKER_PLATFORM),) \
		-f $(REPO_ROOT)/deploy/docker/Dockerfile.backend \
		-t $(REGISTRY)/$(BACKEND_IMAGE):$(TAG) $(REPO_ROOT)
	docker build $(if $(DOCKER_PLATFORM),--platform $(DOCKER_PLATFORM),) \
		-f $(REPO_ROOT)/deploy/docker/Dockerfile.frontend \
		-t $(REGISTRY)/$(FRONTEND_IMAGE):$(TAG) $(REPO_ROOT)

docker-up:
	docker compose -f $(REPO_ROOT)/deploy/docker-compose.yml --project-directory $(REPO_ROOT) up --build -d

docker-down:
	docker compose -f $(REPO_ROOT)/deploy/docker-compose.yml --project-directory $(REPO_ROOT) down
