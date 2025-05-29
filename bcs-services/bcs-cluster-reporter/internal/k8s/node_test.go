package k8s

import (
	"k8s.io/client-go/tools/clientcmd"
	"testing"
)

func TestNode(t *testing.T) {
	config, err := clientcmd.BuildConfigFromFlags("", "./config")
	if err != nil {
		// 处理错误
	}

	c, err := NewNodeController("xxxxxxxxxx", config, "xxxxxxxx", "xxxxxxxxxx")
	if err != nil {
		t.Error(err)
	} else {
		defer c.Close()
		c.NodeGetFile("/etc/kubernetes/*.audit", "/tmp/master-pod/")
	}

}
