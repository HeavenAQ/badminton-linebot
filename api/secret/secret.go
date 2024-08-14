package secret

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

func AccessSecretVersion(name string) ([]byte, error) {
	// create secret manager client
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	// access secret version
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to access secret version: %v", err)
	}

	// return secret data
	return result.Payload.Data, nil
}

func GetSecretNameString(secretID string) string {
	projectID := os.Getenv("GCP_PROJECT_ID")
	return fmt.Sprintf("projects/%s/secrets/%s/versions/latest", projectID, secretID)
}
