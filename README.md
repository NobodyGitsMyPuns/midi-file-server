# MIDI File Server
A GoLang-based server designed to serve MIDI files from Google Cloud Storage (GCS) by generating signed URLs. The service also includes user registration and login functionality, backed by MongoDB for user management.

## OpenApi Documentation
https://nobodygitsmypuns.github.io/midi-file-server/#/paths/~1get-signed-url/post
## Features

- **MIDI File Management**: Upload and serve MIDI files stored in GCS.
- **Signed URLs**: Securely deliver MIDI files to clients via signed URLs.
- **User Registration and Login**: Register and authenticate users with MongoDB.
- **Health Check Endpoint**: Monitor the health of the service.
- **Bucket Content Listing**: List all available MIDI files in the specified GCS bucket.

## Prerequisites

- **Google Cloud SDK**: Ensure you have the Google Cloud SDK installed and configured.
- **Docker**: Install Docker for containerization.
- **Kubernetes**: Set up a Kubernetes cluster (GKE) for deploying the service.
- **MongoDB**: Run a MongoDB instance locally or in the cloud.
- **GoLang**: Install GoLang for local development and testing.

## Setup

### Google Cloud Configuration

1. **Set your GCP project**:
    ```bash
    gcloud config set project YOUR_PROJECT_ID
    gcloud auth login
    ```

2. **Enable required services**:
    ```bash
    gcloud services enable container.googleapis.com
    gcloud services enable artifactregistry.googleapis.com
    gcloud services enable secretmanager.googleapis.com
    ```

3. **Create a GKE cluster**:
    ```bash
    gcloud container clusters create midi-cluster --zone us-central1-c
    ```

4. **Get GKE credentials**:
    ```bash
    gcloud container clusters get-credentials midi-cluster --zone us-central1-c
    ```

### MongoDB Setup

1. **Run MongoDB using Docker**:
    ```bash
    docker run -d -p 27017:27017 --name mongodb mongo:latest
    ```

2. **Connect to MongoDB**:
    ```bash
    mongosh
    use testdb
    show collections
    ```

### Building and Deploying the Service

1. **Build the Docker image**:
    ```bash
    docker build -t gcr.io/YOUR_PROJECT_ID/midi-file-server:latest .
    ```

2. **Push the Docker image to GCR**:
    ```bash
    gcloud auth configure-docker
    docker push gcr.io/YOUR_PROJECT_ID/midi-file-server:latest
    ```

3. **Deploy MongoDB to Kubernetes**:
    ```bash
    kubectl apply -f .k8/mongodb-pv.yaml
    kubectl apply -f .k8/mongodb-pvc.yaml
    kubectl apply -f .k8/mongodb-deployment.yaml
    kubectl apply -f .k8/mongodb-service.yaml
    ```

4. **Deploy the MIDI File Server to Kubernetes**:
    ```bash
    kubectl apply -f .k8/midi-file-server-deployment.yaml
    kubectl apply -f .k8/midi-file-server-service.yaml
    ```

### Kubernetes Secrets and IAM Configuration

1. **Create a Docker registry secret**:
    ```bash
    kubectl create secret docker-registry gcr-secret \
        --docker-server=gcr.io \
        --docker-username=_json_key \
        --docker-password="$(cat /path/to/your/gothic_key.json)" \
        --docker-email=your-email@example.com
    ```

2. **Set up IAM permissions**:
    ```bash
    gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
        --member="serviceAccount:midi-server-admin@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
        --role="roles/secretmanager.secretAccessor"
    ```

### Final Deployment

1. **Restart the deployment**:
    ```bash
    kubectl rollout restart deployment midi-file-server
    ```

2. **Expose the service**:
    ```bash
    kubectl expose deployment midi-file-server --type=LoadBalancer --port=8080
    ```

3. **Check the services**:
    ```bash
    kubectl get services
    ```

## API Endpoints

- **Health Check**: `GET /v1/health` - Check if the service is running.
- **User Registration**: `POST /v1/register` - Register a new user by providing a username, password, OTP, and serial number.
- **User Login**: `POST /v1/login` - Authenticate an existing user with username and password.
- **Get Signed URL**: `POST /v1/get-signed-url` - Retrieve a signed URL for specified MIDI files.
- **List Available MIDI Files**: `GET /v1/list-available-midi-files` - List all MIDI files in the GCS bucket.

## Development

### Running Locally

1. **Start MongoDB**:
    ```bash
    docker run -d -p 27017:27017 --name mongodb mongo:latest
    ```

2. **Run the server**:
    ```bash
    go run main.go
    ```

### Linting and Testing

- **Run Linter**:
    ```bash
    golangci-lint run ./...
    ```

- **Run Tests**:
    ```bash
    go test ./...
    ```

## Additional Resources

- [Google Cloud SDK Documentation](https://cloud.google.com/sdk/docs)
- [MongoDB Documentation](https://docs.mongodb.com/)
- [GoLang Documentation](https://golang.org/doc/)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
