package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedError  bool
		expectedRegion string
	}{
		{
			name:           "デフォルト設定",
			args:           []string{},
			expectedError:  false,
			expectedRegion: "us-east-1",
		},
		{
			name:           "リージョン指定",
			args:           []string{"--region", "us-west-2"},
			expectedError:  false,
			expectedRegion: "us-west-2",
		},
		{
			name:          "不正なリージョン",
			args:          []string{"--region", "invalid-region"},
			expectedError: true,
		},
		{
			name:           "output形式指定",
			args:           []string{"--output", "json"},
			expectedError:  false,
			expectedRegion: "us-east-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCommand()
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// フラグの値を確認
				region, _ := cmd.PersistentFlags().GetString("region")
				assert.Equal(t, tt.expectedRegion, region)
			}
		})
	}
}

func TestRootCommandFlags(t *testing.T) {
	cmd := NewRootCommand()

	// グローバルフラグの存在確認
	assert.NotNil(t, cmd.PersistentFlags().Lookup("region"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("profile"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("output"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("config"))
}

func TestRootCommandVersion(t *testing.T) {
	cmd := NewRootCommand()
	cmd.SetArgs([]string{"--version"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestRootCommandHelp(t *testing.T) {
	cmd := NewRootCommand()

	// Use, Short, Longの設定確認
	assert.Equal(t, "phantom-ecs", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
}
