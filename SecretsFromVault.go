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

type vaultSecret struct {
	Path      string `json:"path,omitempty" yaml:"path,omitempty"`
	Key       string `json:"key,omitempty" yaml:"key,omitempty"`
	SecretKey string `json:"secretKey,omitempty" yaml:"secretKey,omitempty"`
}

type secretSpec struct {
	Secrets []vaultSecret           `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	Options *types.GeneratorOptions `json:"options,omitempty" yaml:"options,omitempty"`
}

type plugin struct {
	rf               *resmap.Factory
	ldr              ifc.Loader
	Spec             secretSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	VaultClient      *api.Client
}

const saPath = "/run/secrets/kubernetes.io/serviceaccount/token"

//nolint: golint
//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(ldr ifc.Loader, rf *resmap.Factory, c []byte) error {
	vaultAddr, ok := os.LookupEnv("VAULT_ADDR")
	if !ok {
		return errors.New("missing `VAULT_ADDR` env var: required")
	}

	config := &api.Config{
		Address: vaultAddr,
	}

	client, err := api.NewClient(config)
	if err != nil {
		return err
	}

	vaultToken, err := getVaultToken(client)
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

	for _, secret := range p.Spec.Secrets {
		value, err := p.getSecretFromVault(secret.Path, secret.Key)
		if err != nil {
			return nil, err
		}

		var key string
		if secret.SecretKey != "" {
			key = secret.SecretKey
		} else {
			key = secret.Key
		}

		entry := fmt.Sprintf("%s=%s", key, value)
		args.LiteralSources = append(args.LiteralSources, entry)
	}

	return p.rf.FromSecretArgs(p.ldr, p.Spec.Options, args)
}

func getVaultToken(client *api.Client) (string, error) {
	t, exists := os.LookupEnv("VAULT_TOKEN")
	if !exists {
		backend, exists := os.LookupEnv("VAULT_BACKEND")
		if exists {
			role, exists := os.LookupEnv("VAULT_ROLE")
			if exists == false {
				return "", errors.New("No vault role is set for backend")
			}

			jwtByte, err := ioutil.ReadFile(saPath)
			jwt := strings.TrimSpace(string(jwtByte))
			options := map[string]interface{}{
				"jwt": jwt,
				"role": role,
			}
			path := fmt.Sprintf("auth/%s/login", backend)
			secret, err := client.Logical().Write(path, options)
			if err != nil {
				return "", err
			}

			token := secret.Auth.ClientToken
			return token, nil
		}
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
