# Kustomize Secret Generator Plugin for Vault

This repo has two components: a Kustomize secret generator plugin for Vault and
a Dockerfile that exposes a version of kustomize that includes the plugin.

## Kustomize Secret Generator Go plugin

This Go plugin allows [Kustomize](https://kustomize.io) to generate Kubernetes
Secret manifests that contain secrets from [HashiCorp
Vault](https://vaultproject.io). See the Kustomize [Generating Secrets
docs](https://github.com/kubernetes-sigs/kustomize/blob/7971ac1/examples/kvSourceGoPlugin.md)
for more information about the mechanics.

## Kustomize executable packaged with the plugin

The Dockerfile exposes a version of kustomize that includes the Vault plugin.

Usage:

```
docker run -it --rm \
  -v (pwd):/working \
  -e "VAULT_ADDR=XXX" -e "VAULT_TOKEN=XXX" \
   kustomize build .
```
