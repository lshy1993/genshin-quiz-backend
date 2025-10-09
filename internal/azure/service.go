package azure

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	goerrors "github.com/go-errors/errors"
)

func uploadChunk(
	ctx context.Context,
	client *azblob.Client,
	containerName string,
	blobName string,
	data []byte,
) error {
	if len(data) == 0 {
		return nil
	}

	_, err := client.UploadBuffer(ctx, containerName, blobName, data, &azblob.UploadBufferOptions{})
	if err != nil {
		return goerrors.WrapPrefix(err, "failed to upload chunk", 0)
	}
	return nil
}

func uploadChunkString(
	ctx context.Context,
	client *azblob.Client,
	containerName string,
	blobName string,
	data string,
) error {
	if len(data) == 0 {
		return nil
	}
	return uploadChunk(ctx, client, containerName, blobName, []byte(data))
}

func deleteChunk(
	ctx context.Context,
	client *azblob.Client,
	containerName string,
	blobName string,
) error {
	_, err := client.DeleteBlob(ctx, containerName, blobName, nil)
	if err != nil {
		if strings.Contains(err.Error(), "BlobNotFound") ||
			strings.Contains(err.Error(), "RESPONSE 404") {
			return nil
		}
		return goerrors.WrapPrefix(err, "failed to delete chunk", 0)
	}
	return nil
}

func initClient() (*azblob.Client, error) {
	azureAccountName := os.Getenv("AZURE_STORAGE_ACCOUNT")
	azureStorageKey := os.Getenv("AZURE_STORAGE_KEY")

	cred, err := azblob.NewSharedKeyCredential(azureAccountName, azureStorageKey)
	if err != nil {
		return nil, goerrors.WrapPrefix(err, "failed to create Azure KeyCredential", 0)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", azureAccountName)
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return nil, goerrors.WrapPrefix(err, "failed to create Azure Blob client", 0)
	}

	return client, nil
}
