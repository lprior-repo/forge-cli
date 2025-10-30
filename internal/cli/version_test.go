package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewVersionCmd(t *testing.T) {
	t.Run("creates version command", func(t *testing.T) {
		cmd := NewVersionCmd()

		assert.NotNil(t, cmd)
		assert.Equal(t, "version", cmd.Use)
		assert.NotEmpty(t, cmd.Short)
	})

	t.Run("has no flags", func(t *testing.T) {
		cmd := NewVersionCmd()

		assert.False(t, cmd.Flags().HasFlags())
	})

	t.Run("executes without error", func(t *testing.T) {
		cmd := NewVersionCmd()

		err := cmd.Execute()
		assert.NoError(t, err)
	})
}
