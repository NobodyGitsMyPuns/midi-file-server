# Load environment variables from .env file if it exists
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

# Define variables from environment
DOCKER_IMAGE := $(DOCKER_IMAGE)
MINIKUBE_PROFILE := minikube
MINIKUBE_IMAGE := midi-file-server:local
GKE_CLUSTER_NAME := midi-cluster
GKE_ZONE := $(GKE_ZONE)
GKE_PROJECT := $(GKE_PROJECT)
MONGODB_URI := $(MONGODB_URI)
MIDI_BUCKET := $(MIDI_BUCKET)
GOOGLE_APPLICATION_CREDENTIALS := $(GOOGLE_APPLICATION_CREDENTIALS)
LOAD_BALANCER_IP := $(LOAD_BALANCER_IP)

# Name of the executable
BINARY_NAME := midi-file-server

# Define variables for Google Cloud Secret Manager
SECRET_NAME := my-service-account-secret
SECRET_DATA := '{"client_email":"midi-server-admin@gothic-oven-433521-e1.iam.gserviceaccount.com"}'
GCP_PROJECT := $(GKE_PROJECT)

# Lint
.PHONY: lint
lint:
	@echo "Running linter..."
	$(GOLINT) run ./...

# Build and Clean
.PHONY: build
build:
	$(GOBUILD) .

.PHONY: clean
clean:
	@echo "Cleaning up Kubernetes resources except for database storage..."
	-kubectl delete -f $(K8S_DIR)/midi-file-server-deployment.yaml --ignore-not-found --wait=false || true
	-kubectl delete -f $(K8S_DIR)/midi-file-server-service.yaml --ignore-not-found --wait=false || true

# Test
.PHONY: test
test:
	$(CACHECLEAN) && $(GOTEST) -v ./...

.PHONY: get
get:
	$(GOGET) -v ./...

# Minikube Setup
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

# Docker Build (Local or Production)
.PHONY: docker-build-local
docker-build-local:
	@echo "Building Docker image for local development..."
	docker build --build-arg COPY_ENV=true -t $(MINIKUBE_IMAGE) .

.PHONY: docker-build
docker-build:
	@echo "Building Docker image for production..."
	docker build --build-arg COPY_ENV=false -t $(DOCKER_IMAGE):latest .

# Push Docker image to GCP
.PHONY: push-docker
push-docker:
	@echo "Pushing Docker image to GCP..."
	docker push $(DOCKER_IMAGE):latest

# Deploy MongoDB
.PHONY: deploy-mongo
deploy-mongo:
	@echo "Deploying MongoDB to GCP..."
	gcloud container clusters get-credentials $(GKE_CLUSTER_NAME) --zone $(GKE_ZONE) --project=$(GKE_PROJECT)
	kubectl apply -f $(K8S_DIR)/mongodb-pv.yaml 
	kubectl apply -f $(K8S_DIR)/mongodb-pvc.yaml
	kubectl apply -f $(K8S_DIR)/mongodb-deployment.yaml 
	kubectl apply -f $(K8S_DIR)/mongodb-service.yaml 

# Deploy App to GCP (with secret upload and cloud build)
.PHONY: deploy-app
deploy-app: upload-secrets cloudbuild-deploy
	@echo "Deploying application to GCP..."
	gcloud container clusters get-credentials $(GKE_CLUSTER_NAME) --zone $(GKE_ZONE) --project $(GKE_PROJECT)
	kubectl apply -f $(K8S_DIR)/midi-file-server-deployment.yaml
	kubectl apply -f $(K8S_DIR)/midi-file-server-service.yaml

# Redeploy service without touching database
.PHONY: redeploy-service
redeploy-service: docker-build push-docker deploy-app
	@echo "Service redeployed to GCP without affecting the database!"

# Complete Deployment
.PHONY: all
all: clean get build test lint docker-build push-docker deploy-mongo deploy-app 
	@echo "Deployment complete!"

# Cloud Build Deployment
.PHONY: cloudbuild-deploy
cloudbuild-deploy:
	@echo "Submitting build to Cloud Build using config from $(K8S_DIR)/cloudbuild.yaml..."
	gcloud builds submit --config=$(K8S_DIR)/cloudbuild.yaml .

# Upload secrets to GCP Secret Manager
.PHONY: upload-secrets
upload-secrets:
	@echo "Uploading secrets to Google Cloud Secret Manager..."
	-gcloud secrets create my-env-secret --replication-policy="automatic" || true
	gcloud secrets versions add my-env-secret --data-file=.env
	@echo "Secrets uploaded successfully."

# Cloud Build Trigger (if you have a trigger)
.PHONY: cloudbuild-trigger
cloudbuild-trigger:
	@echo "Triggering Cloud Build deployment..."
	gcloud builds triggers run <TRIGGER_NAME> --branch=main

# GKE Clean
.PHONY: gke-clean
gke-clean:
	gcloud container clusters get-credentials $(GKE_CLUSTER_NAME) --zone $(GKE_ZONE) --project $(GKE_PROJECT)
	kubectl delete -f $(K8S_DIR)/midi-file-server-deployment.yaml
	kubectl delete -f $(K8S_DIR)/midi-file-server-service.yaml

# Secret Creation for Kubernetes and GCP Secret Manager
.PHONY: create-secret
create-secret:
	@echo "Creating or updating Kubernetes secret..."
	kubectl create secret generic gcr-secret --from-file=gothic_key.json=$(GOOGLE_APPLICATION_CREDENTIALS) --dry-run=client -o yaml | kubectl apply -f -
	kubectl create secret generic signer-secret --from-file=signer.json=$(GOOGLE_APPLICATION_CREDENTIALS) --dry-run=client -o yaml | kubectl apply -f -
	@echo "Secrets created or updated successfully."

.PHONY: create-gcp-secret
create-gcp-secret:
	@echo "Creating or updating Google Cloud Secret Manager secret..."
	gcloud secrets create $(SECRET_NAME) --replication-policy="automatic" --project=$(GCP_PROJECT) || true
	echo -n $(SECRET_DATA) | gcloud secrets versions add $(SECRET_NAME) --data-file=- --project=$(GCP_PROJECT)
	@echo "Google Cloud Secret Manager secret created or updated successfully."
