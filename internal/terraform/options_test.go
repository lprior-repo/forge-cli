package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInitOptions tests Init option builders.
func TestInitOptions(t *testing.T) {
	t.Run("Upgrade option sets Upgrade field", func(t *testing.T) {
		cfg := &InitConfig{}
		Upgrade(true)(cfg)
		assert.True(t, cfg.Upgrade)
	})

	t.Run("Backend option sets Backend field", func(t *testing.T) {
		cfg := &InitConfig{}
		Backend(false)(cfg)
		assert.False(t, cfg.Backend)
	})

	t.Run("BackendConfig option adds backend config", func(t *testing.T) {
		cfg := &InitConfig{}
		BackendConfig("key=value")(cfg)
		assert.Contains(t, cfg.BackendConfig, "key=value")
	})

	t.Run("Reconfigure option sets Reconfigure field", func(t *testing.T) {
		cfg := &InitConfig{}
		Reconfigure(true)(cfg)
		assert.True(t, cfg.Reconfigure)
	})

	t.Run("applyInitOptions applies all options", func(t *testing.T) {
		cfg := applyInitOptions(
			Upgrade(true),
			BackendConfig("test=123"),
			Reconfigure(true),
		)

		assert.True(t, cfg.Upgrade)
		assert.Contains(t, cfg.BackendConfig, "test=123")
		assert.True(t, cfg.Reconfigure)
		assert.True(t, cfg.Backend, "Backend should default to true")
	})
}

// TestPlanOptions tests Plan option builders.
func TestPlanOptions(t *testing.T) {
	t.Run("PlanOut sets Out field", func(t *testing.T) {
		cfg := &PlanConfig{}
		PlanOut("plan.tfplan")(cfg)
		assert.Equal(t, "plan.tfplan", cfg.Out)
	})

	t.Run("PlanDestroy sets Destroy field", func(t *testing.T) {
		cfg := &PlanConfig{}
		PlanDestroy(true)(cfg)
		assert.True(t, cfg.Destroy)
	})

	t.Run("PlanVarFile sets VarFile field", func(t *testing.T) {
		cfg := &PlanConfig{}
		PlanVarFile("vars.tfvars")(cfg)
		assert.Equal(t, "vars.tfvars", cfg.VarFile)
	})

	t.Run("PlanVar sets Var in map", func(t *testing.T) {
		cfg := &PlanConfig{}
		PlanVar("region", "us-west-2")(cfg)
		assert.Equal(t, "us-west-2", cfg.Vars["region"])
	})

	t.Run("applyPlanOptions applies all options", func(t *testing.T) {
		cfg := applyPlanOptions(
			PlanOut("plan.out"),
			PlanDestroy(true),
			PlanVarFile("test.tfvars"),
			PlanVar("env", "prod"),
		)

		assert.Equal(t, "plan.out", cfg.Out)
		assert.True(t, cfg.Destroy)
		assert.Equal(t, "test.tfvars", cfg.VarFile)
		assert.Equal(t, "prod", cfg.Vars["env"])
	})
}

// TestApplyOptions tests Apply option builders.
func TestApplyOptions(t *testing.T) {
	t.Run("AutoApprove sets AutoApprove field", func(t *testing.T) {
		cfg := &ApplyConfig{}
		AutoApprove(true)(cfg)
		assert.True(t, cfg.AutoApprove)
	})

	t.Run("ApplyVarFile sets VarFile field", func(t *testing.T) {
		cfg := &ApplyConfig{}
		ApplyVarFile("production.tfvars")(cfg)
		assert.Equal(t, "production.tfvars", cfg.VarFile)
	})

	t.Run("ApplyVar sets Var in map", func(t *testing.T) {
		cfg := &ApplyConfig{}
		ApplyVar("key", "value")(cfg)
		assert.Equal(t, "value", cfg.Vars["key"])
	})

	t.Run("ApplyPlanFile sets PlanFile field", func(t *testing.T) {
		cfg := &ApplyConfig{}
		ApplyPlanFile("saved.tfplan")(cfg)
		assert.Equal(t, "saved.tfplan", cfg.PlanFile)
	})

	t.Run("applyApplyOptions applies all options", func(t *testing.T) {
		cfg := applyApplyOptions(
			AutoApprove(true),
			ApplyVarFile("vars.tfvars"),
			ApplyVar("env", "staging"),
			ApplyPlanFile("plan.tfplan"),
		)

		assert.True(t, cfg.AutoApprove)
		assert.Equal(t, "vars.tfvars", cfg.VarFile)
		assert.Equal(t, "staging", cfg.Vars["env"])
		assert.Equal(t, "plan.tfplan", cfg.PlanFile)
	})
}

// TestDestroyOptions tests Destroy option builders.
func TestDestroyOptions(t *testing.T) {
	t.Run("DestroyAutoApprove sets AutoApprove field", func(t *testing.T) {
		cfg := &DestroyConfig{}
		DestroyAutoApprove(true)(cfg)
		assert.True(t, cfg.AutoApprove)
	})

	t.Run("DestroyVarFile sets VarFile field", func(t *testing.T) {
		cfg := &DestroyConfig{}
		DestroyVarFile("destroy.tfvars")(cfg)
		assert.Equal(t, "destroy.tfvars", cfg.VarFile)
	})

	t.Run("applyDestroyOptions applies all options", func(t *testing.T) {
		cfg := applyDestroyOptions(
			DestroyAutoApprove(true),
			DestroyVarFile("cleanup.tfvars"),
		)

		assert.True(t, cfg.AutoApprove)
		assert.Equal(t, "cleanup.tfvars", cfg.VarFile)
	})

	t.Run("applyDestroyOptions with no options returns defaults", func(t *testing.T) {
		cfg := applyDestroyOptions()

		assert.False(t, cfg.AutoApprove, "AutoApprove should default to false")
		assert.Equal(t, "", cfg.VarFile, "VarFile should default to empty string")
	})
}

// TestOptionEdgeCases tests edge cases for all option types.
func TestOptionEdgeCases(t *testing.T) {
	t.Run("InitOptions with empty BackendConfig", func(t *testing.T) {
		cfg := applyInitOptions(BackendConfig(""))
		assert.Contains(t, cfg.BackendConfig, "")
		assert.Len(t, cfg.BackendConfig, 1)
	})

	t.Run("InitOptions with multiple BackendConfig values", func(t *testing.T) {
		cfg := applyInitOptions(
			BackendConfig("key1=value1"),
			BackendConfig("key2=value2"),
			BackendConfig("key3=value3"),
		)
		assert.Len(t, cfg.BackendConfig, 3)
		assert.Equal(t, "key1=value1", cfg.BackendConfig[0])
		assert.Equal(t, "key2=value2", cfg.BackendConfig[1])
		assert.Equal(t, "key3=value3", cfg.BackendConfig[2])
	})

	t.Run("PlanOptions with empty VarFile", func(t *testing.T) {
		cfg := applyPlanOptions(PlanVarFile(""))
		assert.Equal(t, "", cfg.VarFile)
	})

	t.Run("PlanOptions with multiple Var calls", func(t *testing.T) {
		cfg := applyPlanOptions(
			PlanVar("key1", "value1"),
			PlanVar("key2", "value2"),
			PlanVar("key1", "updated"), // Override
		)
		assert.Len(t, cfg.Vars, 2)
		assert.Equal(t, "updated", cfg.Vars["key1"])
		assert.Equal(t, "value2", cfg.Vars["key2"])
	})

	t.Run("ApplyOptions with multiple Var calls", func(t *testing.T) {
		cfg := applyApplyOptions(
			ApplyVar("key1", "value1"),
			ApplyVar("key2", "value2"),
			ApplyVar("key1", "updated"), // Override
		)
		assert.Len(t, cfg.Vars, 2)
		assert.Equal(t, "updated", cfg.Vars["key1"])
		assert.Equal(t, "value2", cfg.Vars["key2"])
	})

	t.Run("ApplyOptions with empty VarFile", func(t *testing.T) {
		cfg := applyApplyOptions(ApplyVarFile(""))
		assert.Equal(t, "", cfg.VarFile)
	})

	t.Run("ApplyOptions with empty PlanFile", func(t *testing.T) {
		cfg := applyApplyOptions(ApplyPlanFile(""))
		assert.Equal(t, "", cfg.PlanFile)
	})

	t.Run("DestroyOptions with empty VarFile", func(t *testing.T) {
		cfg := applyDestroyOptions(DestroyVarFile(""))
		assert.Equal(t, "", cfg.VarFile)
	})

	t.Run("Options can be composed in any order", func(t *testing.T) {
		cfg1 := applyInitOptions(Upgrade(true), Backend(false))
		cfg2 := applyInitOptions(Backend(false), Upgrade(true))

		assert.Equal(t, cfg1.Upgrade, cfg2.Upgrade)
		assert.Equal(t, cfg1.Backend, cfg2.Backend)
	})

	t.Run("PlanVar initializes empty map", func(t *testing.T) {
		cfg := &PlanConfig{Vars: nil}
		PlanVar("key", "value")(cfg)
		assert.NotNil(t, cfg.Vars)
		assert.Equal(t, "value", cfg.Vars["key"])
	})

	t.Run("ApplyVar initializes empty map", func(t *testing.T) {
		cfg := &ApplyConfig{Vars: nil}
		ApplyVar("key", "value")(cfg)
		assert.NotNil(t, cfg.Vars)
		assert.Equal(t, "value", cfg.Vars["key"])
	})

	t.Run("Options with false values", func(t *testing.T) {
		cfg := applyInitOptions(
			Upgrade(false),
			Backend(true), // Explicitly set to true (overrides default)
			Reconfigure(false),
		)
		assert.False(t, cfg.Upgrade)
		assert.True(t, cfg.Backend)
		assert.False(t, cfg.Reconfigure)
	})

	t.Run("PlanDestroy with false", func(t *testing.T) {
		cfg := applyPlanOptions(PlanDestroy(false))
		assert.False(t, cfg.Destroy)
	})

	t.Run("AutoApprove with false", func(t *testing.T) {
		cfg := applyApplyOptions(AutoApprove(false))
		assert.False(t, cfg.AutoApprove)
	})

	t.Run("DestroyAutoApprove with false", func(t *testing.T) {
		cfg := applyDestroyOptions(DestroyAutoApprove(false))
		assert.False(t, cfg.AutoApprove)
	})
}
