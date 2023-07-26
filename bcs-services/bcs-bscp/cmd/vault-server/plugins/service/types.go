package service

type keyStorage struct {
	AppID     string
	Name      string
	Algorithm EncryptionAlgorithm
	Key       string
}

type kvStorage struct {
	AppID string
	Name  string
	Value string
}
