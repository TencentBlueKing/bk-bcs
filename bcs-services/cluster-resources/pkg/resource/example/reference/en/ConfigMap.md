# ConfigMap

> A ConfigMap (CM) is an API object used to store non-confidential data in key-value pairs. Pods can consume ConfigMaps as environment variables, command-line arguments, or as configuration files in a volume.

## Using ConfigMaps

Common ConfigMap configuration examples are as follows (note that pod using this configmap must be in the same namespace)

````yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: game-demo
data:
  # property-like keys; each key maps to a simple value
  player_initial_lives: "3"
  ui_properties_file_name: "user-interface.properties"

  # file-like keys
  game.properties: |
    enemy.types=aliens,monsters
    player.maximum-lives=5
  user-interface.properties: |
    color.good=purple
    color.bad=yellow
    allow.textmode=true
````

There are four different ways that you can use a ConfigMap to configure a container inside a Pod:

- Inside a container command and args
- Environment variables for a container
- Add a file in read-only volume, for the application to read
- Write code to run inside the Pod that uses the Kubernetes API to read a ConfigMap

> Note: Starting from v1.19, you can add an immutable field to a ConfigMap definition to create an [Immutable ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/#configmap-immutable).

## References

1. [Kubernetes / Configuration / ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/)
2. [Kubernetes ConfigMap field description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#configmap-v1-core)