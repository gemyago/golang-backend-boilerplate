# Deployments

Install deployment related tools:
```sh
make tools
```

## Kubernetes

Secret to pull images from a private registry may have to be created. The secret must be created in every namespace where the registry is used.
```sh
# Generic format of the command
kubectl create secret docker-registry ghcr-registry \
  --docker-server=https://ghcr.io \
  --docker-username=<user-name> \
  --docker-password="${GITHUB_TOKEN}" \
  --namespace default

# Use below if you have gh cli configured
kubectl create secret docker-registry ghcr-registry \
  --docker-server=https://ghcr.io \
  --docker-username="$(gh auth status | grep -o "account [^ ]*" | cut -d ' ' -f 2)" \
  --docker-password="$(gh auth token)" \
  --namespace golang-backend-boilerplate
```

## Helm 

Working with template:
```sh
# Render all templates with specific value files to review the output
helm template helm/api-service --debug --name-template api-service -f ./helm/api-service/values.yaml

# Install the chart. Most often you would do it in a local scenario
# Non local scenario should usually go via CI/CD pipeline
helm install api-service helm/api-service --namespace golang-backend-boilerplate \
  -f ./helm/api-service/values.yaml \
  --create-namespace \
   --dry-run

# Or upgrade the chart
helm upgrade api-service helm/api-service --namespace golang-backend-boilerplate \
  -f ./helm/api-service/values.yaml \
  --create-namespace \
  --install \
  --dry-run

# Uninstall the chart
helm uninstall api-service --namespace golang-backend-boilerplate
```