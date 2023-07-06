package test

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"testing"
)

func TestStorageCli_QueryK8SGameDeployment(t *testing.T) {
	tlsconfig, err := ssl.ClientTslConfVerity(
		"xxx",
		"xxx",
		"xxx",
		"xxx")

	if err != nil {
		t.Errorf("ssl.ClientTslConfVerity err: %v", err)
	}

	config := &bcsapi.Config{
		Hosts:     []string{"xxx:xxx"},
		TLSConfig: tlsconfig,
		Gateway:   true,
	}

	client := bcsapi.NewClient(config)
	s := client.Storage()
	mesosNamespaces, err := s.QueryMesosNamespace("xxx")
	if err != nil {
		return
	}
	t.Logf("mesosNamespaces : %v", mesosNamespaces)

	for _, ns := range mesosNamespaces {
		t.Log(ns)
	}
}
