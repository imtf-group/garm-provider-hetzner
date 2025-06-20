package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		errString      string
		expectedConfig *Config
	}{
		{
			name: "correct config",
			content: `
			location = "location"
			token = "token"
			`,
			errString: "",
			expectedConfig: &Config{
				Location: "location",
				Token:    "token",
			},
		},
		{
			name: "missing token",
			content: `
			location = "location"
			`,
			errString:      "missing token",
			expectedConfig: nil,
		},
		{
			name: "missing location",
			content: `
			token = "token"
			`,
			errString:      "missing location",
			expectedConfig: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile, err := os.CreateTemp("", "test.toml")
			assert.NoError(t, err, "Failed to create temp file")
			defer os.Remove(tempFile.Name())
			_, err = tempFile.Write([]byte(tt.content))
			err = tempFile.Close()
			config, err := NewConfig(tempFile.Name())
			if tt.errString == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			}
			assert.Equal(t, config, tt.expectedConfig)
		})
	}
}

func TestNewConfigInvalidFile(t *testing.T) {
	t.Run("invalid file", func(t *testing.T) {
		config, err := NewConfig("does-not-exist.toml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
		assert.Nil(t, config)
	})
}
