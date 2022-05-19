# Secret

> Secret is an object that contains a small amount of sensitive information such as a password, token, or secret key, similar to a ConfigMap but designed to hold confidential data.
> Secret information may be placed in the Pod specification or in the image, so there is no need to include secret data in the application code.

## Create Secret

There are several options to create a Secret:

- [Create Secret using kubectl command](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kubectl/)
- [Create Secret using config file](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-config-file/)
- [Using kustomize to create Secret](https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kustomize/)

### Secret Type

When creating a Secret, you can set the type for it using the `type` field of the Secret resource, or its equivalent `kubectl` command-line argument (if any). The `type` of Secret facilitates programmatic handling of different types of confidential data.

Kubernetes provides several built-in types for some common usage scenarios. The legality checks that Kubernetes performs and the restrictions it imposes on these types vary.

| Built-in types                      | Usage                                         |
| ----------------------------------- |-----------------------------------------------|
| Opaque                              | arbitrary user-defined data                   |
| kubernetes.io/service-account-token | ServiceAccount token                          |
| kubernetes.io/dockercfg             | serialized form of ~/.dockercfg file          |
| kubernetes.io/dockerconfigjson      | serialized form of ~/.docker/config.json file |
| kubernetes.io/basic-auth            | credentials for basic authentication          |
| kubernetes.io/ssh-auth              | credentials for SSH authentication            |
| kubernetes.io/tls                   | data for a TLS client or server               |
| bootstrap.kubernetes.io/token       | bootstrap token data                          |


You can also define and use your own Secret type by setting a non-empty string value for the `type` field of the Secret object. If the `type` value is an empty string, it is treated as an `Opaque` type. Kubernetes does not place any restrictions on the names of types. However, if you want to use one of the built-in types, you must meet all the requirements defined for that type.

## use Secret

To use the Secret, the Pod needs to reference the Secret. Pods can use Secrets in one of three ways:

- As [files in volume](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod) mounted on one or more containers 
- As [container environment variable](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-environment-variables)
- By the [kubelet when pulling images]((https://kubernetes.io/docs/concepts/configuration/secret/#using-imagepullsecrets)) for the Pod.

The Kubernetes control plane also uses Secrets; for example, bootstrap token Secrets are a mechanism to help automate node registration.

The name of the Secret object must be a valid DNS subdomain. When writing a configuration file for creating a Secret, you can set the data and/or stringData fields. Both the data and stringData fields are optional. All key values in the data field must be base64 encoded strings. If you don't want to perform this kind of base64 string conversion, you can optionally set the stringData field, which can use any string as its value.

## References

1. [Kubernetes / Configuration / Secret](https://kubernetes.io/docs/concepts/configuration/secret/)
2. [Kubernetes Secret field description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#secret-v1-core)