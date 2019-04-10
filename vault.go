package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
)

var database = map[string]string{
	"TREE":      "oak",
	"ROCKET":    "Saturn V",
	"FRUIT":     "strawberry",
	"VEGETABLE": "carrot",
	"SIMPSON":   "homer",
}

type plugin struct{}

var KVSource plugin

var vaultAddr = os.Getenv("VAULT_ADDR")
var token = os.Getenv("VAULT_TOKEN")

func (p plugin) Get(root string, args []string) (map[string]string, error) {
	config := &api.Config{
		Address: vaultAddr,
	}

	client, err := api.NewClient(config)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	client.SetToken(token)

	mapOfSecrets := make(map[string]string)

	for _, k := range args {
		var t = strings.Split(k, "//")
		var path = t[0]
		var key = t[1]

		secret, err := client.Logical().Read(path)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if secret == nil {
			return nil, fmt.Errorf("the path %q was not found", path)
		}

		vaultSecretData, ok := secret.Data["data"].(map[string]interface{})

		if !ok {
			return nil, fmt.Errorf("malformed ecret data: %q", secret.Data["data"])
		}

		mapOfSecrets[key], ok = vaultSecretData[key].(string)

		if !ok {
			return nil, fmt.Errorf("%q was not found at %q", key, path)
		}
	}

	return mapOfSecrets, nil
}
