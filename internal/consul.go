package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func parseConsulToken(file string) (string, error) {
	content, err := os.ReadFile(filepath.Clean(file))
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	text := string(content)
	temp := strings.Split(text, "\n")
	for _, line := range temp {
		if strings.HasPrefix(line, "SecretID:") {
			return strings.ReplaceAll(strings.ReplaceAll(line, "SecretID:", ""), " ", ""), nil
		}
	}
	return "", nil
}

func regenerateConsulPolicies(config *Config, consul Consul, inventory *Inventory) error {
	err := makeConsulPolicies(config, *inventory)
	if err != nil {
		return err
	}
	fmt.Println("Updating consul policies")
	policyConsul := filepath.Join(config.BaseDir, "consul", "consul-policies.hcl")

	return consul.UpdatePolicy("consul-policies", policyConsul)
}

func BootstrapConsul(config *Config, inventory *Inventory, consul Consul) (bool, error) {
	secrets, err := getSecrets(config.BaseDir)
	if err != nil {
		return false, err
	}

	if secrets.ConsulBootstrapToken != "TBD" {
		err = regenerateConsulPolicies(config, consul, inventory)
		return false, err
	}
	token, err := consul.Bootstrap()
	if secrets.ConsulBootstrapToken != "TBD" {
		return false, err
	}

	secrets.ConsulBootstrapToken = token

	policies := map[string]string{
		"consul-policies":    filepath.Join(config.BaseDir, "consul", "consul-policies.hcl"),
		"nomad-client":       filepath.Join(config.BaseDir, "consul", "nomad-client-policy.hcl"),
		"fabio":              filepath.Join(config.BaseDir, "consul", "fabio-policy.hcl"),
		"nomad-server":       filepath.Join(config.BaseDir, "consul", "nomad-server-policy.hcl"),
		"prometheus":         filepath.Join(config.BaseDir, "consul", "prometheus-policy.hcl"),
		"anonymous-dns-read": filepath.Join(config.BaseDir, "consul", "anonymous-policy.hcl"),
		"vault":              filepath.Join(config.BaseDir, "consul", "vault-policy.hcl"),
	}

	for k, v := range policies {
		err = consul.RegisterPolicy(k, v)
		if err != nil {
			return false, err
		}
	}

	err = consul.UpdateACL("anonymous", "anonymous-dns-read")
	if err != nil {
		return false, err
	}

	acls := map[string]string{
		"agent token":        "consul-policies",
		"client token":       "nomad-client",
		"nomad server token": "nomad-server",
		"prometheus token":   "prometheus",
		"vault token":        "vault",
		"fabio token":        "fabio",
	}
	tokens := map[string]string{}

	for k, v := range acls {
		clientToken, e := consul.RegisterACL(k, v)
		if e != nil {
			return false, e
		}
		tokens[v] = clientToken
	}

	secrets.ConsulAgentToken = tokens["consul-policies"]
	secrets.NomadClientConsulToken = tokens["nomad-client"]
	secrets.NomadServerConsulToken = tokens["nomad-server"]
	secrets.PrometheusConsulToken = tokens["prometheus"]
	secrets.FabioConsulToken = tokens["fabio"]
	secrets.VaultConsulToken = tokens["vault"]

	err = writeSecrets(config.BaseDir, secrets)

	return true, err
}
