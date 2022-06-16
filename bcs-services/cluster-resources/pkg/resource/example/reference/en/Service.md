# Service

> A Service (SVC) is an abstract method for exposing an application running on a set of Pods as a network service

## What is Service

A Kubernetes Service defines an abstraction: a logical set of Pods, a policy by which they can be accessed, commonly referred to as a microservice. The set of Pods that a Service targets is usually determined by a selector.

As an example, consider an image processing backend that runs 3 replicas. These copies are interchangeable, i.e. frontends don't need to care which backend copy they call. However, the Pods that make up this set of backends may actually change, and front-end clients should not and need not know about it, nor do they need to keep track of the state of this set of backends.

The abstraction defined by the Service can decouple this association.

## Define Service

Service is a class of objects in Kubernetes. Like all REST objects, we can request API Server to create a new instance through the POST method.

For example, suppose there is a set of Pods that expose port 9376 and are tagged with `app=MyApp`:

````yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  selector:
    app: MyApp
  ports:
    - protocol: TCP
      port: 80
      targetPort: 9376
````

The above configuration creates a Service object named `my-service` that will proxy requests to a Pod with the label `app=MyApp` using TCP port 9376.

Kubernetes assigns the service an IP address (sometimes called a `cluster IP`), which is used by the service proxy.

A service selector's controller continuously scans for Pods that match its selector, then publishes all updates to an Endpoint object called `my-service`.

It should be noted that a Service can map a receiving port to any targetPort. By default, targetPort will be set to the same value as the port field.

Port definitions in Pods are named, and you can reference these names in the service's targetPort property. This provides a lot of flexibility for deploying and developing services.

Since many services need to expose multiple ports, Kubernetes supports multiple port definitions on service objects. Each port definition can have the same protocol or a different protocol.

## Service type

Kubernetes ServiceTypes allow you to specify the type of Service you need, the default is ClusterIP.

The values and behaviors of Type are as follows:

- ClusterIP: Expose the service through the internal IP of the cluster. When this value is selected, the service can only be accessed within the cluster.

- NodePort: Exposes services via IP and static port (NodePort) on each node. The NodePort service is routed to the automatically created ClusterIP service. The NodePort service can be accessed from outside the cluster by requesting `<node IP>:<node port>`.

- LoadBalancer: Use the cloud provider's load balancer to expose services to the outside world. External load balancers can route traffic to the automatically created NodePort and ClusterIP services.

- ExternalName: By returning the CNAME and corresponding value, the service can be mapped to the contents of the externalName field (for example, foo.bar.example.com).

> Note: You need kube-dns 1.7 and above or CoreDNS 0.0.8 and above to use the ExternalName type.

You can also use Ingress to expose your own services. Ingress is not a service type, but it acts as an entry point to the cluster. It can consolidate routing rules into one resource as it can expose multiple services under the same IP address.

## References

1. [Kubernetes / Network Services / Service](https://kubernetes.io/docs/concepts/services-networking/service/)
2. [Kubernetes Service field description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#service-v1-core)