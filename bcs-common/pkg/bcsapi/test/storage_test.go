package test

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"testing"
)

func TestStorageCli_QueryK8SGameDeployment(t *testing.T) {
	tlsconfig, err := ssl.ClientTslConfVerity(
		"",
		"",
		"",
		"")

	if err != nil {
		t.Errorf("ssl.ClientTslConfVerity err: %v", err)
	}

	config := &bcsapi.Config{
		Hosts:     []string{"9.143.98.44:8081"},
		TLSConfig: tlsconfig,
		Gateway:   true,
	}

	client := bcsapi.NewClient(config)
	s := client.Storage()
	mesosNamespaces, err := s.QueryMesosNamespace("BCS-MESOS-20042")
	if err != nil {
		return
	}
	t.Logf("mesosNamespaces : %v", mesosNamespaces)

	for _, ns := range mesosNamespaces {
		t.Log(ns)
	}
}
