package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Init_Terraform(t *testing.T) {
	conf, err := LoadConfig(filepath.Join("testdata", "config.yaml"))
	ctx := context.Background()
	tf, err := InitTf(ctx, conf, os.Stdin, os.Stderr)
	assert.NoError(t, err)
	v, _, err := tf.Version(ctx, false)
	assert.NoError(t, err)
	assert.NotNil(t, v)
}
