package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestInitOptions tests Init option builders
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

// TestPlanOptions tests Plan option builders
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

// TestApplyOptions tests Apply option builders
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

// TestDestroyOptions tests Destroy option builders
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
}
