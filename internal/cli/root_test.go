package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCmd(t *testing.T) {
	t.Run("creates root command", func(t *testing.T) {
		cmd := NewRootCmd()

		assert.NotNil(t, cmd)
		assert.Equal(t, "forge", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
		assert.NotEmpty(t, cmd.Long)
	})

	t.Run("has verbose flag", func(t *testing.T) {
		cmd := NewRootCmd()

		flag := cmd.PersistentFlags().Lookup("verbose")
		assert.NotNil(t, flag)
		assert.Equal(t, "false", flag.DefValue)
	})

	t.Run("has region flag", func(t *testing.T) {
		cmd := NewRootCmd()

		flag := cmd.PersistentFlags().Lookup("region")
		assert.NotNil(t, flag)
		assert.Equal(t, "", flag.DefValue)
	})

	t.Run("has all subcommands", func(t *testing.T) {
		cmd := NewRootCmd()

		expectedCommands := []string{
			"new",
			"add",
			"build",
			"deploy",
			"destroy",
			"version",
		}

		for _, cmdName := range expectedCommands {
			subCmd, _, err := cmd.Find([]string{cmdName})
			assert.NoError(t, err, "Should find %s command", cmdName)
			assert.NotNil(t, subCmd, "%s command should exist", cmdName)
		}
	})

	t.Run("executes root command without error", func(t *testing.T) {
		cmd := NewRootCmd()
		cmd.SetArgs([]string{})

		// Run should display help without error
		err := cmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("silences usage on error", func(t *testing.T) {
		cmd := NewRootCmd()

		assert.True(t, cmd.SilenceUsage, "Should silence usage on error")
		assert.True(t, cmd.SilenceErrors, "Should silence errors (handled by Execute)")
	})
}

func TestRootCmdFlags(t *testing.T) {
	t.Run("verbose flag short form works", func(t *testing.T) {
		cmd := NewRootCmd()

		flag := cmd.PersistentFlags().Lookup("verbose")
		assert.NotNil(t, flag)
		assert.Equal(t, "v", flag.Shorthand)
	})

	t.Run("region flag short form works", func(t *testing.T) {
		cmd := NewRootCmd()

		flag := cmd.PersistentFlags().Lookup("region")
		assert.NotNil(t, flag)
		assert.Equal(t, "r", flag.Shorthand)
	})
}
