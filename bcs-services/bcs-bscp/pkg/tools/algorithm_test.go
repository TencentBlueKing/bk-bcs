package tools

import (
	"fmt"
	"testing"
)

// TestAesDeEncrytion aes DeEncrytion test
func TestAesEnDecrytion(t *testing.T) {
	//需要16的倍数
	priKey := randStr(32)
	oriStr := randStr(32)
	fmt.Println("original: ", oriStr)
	encrypted, err := AesEncrypt([]byte(oriStr), []byte(priKey))
	if err != nil {
		t.Errorf("encrypt err: %s\n", err.Error())
	}
	fmt.Println("encryptd: ", encrypted)

	original, err := AesDecrypt(encrypted, []byte(priKey))
	if err != nil {
		t.Errorf("decrypt err: %s\n", err.Error())
	}
	fmt.Println("decryptd: ", original)
	if original != oriStr {
		t.Errorf("Decryption Error, old: %s, new: %s", oriStr, original)
	}
}

func TestEnDecryptCredential(t *testing.T) {
	priKey := randStr(32)
	algo := "aes"
	oriStr := randStr(32)
	fmt.Println("original: ", oriStr)
	encrypted, err := EncryptCredential(oriStr, priKey, algo)
	if err != nil {
		t.Errorf("encrypt err: %s\n", err.Error())
		t.Fail()
	}
	fmt.Println("encryptd: ", encrypted)

	decryptd, err := DecryptCredential(encrypted, priKey, algo)
	if err != nil {
		t.Errorf("decrypt err: %s\n", err.Error())
		t.Fail()
	}
	fmt.Println("decryptd: ", decryptd)
	if decryptd != oriStr {
		t.Errorf("Decryption Error, old: %s, new: %s", oriStr, decryptd)
		t.Fail()
	}
}
