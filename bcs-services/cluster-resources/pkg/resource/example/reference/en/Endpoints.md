# Endpoints

> Endpoints (EPs) is a bridge between Service and Pod

## What are Endpoints?

Service and Pod are not directly connected. There is a resource between the two, which is the Endpoints resource.

The Endpoints resource is a list of IP addresses and ports that expose the service. Although we define the Pod to select it in the Service's spec, it is not used directly when redirecting incoming connections. Instead, selectors are used to build a list of IPs and ports, which are then stored in the Endpoints resource. When a client connects to a service, the service proxy chooses one of these IP and port pairs based on policy and redirects incoming connections to the server listening at that location.

## Using Endpoints

**Generally, we do not need to manually manage Endpoints resources**, it should be automatically generated, managed and maintained by the cluster according to the Service definition. However, if [Service does not define selectors](https://kubernetes.io/docs/concepts/services-networking/service/#services-without-selectors), Endpoints will not be created automatically, we will It needs to be created and updated manually.

The following is an example of manually managing Endpoints resources

1. To create a service without selectors, we define a service without selectors that will receive incoming connections on port 80.

````yaml
apiVersion: v1
kind: Service
metadata:
  name: external-service
spec:
  ports:
    - port: 80
````

2. Create Endpoints resource for service without selector

````yaml
apiVersion: v1
kind: Endpoints
metadata:
  name: external-service
subsets:
  - addresses:
      - ip: 1.1.1.1 # The IP address where the service redirects connections to Endpoints
      - ip: 2.2.2.2
    ports:
      - port: 80 # Destination port of Endpoint
````

It is important to note that the Endpoints object needs to have the same name as the service and contains a list of destination IP addresses and ports for the service. Once both the service and the Endpoints resource are published to the server, the service is available as normal with a pod selector. The container created after the service is created will contain the service's environment variables and all connections to its IP:port pair will be load balanced across the service endpoints.

> Manually created Endpoints will not be automatically maintained. If an event such as Pod rescheduling occurs, the IP configured in Endpoints may become invalid, which may cause the service to become unavailable.

## References

1. [Kubernetes Endpoints field description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#endpoints-v1-core)