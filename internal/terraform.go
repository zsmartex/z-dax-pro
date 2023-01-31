package internal

import (
	"context"
	"io"
	"path/filepath"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
)

func InitTf(ctx context.Context, config *Config, stdOut, stdErr io.Writer) (*tfexec.Terraform, error) {
	i := install.NewInstaller()

	version := version.Must(version.NewVersion("1.2.6"))

	execPath, err := i.Ensure(ctx, []src.Source{
		&fs.ExactVersion{
			Product: product.Terraform,
			Version: version,
		},
		&releases.ExactVersion{
			Product: product.Terraform,
			Version: version,
		},
	})
	if err != nil {
		return nil, err
	}

	ips, err := GetCloudflareIPs(ctx)
	if err != nil {
		return nil, err
	}

	allowIps := make([]string, 0)
	allowIps = append(allowIps, ips.IPV4...)
	allowIps = append(allowIps, ips.IPV6...)

	err = writeTemplate(config, WriteConfig{
		Folder: "terraform",
		Variables: map[string]any{
			"HTTPSAllowedIPs": allowIps,
		},
	})
	if err != nil {
		return nil, err
	}

	tf, err := tfexec.NewTerraform(filepath.Join(config.BaseDir, "terraform"), execPath)
	if err != nil {
		return nil, err
	}

	err = tf.Init(ctx, tfexec.Upgrade(true))
	if err != nil {
		return nil, err
	}

	tf.SetStdout(stdOut)
	tf.SetStderr(stdErr)
	return tf, nil
}
