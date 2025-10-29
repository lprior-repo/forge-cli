package terraform

// InitOption is a function that configures InitConfig
type InitOption func(*InitConfig)

// InitConfig holds init configuration
type InitConfig struct {
	Upgrade       bool
	Backend       bool
	Reconfigure   bool
	BackendConfig []string
}

// Upgrade enables the -upgrade flag
func Upgrade(v bool) InitOption {
	return func(cfg *InitConfig) {
		cfg.Upgrade = v
	}
}

// Backend enables/disables backend configuration
func Backend(v bool) InitOption {
	return func(cfg *InitConfig) {
		cfg.Backend = v
	}
}

// Reconfigure enables the -reconfigure flag
func Reconfigure(v bool) InitOption {
	return func(cfg *InitConfig) {
		cfg.Reconfigure = v
	}
}

// BackendConfig adds a backend configuration option
func BackendConfig(value string) InitOption {
	return func(cfg *InitConfig) {
		cfg.BackendConfig = append(cfg.BackendConfig, value)
	}
}

// applyInitOptions applies all options and returns the config
func applyInitOptions(opts ...InitOption) InitConfig {
	cfg := InitConfig{
		Backend: true, // default to true
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// PlanOption is a function that configures PlanConfig
type PlanOption func(*PlanConfig)

// PlanConfig holds plan configuration
type PlanConfig struct {
	Out     string
	Destroy bool
	VarFile string
	Vars    map[string]string
}

// PlanOut specifies the output file for the plan
func PlanOut(path string) PlanOption {
	return func(cfg *PlanConfig) {
		cfg.Out = path
	}
}

// PlanDestroy creates a destroy plan
func PlanDestroy(v bool) PlanOption {
	return func(cfg *PlanConfig) {
		cfg.Destroy = v
	}
}

// PlanVarFile specifies a tfvars file
func PlanVarFile(path string) PlanOption {
	return func(cfg *PlanConfig) {
		cfg.VarFile = path
	}
}

// PlanVar adds a variable
func PlanVar(key, value string) PlanOption {
	return func(cfg *PlanConfig) {
		if cfg.Vars == nil {
			cfg.Vars = make(map[string]string)
		}
		cfg.Vars[key] = value
	}
}

// applyPlanOptions applies all options and returns the config
func applyPlanOptions(opts ...PlanOption) PlanConfig {
	cfg := PlanConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// ApplyOption is a function that configures ApplyConfig
type ApplyOption func(*ApplyConfig)

// ApplyConfig holds apply configuration
type ApplyConfig struct {
	AutoApprove bool
	VarFile     string
	Vars        map[string]string
	PlanFile    string
}

// AutoApprove skips interactive approval
func AutoApprove(v bool) ApplyOption {
	return func(cfg *ApplyConfig) {
		cfg.AutoApprove = v
	}
}

// ApplyVarFile specifies a tfvars file
func ApplyVarFile(path string) ApplyOption {
	return func(cfg *ApplyConfig) {
		cfg.VarFile = path
	}
}

// ApplyVar adds a variable
func ApplyVar(key, value string) ApplyOption {
	return func(cfg *ApplyConfig) {
		if cfg.Vars == nil {
			cfg.Vars = make(map[string]string)
		}
		cfg.Vars[key] = value
	}
}

// ApplyPlanFile applies a saved plan file
func ApplyPlanFile(path string) ApplyOption {
	return func(cfg *ApplyConfig) {
		cfg.PlanFile = path
	}
}

// applyApplyOptions applies all options and returns the config
func applyApplyOptions(opts ...ApplyOption) ApplyConfig {
	cfg := ApplyConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// DestroyOption is a function that configures DestroyConfig
type DestroyOption func(*DestroyConfig)

// DestroyConfig holds destroy configuration
type DestroyConfig struct {
	AutoApprove bool
	VarFile     string
	Vars        map[string]string
}

// DestroyAutoApprove skips interactive approval for destroy
func DestroyAutoApprove(v bool) DestroyOption {
	return func(cfg *DestroyConfig) {
		cfg.AutoApprove = v
	}
}

// DestroyVarFile specifies a tfvars file for destroy
func DestroyVarFile(path string) DestroyOption {
	return func(cfg *DestroyConfig) {
		cfg.VarFile = path
	}
}

// applyDestroyOptions applies all options and returns the config
func applyDestroyOptions(opts ...DestroyOption) DestroyConfig {
	cfg := DestroyConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
