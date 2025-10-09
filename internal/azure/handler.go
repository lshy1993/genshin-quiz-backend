package azure

import (
	"context"
	"fmt"
	"os"

	goerrors "github.com/go-errors/errors"
	"github.com/google/uuid"
)

func UploadItemEvidence(
	ctx context.Context,
	storeNameID string,
	evidenceUUID uuid.UUID,
	data []byte,
) error {
	if len(data) == 0 {
		return nil
	}

	client, err := initClient()
	if err != nil {
		return err
	}

	azureContainerName := os.Getenv("AZURE_STORAGE_CONTAINER_EVIDENCE")
	err = uploadChunk(
		ctx,
		client,
		azureContainerName,
		genEvidenceBlobName(storeNameID, evidenceUUID),
		data,
	)
	if err != nil {
		return err
	}

	return nil
}

func DeleteItemEvidence(
	ctx context.Context,
	storeNameID string,
	evidenceUUID uuid.UUID,
) error {
	client, err := initClient()
	if err != nil {
		return err
	}

	azureContainerName := os.Getenv("AZURE_STORAGE_CONTAINER_EVIDENCE")
	err = deleteChunk(
		ctx,
		client,
		azureContainerName,
		genEvidenceBlobName(storeNameID, evidenceUUID),
	)
	if err != nil {
		return err
	}

	return nil
}

func UploadItemImage(
	ctx context.Context,
	storeNameID string,
	itemID uuid.UUID,
	data []byte,
) error {
	if len(data) == 0 {
		return nil
	}

	client, err := initClient()
	if err != nil {
		return goerrors.WrapPrefix(err, "failed to create Azure shared key credential", 0)
	}

	azureContainerName := os.Getenv("AZURE_STORAGE_CONTAINER_ITEMS")
	err = uploadChunk(
		ctx,
		client,
		azureContainerName,
		genItemImageBlobName(storeNameID, itemID),
		data,
	)
	if err != nil {
		return goerrors.WrapPrefix(err, "failed to upload blob to Azure", 0)
	}

	return nil
}

func UploadItemImageBase64(
	ctx context.Context,
	storeNameID string,
	itemID uuid.UUID,
	data string,
) error {
	if len(data) == 0 {
		return nil
	}

	client, err := initClient()
	if err != nil {
		return goerrors.WrapPrefix(err, "failed to create Azure shared key credential", 0)
	}

	azureContainerName := os.Getenv("AZURE_STORAGE_CONTAINER_ITEMS")
	err = uploadChunkString(
		ctx,
		client,
		azureContainerName,
		genItemImageBlobName(storeNameID, itemID),
		data,
	)
	if err != nil {
		return goerrors.WrapPrefix(err, "failed to upload blob to Azure", 0)
	}

	return nil
}

func DeleteItemImage(
	ctx context.Context,
	storeNameID string,
	itemID uuid.UUID,
) error {
	client, err := initClient()
	if err != nil {
		return err
	}

	azureContainerName := os.Getenv("AZURE_STORAGE_CONTAINER_ITEMS")
	err = deleteChunk(
		ctx,
		client,
		azureContainerName,
		genItemImageBlobName(storeNameID, itemID),
	)
	if err != nil {
		return err
	}

	return nil
}

func genEvidenceBlobName(storeNameID string, evidenceUUID uuid.UUID) string {
	// blob name is "storeName/evidence_evidenceUUID"
	return fmt.Sprintf("%s/evidence_%s", storeNameID, evidenceUUID)
}

func genItemImageBlobName(storeNameID string, itemID uuid.UUID) string {
	// blob name is "storeName/itemUUID"
	return fmt.Sprintf("%s/%s", storeNameID, itemID)
}
