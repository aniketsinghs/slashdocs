package blob

import (
	"testing"
)

func TestAzureStoreContract(t *testing.T) {
	// Simple tests for interface implementation, actual logic requires Azure credentials
	store := &Azure{}
	_ = Store(store)
	_ = Pinger(store)
}
