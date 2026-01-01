package cmd

import (
	"fmt"

	"github.com/salmonumbrella/line-official-cli/internal/api"
)

func newAPIClient() (*api.Client, error) {
	accountName, err := requireAccount(&flags)
	if err != nil {
		return nil, err
	}

	store, err := openSecretsStore()
	if err != nil {
		return nil, fmt.Errorf("failed to open keyring: %w", err)
	}

	creds, err := store.Get(accountName)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials for %s: %w", accountName, err)
	}

	return api.NewClient(creds.ChannelAccessToken, flags.Debug, flags.DryRun), nil
}
