package tools

import (
	"encoding/base64"
	"fmt"
	"testing"
)

// TestDeEncrytion des DeEncrytion test
func TestDeEncrytion(t *testing.T) {
	priKey := "abcdefghijklmnopqrstuvwx"
	oriStr := "abcdefgjijklmn"
	fmt.Println("original: ", oriStr)
	b64Str, err := DesEncryptToBase([]byte(oriStr), []byte(priKey))
	if err != nil {
		t.Errorf("encrypt err: %s\n", err.Error())
	}
	fmt.Println("base64 out string: ", b64Str)

	original, err := DesDecryptFromBase([]byte(b64Str), []byte(priKey))
	if err != nil {
		t.Errorf("decrypt err: %s\n", err.Error())
	}
	fmt.Println("decrypt: ", original)
	if original != oriStr {
		t.Errorf("Decryption Error, old: %s, new: %s", oriStr, original)
	}
}

// TestAesDeEncrytion aes DeEncrytion test
func TestAesDeEncrytion(t *testing.T) {
	//需要16的倍数
	priKey := "#HvL%$o0oNNoOZnk#o2qbqCeQB1iXeIR"
	oriStr := "abcdefgjijklmn"
	fmt.Println("original: ", oriStr)
	b64Str, err := AesEncrypt([]byte(oriStr), []byte(priKey))
	if err != nil {
		t.Errorf("encrypt err: %s\n", err.Error())
	}
	fmt.Println("base64 out string: ", b64Str)

	b64Byte, _ := base64.StdEncoding.DecodeString(b64Str)
	original, err := AesDecrypt(b64Byte, []byte(priKey))
	if err != nil {
		t.Errorf("decrypt err: %s\n", err.Error())
	}
	fmt.Println("decrypt: ", original)
	if original != oriStr {
		t.Errorf("Decryption Error, old: %s, new: %s", oriStr, original)
	}
}
