package terraform

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-exec/tfexec"
)

// makeInitFunc returns a closure that executes terraform init
func makeInitFunc(tfPath string) InitFunc {
	return func(ctx context.Context, dir string, opts ...InitOption) error {
		tf, err := tfexec.NewTerraform(dir, tfPath)
		if err != nil {
			return fmt.Errorf("failed to create terraform: %w", err)
		}

		cfg := applyInitOptions(opts...)

		var tfOpts []tfexec.InitOption
		if cfg.Upgrade {
			tfOpts = append(tfOpts, tfexec.Upgrade(true))
		}
		if !cfg.Backend {
			tfOpts = append(tfOpts, tfexec.Backend(false))
		}
		if cfg.Reconfigure {
			tfOpts = append(tfOpts, tfexec.Reconfigure(true))
		}

		return tf.Init(ctx, tfOpts...)
	}
}

// makePlanFunc returns a closure that executes terraform plan
func makePlanFunc(tfPath string) PlanFunc {
	return func(ctx context.Context, dir string, opts ...PlanOption) (bool, error) {
		tf, err := tfexec.NewTerraform(dir, tfPath)
		if err != nil {
			return false, fmt.Errorf("failed to create terraform: %w", err)
		}

		cfg := applyPlanOptions(opts...)

		var tfOpts []tfexec.PlanOption
		if cfg.Out != "" {
			tfOpts = append(tfOpts, tfexec.Out(cfg.Out))
		}
		if cfg.Destroy {
			tfOpts = append(tfOpts, tfexec.Destroy(true))
		}
		if cfg.VarFile != "" {
			tfOpts = append(tfOpts, tfexec.VarFile(cfg.VarFile))
		}

		return tf.Plan(ctx, tfOpts...)
	}
}

// makeApplyFunc returns a closure that executes terraform apply
func makeApplyFunc(tfPath string) ApplyFunc {
	return func(ctx context.Context, dir string, opts ...ApplyOption) error {
		tf, err := tfexec.NewTerraform(dir, tfPath)
		if err != nil {
			return fmt.Errorf("failed to create terraform: %w", err)
		}

		cfg := applyApplyOptions(opts...)

		// If auto-approve is set, pass the plan file as nil to trigger auto-apply
		// Otherwise, terraform will prompt for confirmation
		if cfg.AutoApprove {
			tf.SetStdout(nil)
		}

		var tfOpts []tfexec.ApplyOption
		if cfg.VarFile != "" {
			tfOpts = append(tfOpts, tfexec.VarFile(cfg.VarFile))
		}
		if cfg.PlanFile != "" {
			tfOpts = append(tfOpts, tfexec.DirOrPlan(cfg.PlanFile))
		}

		return tf.Apply(ctx, tfOpts...)
	}
}

// makeDestroyFunc returns a closure that executes terraform destroy
func makeDestroyFunc(tfPath string) DestroyFunc {
	return func(ctx context.Context, dir string, opts ...DestroyOption) error {
		tf, err := tfexec.NewTerraform(dir, tfPath)
		if err != nil {
			return fmt.Errorf("failed to create terraform: %w", err)
		}

		cfg := applyDestroyOptions(opts...)

		// If auto-approve is set, suppress prompts
		if cfg.AutoApprove {
			tf.SetStdout(nil)
		}

		var tfOpts []tfexec.DestroyOption
		if cfg.VarFile != "" {
			tfOpts = append(tfOpts, tfexec.VarFile(cfg.VarFile))
		}

		return tf.Destroy(ctx, tfOpts...)
	}
}

// makeOutputFunc returns a closure that retrieves terraform outputs
func makeOutputFunc(tfPath string) OutputFunc {
	return func(ctx context.Context, dir string) (map[string]interface{}, error) {
		tf, err := tfexec.NewTerraform(dir, tfPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create terraform: %w", err)
		}

		outputs, err := tf.Output(ctx)
		if err != nil {
			return nil, err
		}

		// Convert terraform-exec's OutputMeta to simple map
		// OutputMeta.Value is json.RawMessage, need to unmarshal it
		result := make(map[string]interface{})
		for k, v := range outputs {
			var value interface{}
			if err := json.Unmarshal(v.Value, &value); err != nil {
				return nil, fmt.Errorf("failed to unmarshal output %s: %w", k, err)
			}
			result[k] = value
		}

		return result, nil
	}
}

// makeValidateFunc returns a closure that validates terraform configuration
func makeValidateFunc(tfPath string) ValidateFunc {
	return func(ctx context.Context, dir string) error {
		tf, err := tfexec.NewTerraform(dir, tfPath)
		if err != nil {
			return fmt.Errorf("failed to create terraform: %w", err)
		}

		_, err = tf.Validate(ctx)
		return err
	}
}
