package repository

import (
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/data/aztables"
	"time"
)

// TableNames defines constant names for our Azure tables
const (
	FCMTokensTable = "FCMTokens"
	MessagesTable  = "Messages"
	// Using uppercase because Azure Table names must start with a letter and can only contain alphanumeric characters
)

// NewTableClient creates a new Azure Table client with the given connection string
func NewTableClient(connectionString string) (*aztables.ServiceClient, error) {
	client, err := aztables.NewServiceClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create table client: %w", err)
	}
	return client, nil
}
