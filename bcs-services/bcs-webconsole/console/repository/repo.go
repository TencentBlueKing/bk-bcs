package repository

import (
	"context"
	"fmt"
	"io"
)

type Provider interface {
	UploadFile(ctx context.Context, localFile, filePath string) error
	ListFile(ctx context.Context, folderName string) ([]string, error)
	DownloadFile(context.Context, string) (io.ReadCloser, error)
}

// NewProvider init provider factory by storage type
func NewProvider(providerType string) (Provider, error) {
	switch providerType {
	case "cos":
		return newCosStorage()
	case "bkRepo":
		return newBkRepoStorage()
	default:
		return nil, fmt.Errorf("%s is not supported", providerType)
	}
}
