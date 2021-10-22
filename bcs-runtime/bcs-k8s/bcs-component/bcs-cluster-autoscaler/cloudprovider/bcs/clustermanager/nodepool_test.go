package clustermanager

import (
	"reflect"
	"testing"
)

var (
	testNodeGroup = ""
	operator      = ""
	apiUrl        = ""
	token         = ""
	ips           = []string{""}
)

func TestGetPool(t *testing.T) {
	client, err := NewNodePoolClient(operator, apiUrl, token)
	if err != nil {
		t.Errorf("NewPoolClient failed. err: %v", err)
	}
	ng, err := client.GetPool(testNodeGroup)
	if err != nil {
		t.Errorf("GetPool failed. err: %v", err)
	}
	_, ok := reflect.ValueOf(ng).Interface().(*NodeGroup)
	if !ok {
		t.Errorf("GetPool returns bad values")
	}
}

func TestGetPoolConfig(t *testing.T) {
	client, err := NewNodePoolClient(operator, apiUrl, token)
	if err != nil {
		t.Errorf("NewPoolClient failed. err: %v", err)
	}
	config, err := client.GetPoolConfig(testNodeGroup)
	if err != nil {
		t.Errorf("GetPoolConfig failed. err: %v", err)
	}
	_, ok := reflect.ValueOf(config).Interface().(*AutoScalingGroup)
	if !ok {
		t.Errorf("GetPoolConfig returns bad values")
	}
}

func TestGetPoolNodeTemplate(t *testing.T) {
	client, err := NewNodePoolClient(operator, apiUrl, token)
	if err != nil {
		t.Errorf("NewPoolClient failed. err: %v", err)
	}
	template, err := client.GetPoolNodeTemplate(testNodeGroup)
	if err != nil {
		t.Errorf("GetPoolNodeTemplate failed. err: %v", err)
	}
	_, ok := reflect.ValueOf(template).Interface().(*LaunchConfiguration)
	if !ok {
		t.Errorf("GetPoolNodeTemplate returns bad values")
	}
}

func TestGetNodes(t *testing.T) {
	client, err := NewNodePoolClient(operator, apiUrl, token)
	if err != nil {
		t.Errorf("NewPoolClient failed. err: %v", err)
	}
	nodes, err := client.GetNodes(testNodeGroup)
	if err != nil {
		t.Errorf("GetNodes failed. err: %v", err)
	}
	_, ok := reflect.ValueOf(nodes).Interface().([]*Node)
	if !ok {
		t.Errorf("GetNodes returns bad values")
	}
}

func TestGetAutoScalingNodes(t *testing.T) {
	client, err := NewNodePoolClient(operator, apiUrl, token)
	if err != nil {
		t.Errorf("NewPoolClient failed. err: %v", err)
	}
	nodes, err := client.GetAutoScalingNodes(testNodeGroup)
	if err != nil {
		t.Errorf("GetAutoScalingNodes failed. err: %v", err)
	}
	_, ok := reflect.ValueOf(nodes).Interface().([]*Node)
	if !ok {
		t.Errorf("GetAutoScalingNodes returns bad values")
	}
}

func TestGetNode(t *testing.T) {
	client, err := NewNodePoolClient(operator, apiUrl, token)
	if err != nil {
		t.Errorf("NewPoolClient failed. err: %v", err)
	}
	node, err := client.GetNode(ips[0])
	if err != nil {
		t.Errorf("GetNodes failed. err: %v", err)
	}
	_, ok := reflect.ValueOf(node).Interface().(*Node)
	if !ok {
		t.Errorf("GetNodes returns bad values")
	}
}

func TestUpdateDesiredNode(t *testing.T) {
	client, err := NewNodePoolClient(operator, apiUrl, token)
	if err != nil {
		t.Errorf("NewPoolClient failed. err: %v", err)
	}
	err = client.UpdateDesiredNode(testNodeGroup, 2)
	if err != nil {
		t.Errorf("UpdateDesiredNode failed. err: %v", err)
	}
}

func TestRemoveNodes(t *testing.T) {
	client, err := NewNodePoolClient(operator, apiUrl, token)
	if err != nil {
		t.Errorf("NewPoolClient failed. err: %v", err)
	}

	err = client.RemoveNodes(testNodeGroup, ips)
	if err != nil {
		t.Errorf("RemoveNodes failed. err: %v", err)
	}
}
