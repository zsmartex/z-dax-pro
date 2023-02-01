package internal

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

//go:embed templates/nomad/cfssl.json
var cfssl string

//go:embed templates/ansible/base.yml
var baseAnsible string

//go:embed templates/ansible/consul.yml
var consulAnsible string

//go:embed templates/ansible/nomad.yml
var nomadAnsible string

//go:embed templates/ansible/vault.yml
var vaultAnsible string

// calculate bootstrap expect from files
func Configure(config *Config, inventory Inventory) error {
	err := os.MkdirAll(filepath.Join(config.BaseDir), 0750)
	if err != nil {
		return err
	}

	err = makeConfigs(config, inventory)
	if err != nil {
		return err
	}

	err = Secrets(config, inventory)
	return err
}

func makeConsulPolicies(config *Config, inventory Inventory) error {
	hosts := inventory.GetAllPrivateHosts()

	return writeTemplate(config, WriteConfig{
		Folder:     "consul",
		SourceFile: "consul-policies.hcl",
		DestFile:   "consul-policies.hcl",
		Variables: map[string]any{
			"Hosts": hosts,
		},
	})
}

func makeConfigs(config *Config, inventory Inventory) error {
	hostMap := make(map[string]string)
	hosts := ""
	first := true

	for _, v := range inventory.All.Children.ConsulServers.Hosts {
		if first {
			hosts = fmt.Sprintf(`"%v"`, v.PrivateIP)
			first = false
		} else {
			hosts = hosts + `, ` + fmt.Sprintf(`"%v"`, v.PrivateIP)
		}
		hostMap[v.PrivateIP] = v.PrivateIP
	}

	toWrites := []WriteConfig{
		{
			Folder: "ansible",
			Variables: map[string]any{
				"DATACENTER": config.DC,
			},
		},
		{
			Folder:   "consul",
			Excludes: []string{"consul-server-client.hcl", "consul-server-config.hcl", "consul-policies.hcl"},
			Extends: []WriteConfig{
				{
					SourceFile: "consul-client-config.hcl",
					DestFile:   "client.j2",
				},
				{
					SourceFile: "consul-server-config.hcl",
					DestFile:   "server.j2",
					Variables: map[string]any{
						"EXPECTS_NO": fmt.Sprintf("%v", len(inventory.All.Children.ConsulServers.GetHosts())),
					},
				},
				{
					SourceFile: "consul-policies.hcl",
					DestFile:   "consul-policies.hcl",
					Variables: map[string]any{
						"Hosts": inventory.GetAllPrivateHosts(),
					},
					WriteKind: WriteKindTemplate,
				},
			},
			Variables: map[string]any{
				"DATACENTER":   config.DC,
				"JOIN_SERVERS": hosts,
			},
		},
		{
			Folder: "fabio",
		},
		{
			Folder:   "nomad",
			Excludes: []string{"nomad.service"},
			Extends: []WriteConfig{
				{
					SourceFile: "nomad.service",
					DestFile:   "nomad-server.service",
					Variables: map[string]any{
						"NOMAD_USER": "nomad",
					},
				},
				{
					SourceFile: "nomad.service",
					DestFile:   "nomad-client.service",
					Variables: map[string]any{
						"NOMAD_USER": "root",
					},
				},
			},
			Variables: map[string]any{
				"DATACENTER": config.DC,
				"EXPECTS_NO": fmt.Sprintf("%v", len(inventory.All.Children.NomadServers.GetHosts())),
			},
		},
		{
			Folder: "consul",
		},
		{
			Folder: "vault",
		},
	}

	for _, wc := range toWrites {
		err := writeTemplate(config, wc)
		if err != nil {
			return err
		}
	}

	return nil
}

func getSecrets(baseDir string) (*secretsConfig, error) {
	bytes, err := os.ReadFile(filepath.Clean(filepath.Join(baseDir, "secrets", "secrets.yml")))
	if err != nil {
		return nil, err
	}
	var secrets secretsConfig
	err = yaml.Unmarshal(bytes, &secrets)
	if err != nil {
		return nil, err
	}
	return &secrets, nil
}

func writeSecrets(baseDir string, secrets *secretsConfig) error {
	bytes, err := yaml.Marshal(secrets)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(baseDir, "secrets", "secrets.yml"), bytes, 0600)
	if err != nil {
		return err
	}
	return nil
}

type WriteKind string

var (
	WriteKindReplacement = WriteKind("replacement") // using strings.ReplaceAll
	WriteKindTemplate    = WriteKind("template")    // for go text/template
)

type WriteConfig struct {
	Folder     string
	SourceFile string
	DestFile   string
	Excludes   []string
	Extends    []WriteConfig
	Variables  map[string]any
	WriteKind  WriteKind // default replacement
}

func (w WriteConfig) IsFolder() bool {
	return w.SourceFile == "" && w.DestFile == ""
}

func writeTemplate(config *Config, wc WriteConfig) error {
	isFolder := wc.IsFolder()

	if isFolder {
		fs, err := os.ReadDir(filepath.Join("internal", "templates", wc.Folder))
		if err != nil {
			return err
		}

		for _, fi := range fs {
			if fi.IsDir() {
				writeTemplate(config, WriteConfig{
					Folder:    filepath.Join(wc.Folder, fi.Name()),
					Variables: wc.Variables,
					WriteKind: wc.WriteKind,
				})
			} else if !Contains(wc.Excludes, fi.Name()) {
				if err := writeTemplateFile(config, WriteConfig{
					Folder:     wc.Folder,
					SourceFile: fi.Name(),
					DestFile:   fi.Name(),
					Variables:  wc.Variables,
					WriteKind:  wc.WriteKind,
				}); err != nil {
					return err
				}
			}
		}

		for _, wc2 := range wc.Extends {
			if err := mergo.Map(&wc2.Variables, wc.Variables, mergo.WithOverride); err != nil {
				return err
			}

			if len(wc2.WriteKind) == 0 {
				wc2.WriteKind = wc.WriteKind
			}

			wc2.Folder = filepath.Join(wc.Folder, wc2.Folder)

			if wc2.IsFolder() {
				err := writeTemplate(config, wc2)
				if err != nil {
					return err
				}
			} else {
				if err := writeTemplateFile(config, wc2); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func writeTemplateFile(config *Config, wc WriteConfig) error {
	isFolder := wc.IsFolder()

	if isFolder {
		panic("not support for folder")
	}

	bytes, err := os.ReadFile(filepath.Join("internal", "templates", wc.Folder, wc.SourceFile))
	if err != nil {
		return err
	}

	outTmpl := string(bytes)

	data := make(map[string]interface{})
	data["Config"] = config

	for k, v := range wc.Variables {
		data[k] = v
	}

	writeKind := wc.WriteKind
	if len(writeKind) == 0 {
		writeKind = WriteKindReplacement
	}

	buf, err := GenerateTemplate(outTmpl, data, writeKind)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Join(config.BaseDir, wc.Folder), 0750); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(config.BaseDir, wc.Folder, wc.DestFile), buf, 0600); err != nil {
		return err
	}

	return nil
}

func Contains[T any](slice []T, value T) bool {
	for _, item := range slice {
		if reflect.DeepEqual(item, value) {
			return true
		}
	}
	return false
}

func Secrets(config *Config, inventory Inventory) error {
	var out bytes.Buffer
	err := runCmd("", "consul keygen", &out)
	if err != nil {
		return err
	}
	consulSecretDir := filepath.Join(config.BaseDir, "secrets", "consul")
	nomadSecretDir := filepath.Join(config.BaseDir, "secrets", "nomad")
	err = os.MkdirAll(consulSecretDir, 0750)
	if err != nil {
		return err
	}
	consulGossipKey := strings.ReplaceAll(out.String(), "\n", "")

	var out2 bytes.Buffer
	err = runCmd("", "nomad operator keygen", &out2)

	if err != nil {
		return err
	}
	nomadGossipKey := strings.ReplaceAll(out2.String(), "\n", "")
	if os.Getenv("S3_ENDPOINT") == "" || os.Getenv("S3_SECRET_KEY") == "" || os.Getenv("S3_ACCESS_KEY") == "" {
		return fmt.Errorf("s3 compatible env variables missing for storing state: please set S3_ENDPOINT, S3_SECRET_KEY & S3_ACCESS_KEY")
	}

	secrets := &secretsConfig{
		ConsulGossipKey:        consulGossipKey,
		NomadGossipKey:         nomadGossipKey,
		NomadClientConsulToken: "TBD",
		NomadServerConsulToken: "TBD",
		ConsulAgentToken:       "TBD",
		ConsulBootstrapToken:   "TBD",
		FabioConsulToken:       "TBD",
		S3Endpoint:             os.Getenv("S3_ENDPOINT"),
		S3SecretKey:            os.Getenv("S3_SECRET_KEY"),
		S3AccessKey:            os.Getenv("S3_ACCESS_KEY"),
	}

	if _, err1 := os.Stat(filepath.Join(config.BaseDir, "secrets", "secrets.yml")); errors.Is(err1, os.ErrNotExist) {
		d, e := yaml.Marshal(&secrets)
		if e != nil {
			return e
		}
		e = os.WriteFile(filepath.Join(config.BaseDir, "secrets", "secrets.yml"), d, 0600)
		if e != nil {
			return e
		}
	}

	if err != nil {
		return err
	}
	err = os.MkdirAll(nomadSecretDir, 0750)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(config.BaseDir, "secrets", "consul", "consul-agent-ca.pem")); errors.Is(err, os.ErrNotExist) {
		err = runCmd(consulSecretDir, "consul tls ca create", os.Stdout)
		if err != nil {
			return err
		}
		err = runCmd(consulSecretDir, fmt.Sprintf("consul tls cert create -server -dc %s", config.DC), os.Stdout)
		if err != nil {
			return err
		}

	}

	if _, err := os.Stat(filepath.Join(config.BaseDir, "secrets", "nomad", "cli.pem")); errors.Is(err, os.ErrNotExist) {
		err = runCmd(nomadSecretDir, "cfssl print-defaults csr | cfssl gencert -initca - | cfssljson -bare nomad-ca", os.Stdout)
		if err != nil {
			return err
		}
		hosts := inventory.All.Children.NomadServers.GetHosts()
		privateHosts := inventory.All.Children.NomadServers.GetPrivateHosts()
		hostString := fmt.Sprintf("server.global.nomad,%s,%s", strings.Join(hosts, ","), strings.Join(privateHosts, ","))
		fmt.Println("generating cert for hosts: " + hostString)

		err = os.WriteFile(filepath.Join(nomadSecretDir, "cfssl.json"), []byte(cfssl), 0600)
		if err != nil {
			return err
		}
		err = runCmd(nomadSecretDir, fmt.Sprintf(`echo '{}' | cfssl gencert -ca=nomad-ca.pem -ca-key=nomad-ca-key.pem -config=cfssl.json -hostname="%s" - | cfssljson -bare server`, hostString), os.Stdout)
		if err != nil {
			return err
		}

		err = runCmd(nomadSecretDir, fmt.Sprintf(`echo '{}' | cfssl gencert -ca=nomad-ca.pem -ca-key=nomad-ca-key.pem -config=cfssl.json -hostname="%s" - | cfssljson -bare client`, hostString), os.Stdout)
		if err != nil {
			return err
		}

		err = runCmd(nomadSecretDir, fmt.Sprintf(`echo '{}' | cfssl gencert -ca=nomad-ca.pem -ca-key=nomad-ca-key.pem -config=cfssl.json -hostname="%s" - | cfssljson -bare cli`, hostString), os.Stdout)
		if err != nil {
			return err
		}

	}
	return nil
}

func runCmd(dir, command string, stdOut io.Writer) error {
	cmd := exec.Command("/bin/sh", "-c", command)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = stdOut
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}
