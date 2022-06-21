# Ingress

> Ingress is an API object that manages external access to services in the cluster. It can provide load balancing, SSL termination, and name-based virtual hosting. The typical access method is HTTP.

## What is an Ingress?

Ingress exposes HTTP and HTTPS routes from outside the cluster to services within the cluster. Traffic routing is controlled by rules defined on the Ingress resource.

Here is a simple example where an Ingress sends all its traffic to one Service:

![img](https://i.stack.imgur.com/qF2u2.png)

An Ingress may be configured to give Services externally-reachable URLs, load balance traffic, terminate SSL / TLS, and offer name-based virtual hosting. An [Ingress controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) is responsible for fulfilling the Ingress, usually with a load balancer, though it may also configure your edge router or additional frontends to help handle the traffic.

An Ingress does not expose arbitrary ports or protocols. Exposing services other than HTTP and HTTPS to the internet typically uses a service of type [Service.Type=NodePort](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport) or [Service.Type=LoadBalancer](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer) .

## Using Ingress

> Attention! You need to have [Ingress Controllers](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) to satisfy Ingress needs, just creating an Ingress resource by itself has no effect.
>
> You may need to deploy an Ingress controller like [ingress-nginx](https://kubernetes.github.io/ingress-nginx/deploy/) . Other types of controllers can also be selected.

An example of a simple Ingress resource is as follows:

````yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: minimal-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
    - host: "*.foo.com"
      http:
        paths:
          - pathType: Prefix
            path: /foo
            backend:
              service:
                name: svc-alpha
                port:
                  number: 80
````

> Note: Ingress often uses annotations to configure some options, depending on the Ingress controller, such as [Rewrite target annotations](https://github.com/kubernetes/ingress-nginx/blob/main/docs/examples/rewrite/README.md) . Different Ingress controllers support different annotations. Users need to check the corresponding documentation to see which annotations are supported.

Ingress [protocol](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status) provides configuration of load balancer or proxy server All the information needed, which contains a list of rules that match all incoming requests. The Ingress resource only supports rules for forwarding HTTP traffic.

### Ingress Rules

Each HTTP rule (spec.rules) contains the following information:

- optional host. In this example, host (\*.foo.com) is specified, so Host traffic to which the rule applies will be forwarded to the corresponding Service; if host is not specified, the rule applies to all inbound HTTP traffic through the specified IP address.
- A list of paths paths (eg /foo ), each path has an associated backend defined by `Service Name & Port`. Both the host and path must match the content of the incoming request before the load balancer can direct traffic to the referenced service.
- backend is a combination of Service service and port name. HTTP (and HTTPS) requests to the Ingress that match the rule's host and path will be sent to the listed backends. (Note: Ingress Backend configuration structure of different apiVersion is slightly different)

[DefaultBackend](https://kubernetes.io/docs/concepts/services-networking/ingress/#default-backend) is usually configured in the Ingress controller to serve any non-compliance with the rules request at path.

### Path types

Each path in an Ingress needs to have a corresponding path type. Paths that do not explicitly set pathType fail the validity check. There are currently three supported path types:

- `ImplementationSpecific`: For this path type, the matching method depends on [IngressClass](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class). Implementations can handle this as a separate pathType or the same as a Prefix or Exact type.
- `Exact`: Exactly match the URL path, and is case sensitive.
- `Prefix`: matches based on URL path prefixes separated by /. Matching is case-sensitive and is done element by element in the path. A path element refers to a list of tags in a path separated by the / delimiter. The request matches the path `p` if each `p` is an element prefix of the request path `p`.

> Explanation: Will not match if the last element of the path is a substring of the last element in the request path (eg: /foo/bar matches /foo/bar/baz, but not /foo/barbaz).

#### Examples

| Type   | Path                        | Request Path | Match or Not?                     |
| ------ | --------------------------- | ------------ |-----------------------------------|
| Prefix | /                           | (all paths)  | Yes                               |
| Exact  | /foo                        | /foo         | Yes                               |
| Exact  | /foo                        | /bar         | No                                |
| Exact  | /foo                        | /foo/        | No                                |
| Exact  | /foo/                       | /foo         | No                                |
| Prefix | /foo                        | /foo, /foo/  | Yes                               |
| Prefix | /foo/                       | /foo, /foo/  | Yes                               |
| Prefix | /aaa/bb                     | /aaa/bbb     | No                                |
| Prefix | /aaa/bbb                    | /aaa/bbb     | Yes                               |
| Prefix | /aaa/bbb/                   | /aaa/bbb     | Yes, trailing slashes are ignored |
| Prefix | /aaa/bbb                    | /aaa/bbb/    | Yes, matches trailing slash       |
| Prefix | /aaa/bbb                    | /aaa/bbb/ccc | Yes, matches subpaths             |
| Prefix | /aaa/bbb                    | /aaa/bbbxyz  | No, string prefix does not match  |
| Prefix | /, /aaa                     | /aaa/ccc     | Yes, matches /aaa prefix          |
| Prefix | /, /aaa, /aaa/bbb           | /aaa/bbb     | Yes, matches /aaa/bbb prefix      |
| Prefix | /, /aaa, /aaa/bbb           | /ccc         | Yes, matches / prefix             |
| Prefix | /aaa                        | /ccc         | No, use default backend           |
| Mixed  | /foo (Prefix), /foo (Exact) | /foo         | Yes, preferably Exact type        |

#### Multiple matches

In some cases, multiple paths in an Ingress will match the same request. In this case the longest matching path takes precedence. If there are still two equal matching paths, the exact path type takes precedence over the prefix path type.

### Hostname wildcards

The hostname can be an exact match (eg "foo.bar.com") or a wildcard match (eg "\*.foo.com"). Exact matching requires that the HTTP host header field exactly matches the host field value. Wildcard matching requires that the HTTP host header field is the same as the suffix part of the wildcard rule.

| host       | host header     | match or not?                                     |
| ---------- | --------------- |---------------------------------------------------|
| \*.foo.com | bar.foo.com     | Matches based on shared suffix                    |
| \*.foo.com | baz.bar.foo.com | No match, wildcard only covers a single DNS label |
| \*.foo.com | foo.com         | No match, wildcard only covers a single DNS label |


## References

1. [Kubernetes / Network Services / Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)
2. [Kubernetes Ingress field description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#ingress-v1-networking-k8s-io)