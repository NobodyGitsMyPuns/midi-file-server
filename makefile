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

# Define variables for Google Cloud Secret Manager
SECRET_NAME := my-service-account-secret
SECRET_DATA := '{"client_email":"midi-server-admin@gothic-oven-433521-e1.iam.gserviceaccount.com"}'
GCP_PROJECT := gothic-oven-433521-e1

.PHONY: lint
lint:
	@echo "Running linter..."
	$(GOLINT) run ./...

.PHONY: build
build:
	$(GOBUILD) .

.PHONY: clean
clean:
	@echo "Cleaning up Kubernetes resources except for database storage..."
	-kubectl delete -f $(K8S_DIR)/midi-file-server-deployment.yaml --ignore-not-found --wait=false || true
	-kubectl delete -f $(K8S_DIR)/midi-file-server-service.yaml --ignore-not-found --wait=false || true
	# Note: Do not delete PV or PVC here to avoid data loss.

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
	gcloud container clusters get-credentials $(GKE_CLUSTER_NAME) --zone $(GKE_ZONE) --project=$(GKE_PROJECT)
	kubectl apply -f $(K8S_DIR)/mongodb-pv.yaml 
	kubectl apply -f $(K8S_DIR)/mongodb-pvc.yaml
	kubectl apply -f $(K8S_DIR)/mongodb-deployment.yaml 
	kubectl apply -f $(K8S_DIR)/mongodb-service.yaml 

.PHONY: deploy-app
deploy-app: create-gcp-secret
	@echo "Deploying application to GCP..."
	gcloud container clusters get-credentials $(GKE_CLUSTER_NAME) --zone $(GKE_ZONE) --project $(GKE_PROJECT)
	kubectl apply -f $(K8S_DIR)/midi-file-server-deployment.yaml
	kubectl apply -f $(K8S_DIR)/midi-file-server-service.yaml

# Safe redeploy without touching database
.PHONY: redeploy-service
redeploy-service: build-docker push-docker deploy-app
	@echo "Service redeployed to GCP without affecting the database!"

.PHONY: all
all: clean get build test lint build-docker push-docker deploy-mongo deploy-app
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
gke-deploy: docker-build create-gcp-secret
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

# New targets for rebuild and deployment
.PHONY: local-rebuild
local-rebuild: clean get build test lint minikube-deploy
	@echo "Local rebuild and deployment to Minikube complete!"

.PHONY: full-rebuild-deploy
full-rebuild-deploy: clean get build test lint build-docker push-docker deploy-mongo deploy-app
	@echo "Full rebuild and deployment to GCP complete!"

.PHONY: full-rebuild-deploy-secret
full-rebuild-deploy-secret: create-secret clean get build test lint build-docker push-docker deploy-mongo deploy-app
	@echo "Full rebuild and deployment to GCP with secret creation complete!"

.PHONY: create-secret
create-secret:
	@echo "Creating or updating Kubernetes secret..."
	kubectl create secret generic gcr-secret --from-file=gothic_key.json=/Users/jesselopez/Documents/repos/midi-file-server/gothic_key.json --dry-run=client -o yaml | kubectl apply -f -
	@echo "Secret created or updated successfully."
	@echo "Creating or updating Kubernetes secret..."
	kubectl create secret generic signer-secret --from-file=signer.json=/Users/jesselopez/Documents/repos/midi-file-server/signer.json --dry-run=client -o yaml | kubectl apply -f -
	@echo "Secret created or updated successfully."

# Create Google Cloud Secret Manager Secret
.PHONY: create-gcp-secret
create-gcp-secret:
	@echo "Creating or updating Google Cloud Secret Manager secret..."
	gcloud secrets create $(SECRET_NAME) --replication-policy="automatic" --project=$(GCP_PROJECT) || true
	echo -n $(SECRET_DATA) | gcloud secrets versions add $(SECRET_NAME) --data-file=- --project=$(GCP_PROJECT)
	@echo "Google Cloud Secret Manager secret created or updated successfully."
