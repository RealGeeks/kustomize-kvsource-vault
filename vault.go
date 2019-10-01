package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"sigs.k8s.io/yaml"
)

type VaultSecret struct {
	Path    string
	Key     string
	Relabel string
}

type plugin struct {
	rf               *resmap.Factory
	ldr              ifc.Loader
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Secrets          []VaultSecret `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	VaultClient      *api.Client
}

//nolint: golint
//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

var database = map[string]string{
	"secret/data/prd/am1/kube0/newman-api": "SaturnV",
}

func (p *plugin) Config(ldr ifc.Loader, rf *resmap.Factory, c []byte) error {
	vaultAddr, ok := os.LookupEnv("VAULT_ADDR")
	if !ok {
		return errors.New("Missing `VAULT_ADDR` env var: required")
	}

	vaultToken, err := getToken()
	if err != nil {
		return err
	}

	config := &api.Config{
		Address: vaultAddr,
	}

	client, err := api.NewClient(config)
	if err != nil {
		return err
	}

	client.SetToken(vaultToken)

	p.rf = rf
	p.ldr = ldr
	p.VaultClient = client

	return yaml.Unmarshal(c, p)
}

func (p *plugin) Generate() (resmap.ResMap, error) {
	args := types.SecretArgs{}
	args.Name = p.Name
	args.Namespace = p.Namespace

	for _, secret := range p.Secrets {
		value, err := p.getSecretFromVault(secret.Path, secret.Key)
		if err != nil {
			return nil, err
		}

		var key string
		if secret.Relabel != "" {
			key = secret.Relabel
		} else {
			key = secret.Key
		}

		entry := fmt.Sprintf("%s=%s", key, value)
		args.LiteralSources = append(args.LiteralSources, entry)
	}

	return p.rf.FromSecretArgs(p.ldr, nil, args)
}

func getToken() (string, error) {
	t, exists := os.LookupEnv("VAULT_TOKEN")
	if !exists {
		tokenPath, exists := os.LookupEnv("VAULT_TOKEN_PATH")
		if exists == false {
			return "", errors.New("No vault token and no vault token path")
		}

		tBytes, err := ioutil.ReadFile(tokenPath)
		if err != nil {
			fmt.Println("Could not read Vault token from $VAULT_TOKEN_PATH")
			return "", err
		}

		t = strings.TrimSpace(string(tBytes))
		if len(t) == 0 {
			fmt.Println("Vault token file is empty")
			return "", err
		}
	}

	return t, nil
}

func (p *plugin) getSecretFromVault(path string, key string) (value string, err error) {
	secret, err := p.VaultClient.Logical().Read(path)
	if err != nil {
		return "", err
	}
	if secret == nil {
		return "", fmt.Errorf("the path %s was not found", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("malformed secret data: %q", secret.Data["data"])
	}
	if v, ok := data[key].(string); ok {
		return v, nil
	}

	return "", fmt.Errorf("Failed to get secret from Vault: %s:%s", path, key)
}
