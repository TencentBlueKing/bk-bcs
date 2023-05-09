package tools

import (
	"fmt"
	"testing"
)

// TestAesDeEncrytion aes DeEncrytion test
func TestAesDeEncrytion(t *testing.T) {
	//需要16的倍数
	priKey := "#HvL%$o0oNNoOZnk#o2qbqCeQB1iXeIR"
	oriStr := "abcdefgjijklmn"
	fmt.Println("original: ", oriStr)
	encrypted, err := AesEncrypt([]byte(oriStr), []byte(priKey))
	if err != nil {
		t.Errorf("encrypt err: %s\n", err.Error())
	}

	original, err := AesDecrypt(encrypted, []byte(priKey))
	if err != nil {
		t.Errorf("decrypt err: %s\n", err.Error())
	}
	fmt.Println("decrypt: ", original)
	if original != oriStr {
		t.Errorf("Decryption Error, old: %s, new: %s", oriStr, original)
	}
}
