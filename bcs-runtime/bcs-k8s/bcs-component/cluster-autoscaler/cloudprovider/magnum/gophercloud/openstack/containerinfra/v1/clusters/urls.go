package clusters

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud"
)

var apiName = "clusters"

func commonURL(client *gophercloud.ServiceClient) string {
	return client.ServiceURL(apiName)
}

func idURL(client *gophercloud.ServiceClient, id string) string {
	return client.ServiceURL(apiName, id)
}

func createURL(client *gophercloud.ServiceClient) string {
	return commonURL(client)
}

func deleteURL(client *gophercloud.ServiceClient, id string) string {
	return idURL(client, id)
}

func getURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL("clusters", id)
}

func listURL(client *gophercloud.ServiceClient) string {
	return client.ServiceURL("clusters")
}

func updateURL(client *gophercloud.ServiceClient, id string) string {
	return idURL(client, id)
}
