package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

func LoadConsulEnv(config *Config) error {
	secrets, err := getSecrets(config.BaseDir)
	if err != nil {
		return err
	}
	inv, err := LoadInventory(filepath.Clean(filepath.Join(config.BaseDir, "inventory")))
	if err != nil {
		return err
	}
	consulServer := inv.All.Children.ConsulServers.GetHosts()[0]

	envMap := map[string]string{
		"CONSUL_HTTP_ADDR":       fmt.Sprintf("https://%s:8501", consulServer),
		"CONSUL_HTTP_TOKEN":      fmt.Sprintf("https://%s:8501", secrets.ConsulBootstrapToken),
		"CONSUL_HTTP_SSL":        "true",
		"CONSUL_HTTP_SSL_VERIFY": "false",
		"CONSUL_CLIENT_CERT":     fmt.Sprintf("%s/secrets/consul/consul-agent-ca.pem", config.BaseDir),
		"CONSUL_CLIENT_KEY":      fmt.Sprintf("%s/secrets/consul/consul-agent-ca-key.pem", config.BaseDir),
	}

	return LoadEnv(envMap)
}

func LoadVaultEnv(config *Config) error {
	inv, err := LoadInventory(filepath.Clean(filepath.Join(config.BaseDir, "inventory")))
	if err != nil {
		return err
	}
	vaultServer := inv.All.Children.VaultServers.GetHosts()[0]

	envMap := map[string]string{
		"VAULT_ADDR":        fmt.Sprintf("https://%s:8200", vaultServer),
		"VAULT_SKIP_VERIFY": "true",
	}

	return LoadEnv(envMap)
}

func LoadNomadEnv(config *Config) error {
	inv, err := LoadInventory(filepath.Clean(filepath.Join(config.BaseDir, "inventory")))
	if err != nil {
		return err
	}
	nomadServer := inv.All.Children.NomadServers.GetHosts()[0]

	envMap := map[string]string{
		"NOMAD_ADDR":        fmt.Sprintf("https://%s:4646", nomadServer),
		"NOMAD_CACERT":      fmt.Sprintf("%s/secrets/nomad/nomad-ca.pem", config.BaseDir),
		"NOMAD_CLIENT_CERT": fmt.Sprintf("%s/secrets/nomad/client.pem", config.BaseDir),
		"NOMAD_CLIENT_KEY":  fmt.Sprintf("%s/secrets/nomad/client-key.pem", config.BaseDir),
	}

	return LoadEnv(envMap)
}

func LoadEnv(envMap map[string]string) error {
	for k, v := range envMap {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}

	return nil
}
