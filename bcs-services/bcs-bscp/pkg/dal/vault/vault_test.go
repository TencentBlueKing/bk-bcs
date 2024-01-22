package vault

import (
	"os"
	"testing"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

func initClient(t testing.TB) Set {
	vaultToken := os.Getenv("VAULT_TOKEN")
	vaultAddr := os.Getenv("VAULT_ADDR")
	if vaultAddr == "" || vaultToken == "" {
		t.Skipf("VAULT_ADDR or VAULT_TOKEN env is missing")
	}

	s, err := NewSet(cc.Vault{
		Address: vaultAddr,
		Token:   vaultToken,
	})
	if err != nil {
		t.Fatalf("new set err: %s", err)
	}

	return s
}

func TestGetKv(t *testing.T) {
	s := initClient(t)
	kt := kit.New()

	opt := &types.GetLastKvOpt{
		BizID: 2,
		AppID: 1,
		Key:   "conf",
	}

	_, _, err := s.GetLastKv(kt, opt)
	if err != nil {
		t.Fatalf("GetLastKv err: %s", err)
	}
}

func BenchmarkGetKv(b *testing.B) {
	s := initClient(b)
	kt := kit.New()

	opt := &types.GetLastKvOpt{
		BizID: 2,
		AppID: 1,
		Key:   "conf",
	}

	for i := 0; i < b.N; i++ {
		_, _, err := s.GetLastKv(kt, opt)
		if err != nil {
			b.Fatalf("GetLastKv err: %s", err)
		}
	}
}

func BenchmarkParallelGetKv(b *testing.B) {
	s := initClient(b)
	kt := kit.New()

	opt := &types.GetLastKvOpt{
		BizID: 2,
		AppID: 1,
		Key:   "conf",
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _, err := s.GetLastKv(kt, opt)
			if err != nil {
				b.Fatalf("GetLastKv err: %s", err)
			}
		}
	})
}
