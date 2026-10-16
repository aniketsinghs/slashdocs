package blob

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
)

type AzureConfig struct {
	Container        string
	AccountName      string
	AccountKey       string
	ConnectionString string
}

type Azure struct {
	client    *azblob.Client
	container string
}

func NewAzure(ctx context.Context, c AzureConfig) (*Azure, error) {
	var client *azblob.Client
	var err error

	if c.ConnectionString != "" {
		client, err = azblob.NewClientFromConnectionString(c.ConnectionString, nil)
	} else if c.AccountName != "" && c.AccountKey != "" {
		cred, errCred := azblob.NewSharedKeyCredential(c.AccountName, c.AccountKey)
		if errCred != nil {
			return nil, fmt.Errorf("azure credentials: %w", errCred)
		}
		endpoint := fmt.Sprintf("https://%s.blob.core.windows.net/", c.AccountName)
		client, err = azblob.NewClientWithSharedKeyCredential(endpoint, cred, nil)
	} else {
		return nil, errors.New("azure storage requires either connection string or account name/key")
	}

	if err != nil {
		return nil, fmt.Errorf("azure client: %w", err)
	}

	return &Azure{client: client, container: c.Container}, nil
}

func (a *Azure) Ping(ctx context.Context) error {
	_, err := a.client.ServiceClient().NewContainerClient(a.container).GetProperties(ctx, nil)
	if err != nil {
		return fmt.Errorf("%w: azure ping %q: %v", ErrStorage, a.container, err)
	}
	return nil
}

func (a *Azure) Put(ctx context.Context, key, contentType string, body []byte) error {
	_, err := a.client.UploadBuffer(ctx, a.container, key, body, &azblob.UploadBufferOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
	})
	if err != nil {
		return fmt.Errorf("%w: azure put %q: %v", ErrStorage, key, err)
	}
	return nil
}

func (a *Azure) Get(ctx context.Context, key string) ([]byte, string, error) {
	resp, err := a.client.DownloadStream(ctx, a.container, key, nil)
	if err != nil {
		if bloberror.HasCode(err, bloberror.BlobNotFound) {
			return nil, "", ErrNotFound
		}
		return nil, "", fmt.Errorf("%w: azure get %q: %v", ErrStorage, key, err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("%w: azure read %q: %v", ErrStorage, key, err)
	}

	ct := ""
	if resp.ContentType != nil {
		ct = *resp.ContentType
	}
	return b, ct, nil
}

func (a *Azure) Delete(ctx context.Context, key string) error {
	_, err := a.client.DeleteBlob(ctx, a.container, key, nil)
	if err != nil {
		return fmt.Errorf("%w: azure delete %q: %v", ErrStorage, key, err)
	}
	return nil
}
