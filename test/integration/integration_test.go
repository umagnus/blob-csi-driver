/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package integration

import (
	"context"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"

	"sigs.k8s.io/blob-csi-driver/test/utils/azure"
	"sigs.k8s.io/blob-csi-driver/test/utils/credentials"

	"github.com/stretchr/testify/assert"
)

func TestIntegration(t *testing.T) {
	creds, err := credentials.CreateAzureCredentialFile()
	defer func() {
		err := credentials.DeleteAzureCredentialFile()
		assert.NoError(t, err)
	}()
	assert.NoError(t, err)
	assert.NotNil(t, creds)

	os.Setenv("AZURE_CREDENTIAL_FILE", credentials.TempAzureCredentialFilePath)

	azureClient, err := azure.GetClient(creds.Cloud, creds.SubscriptionID, creds.AADClientID, creds.TenantID, creds.AADClientSecret)
	assert.NoError(t, err)

	ctx := context.Background()
	// Create an empty resource group for integration test
	log.Printf("Creating resource group %s in %s", creds.ResourceGroup, creds.Cloud)
	_, err = azureClient.EnsureResourceGroup(ctx, creds.ResourceGroup, creds.Location, nil)
	assert.NoError(t, err)
	defer func() {
		// Only delete resource group the test created
		if strings.HasPrefix(creds.ResourceGroup, credentials.ResourceGroupPrefix) {
			log.Printf("Deleting resource group %s", creds.ResourceGroup)
			err := azureClient.DeleteResourceGroup(ctx, creds.ResourceGroup)
			assert.NoError(t, err)
		}
	}()

	// Execute the script from project root
	err = os.Chdir("../..")
	assert.NoError(t, err)
	// Change directory back to test/integration in preparation for next test
	defer func() {
		err := os.Chdir("test/integration")
		assert.NoError(t, err)
	}()

	cwd, err := os.Getwd()
	assert.NoError(t, err)
	assert.True(t, strings.HasSuffix(cwd, "blob-csi-driver"))

	// Pass in resource group name, storage account name and cloud type
	cmd := exec.Command("./test/integration/run-tests-all-clouds.sh", creds.ResourceGroup, creds.Cloud)
	cmd.Dir = cwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Integration test failed %v", err)
	}
}