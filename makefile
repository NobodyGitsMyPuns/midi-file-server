# Load environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Go parameters
GOCMD := go
K8S_DIR := .k8

GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
CACHECLEAN := $(GOCMD) clean --testcache
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOLINT := golangci-lint

# Define variables
DOCKER_IMAGE := $(DOCKER_IMAGE)
MINIKUBE_PROFILE := minikube
MINIKUBE_IMAGE := midi-file-server:latest
GKE_CLUSTER_NAME := midi-cluster
GKE_ZONE := $(GKE_ZONE)
GKE_PROJECT := $(GKE_PROJECT)

# Name of the executable
BINARY_NAME := midi-file-server

.PHONY: lint
lint:
	@echo "Running linter..."
	$(GOLINT) run ./...

.PHONY: build
build:
	$(GOBUILD) .

.PHONY: clean
clean:
	@echo "Cleaning up Kubernetes resources..."
	-kubectl delete -f $(K8S_DIR)/mongodb-pv.yaml --ignore-not-found --wait=false || true
	-kubectl delete -f $(K8S_DIR)/mongodb-pvc.yaml --ignore-not-found --wait=false || true
	-kubectl delete -f $(K8S_DIR)/mongodb-deployment.yaml --ignore-not-found --wait=false || true
	-kubectl delete -f $(K8S_DIR)/mongodb-service.yaml --ignore-not-found --wait=false || true
	-kubectl delete -f $(K8S_DIR)/midi-file-server-deployment.yaml --ignore-not-found --wait=false || true
	-kubectl delete -f $(K8S_DIR)/midi-file-server-service.yaml --ignore-not-found --wait=false || true

.PHONY: test
test:
	$(CACHECLEAN) && $(GOTEST) -v ./...

.PHONY: get
get:
	$(GOGET) -v ./...

.PHONY: start-minikube
start-minikube:
	@echo "Starting Minikube..."
	minikube start --profile=$(MINIKUBE_PROFILE)
	@echo "Setting up Docker to use Minikube's Docker daemon..."
	eval $$(minikube -p $(MINIKUBE_PROFILE) docker-env)

.PHONY: stop-minikube
stop-minikube:
	@echo "Stopping Minikube..."
	minikube stop --profile=$(MINIKUBE_PROFILE)

.PHONY: build-docker
build-docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):latest .

.PHONY: push-docker
push-docker:
	@echo "Pushing Docker image to GCP..."
	docker push $(DOCKER_IMAGE):latest

.PHONY: deploy-mongo
deploy-mongo:
	@echo "Deploying MongoDB to GCP..."
	gcloud container clusters get-credentials $(GKE_CLUSTER_NAME) --zone $(GKE_ZONE) --project $(GKE_PROJECT)
	kubectl apply -f $(K8S_DIR)/mongodb-pv.yaml 
	kubectl apply -f $(K8S_DIR)/mongodb-pvc.yaml
	kubectl apply -f $(K8S_DIR)/mongodb-deployment.yaml 
	kubectl apply -f $(K8S_DIR)/mongodb-service.yaml 

.PHONY: deploy-app
deploy-app:
	@echo "Deploying application to GCP..."
	gcloud container clusters get-credentials $(GKE_CLUSTER_NAME) --zone $(GKE_ZONE) --project $(GKE_PROJECT)
	kubectl apply -f $(K8S_DIR)/midi-file-server-deployment.yaml
	kubectl apply -f $(K8S_DIR)/midi-file-server-service.yaml

.PHONY: all
all: clean get build test lint build-docker push-docker deploy-mongo deploy-app deploy-service
	@echo "Deployment complete!"

.PHONY: deploy-service
deploy-service:
	@echo "Deploying service to Minikube..."
	sed 's/${LOAD_BALANCER_IP}/$(LOAD_BALANCER_IP)/g' $(K8S_DIR)/midi-file-server-service.yaml | kubectl apply -f -

.PHONY: docker-build
docker-build:
	docker build -t $(DOCKER_IMAGE) .
	docker push $(DOCKER_IMAGE)

.PHONY: gke-deploy
gke-deploy: docker-build
	gcloud container clusters get-credentials $(GKE_CLUSTER_NAME) --zone $(GKE_ZONE) --project $(GKE_PROJECT)
	kubectl apply -f $(K8S_DIR)/midi-file-server-deployment.yaml
	sed 's/${LOAD_BALANCER_IP}/$(LOAD_BALANCER_IP)/g' $(K8S_DIR)/midi-file-server-service.yaml | kubectl apply -f -

.PHONY: minikube-deploy
minikube-deploy:
	eval $$(minikube docker-env)
	docker build -t $(MINIKUBE_IMAGE) .
	sed 's/${LOAD_BALANCER_IP}/$(LOAD_BALANCER_IP)/g' $(K8S_DIR)/midi-file-server-deployment.yaml | kubectl apply -f -
	sed 's/${LOAD_BALANCER_IP}/$(LOAD_BALANCER_IP)/g' $(K8S_DIR)/midi-file-server-service.yaml | kubectl apply -f -

.PHONY: minikube-clean
minikube-clean:
	kubectl delete -f $(K8S_DIR)/midi-file-server-deployment.yaml
	kubectl delete -f $(K8S_DIR)/midi-file-server-service.yaml

.PHONY: gke-clean
gke-clean:
	gcloud container clusters get-credentials $(GKE_CLUSTER_NAME) --zone $(GKE_ZONE) --project $(GKE_PROJECT)
	kubectl delete -f $(K8S_DIR)/midi-file-server-deployment.yaml
	kubectl delete -f $(K8S_DIR)/midi-file-server-service.yaml
