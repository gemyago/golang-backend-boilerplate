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
echo kubectl create secret docker-registry ghcr-registry \
  --docker-server=https://ghcr.io \
  --docker-username="$(gh auth status | grep -o "account [^ ]*" | cut -d ' ' -f 2)" \
  --docker-password="$(gh auth token)" \
  --namespace default
```

## Helm 

Debug helm template:
```sh
# Render all templates with specific value files
helm template helm/api-service --debug --name-template api-service -f ./helm/api-service/values.yaml
```