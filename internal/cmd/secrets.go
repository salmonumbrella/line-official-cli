package cmd

import (
	"github.com/salmonumbrella/line-official-cli/internal/secrets"
)

func openSecretsStore() (secrets.Store, error) {
	return secrets.NewKeychainStore()
}
