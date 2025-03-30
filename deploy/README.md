# Deployments

Install deployment related tools:
```sh
make tools
```

## Helm 

Debug helm template:
```sh
# Render all templates with specific value files
helm template helm/api-service --debug --name-template api-service -f ./helm/api-service/values.yaml
```