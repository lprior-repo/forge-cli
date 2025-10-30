package appconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	t.Run("creates module with sensible defaults", func(t *testing.T) {
		name := "test_app"
		module := NewModule(name)

		// Verify basic properties
		require.NotNil(t, module)
		assert.Equal(t, "terraform-aws-modules/appconfig/aws", module.Source)
		assert.Equal(t, "~> 2.0", module.Version)
		assert.NotNil(t, module.Name)
		assert.Equal(t, name, *module.Name)

		// Verify sensible defaults
		assert.NotNil(t, module.Create)
		assert.True(t, *module.Create)

		assert.NotNil(t, module.ConfigProfileLocationURI)
		assert.Equal(t, "hosted", *module.ConfigProfileLocationURI)
	})

	t.Run("creates module with different names", func(t *testing.T) {
		names := []string{"app1", "my-config", "feature_flags"}
		for _, name := range names {
			module := NewModule(name)
			assert.NotNil(t, module.Name)
			assert.Equal(t, name, *module.Name)
		}
	})

	t.Run("creates module with empty name", func(t *testing.T) {
		module := NewModule("")
		assert.NotNil(t, module.Name)
		assert.Equal(t, "", *module.Name)
	})
}

func TestModule_WithEnvironment(t *testing.T) {
	t.Run("adds environment", func(t *testing.T) {
		module := NewModule("test_app")
		desc := "Production environment"
		env := Environment{
			Name:        "production",
			Description: &desc,
			Tags: map[string]string{
				"Environment": "prod",
			},
		}
		result := module.WithEnvironment("prod", env)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify environment is added
		assert.NotNil(t, module.Environments)
		assert.Len(t, module.Environments, 1)
		assert.Equal(t, "production", module.Environments["prod"].Name)
		assert.Equal(t, "Production environment", *module.Environments["prod"].Description)
	})

	t.Run("adds multiple environments", func(t *testing.T) {
		module := NewModule("test_app")

		module.WithEnvironment("dev", Environment{Name: "development"})
		module.WithEnvironment("prod", Environment{Name: "production"})

		assert.Len(t, module.Environments, 2)
		assert.Equal(t, "development", module.Environments["dev"].Name)
		assert.Equal(t, "production", module.Environments["prod"].Name)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test_app").
			WithEnvironment("dev", Environment{Name: "development"})

		assert.NotNil(t, module.Environments)
		assert.Len(t, module.Environments, 1)
	})
}

func TestModule_WithFeatureFlags(t *testing.T) {
	t.Run("configures feature flags", func(t *testing.T) {
		content := `{"flags": {"new_feature": {"enabled": true}}}`
		module := NewModule("test_app")
		result := module.WithFeatureFlags(content)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify configuration
		assert.NotNil(t, module.ConfigProfileType)
		assert.Equal(t, "AWS.AppConfig.FeatureFlags", *module.ConfigProfileType)

		assert.NotNil(t, module.CreateHostedConfigurationVersion)
		assert.True(t, *module.CreateHostedConfigurationVersion)

		assert.NotNil(t, module.HostedConfigurationVersionContent)
		assert.Equal(t, content, *module.HostedConfigurationVersionContent)

		assert.NotNil(t, module.HostedConfigurationVersionContentType)
		assert.Equal(t, "application/json", *module.HostedConfigurationVersionContentType)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test_app").
			WithFeatureFlags(`{"flags": {}}`)

		assert.NotNil(t, module.ConfigProfileType)
		assert.Equal(t, "AWS.AppConfig.FeatureFlags", *module.ConfigProfileType)
	})
}

func TestModule_WithFreeformConfig(t *testing.T) {
	t.Run("configures freeform configuration", func(t *testing.T) {
		content := `{"key": "value"}`
		contentType := "application/json"
		module := NewModule("test_app")
		result := module.WithFreeformConfig(content, contentType)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify configuration
		assert.NotNil(t, module.ConfigProfileType)
		assert.Equal(t, "AWS.Freeform", *module.ConfigProfileType)

		assert.NotNil(t, module.CreateHostedConfigurationVersion)
		assert.True(t, *module.CreateHostedConfigurationVersion)

		assert.NotNil(t, module.HostedConfigurationVersionContent)
		assert.Equal(t, content, *module.HostedConfigurationVersionContent)

		assert.NotNil(t, module.HostedConfigurationVersionContentType)
		assert.Equal(t, contentType, *module.HostedConfigurationVersionContentType)
	})

	t.Run("supports different content types", func(t *testing.T) {
		tests := []struct {
			name        string
			content     string
			contentType string
		}{
			{"json", `{"key": "value"}`, "application/json"},
			{"yaml", "key: value", "application/x-yaml"},
			{"text", "plain text", "text/plain"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				module := NewModule("test_app").WithFreeformConfig(tt.content, tt.contentType)

				assert.Equal(t, tt.content, *module.HostedConfigurationVersionContent)
				assert.Equal(t, tt.contentType, *module.HostedConfigurationVersionContentType)
			})
		}
	})
}

func TestModule_WithDeploymentStrategy(t *testing.T) {
	t.Run("adds deployment strategy", func(t *testing.T) {
		module := NewModule("test_app")
		result := module.WithDeploymentStrategy(10, 25.0, 5)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify configuration
		assert.NotNil(t, module.CreateDeploymentStrategy)
		assert.True(t, *module.CreateDeploymentStrategy)

		assert.NotNil(t, module.DeploymentDurationInMinutes)
		assert.Equal(t, 10, *module.DeploymentDurationInMinutes)

		assert.NotNil(t, module.GrowthFactor)
		assert.Equal(t, 25.0, *module.GrowthFactor)

		assert.NotNil(t, module.GrowthType)
		assert.Equal(t, "LINEAR", *module.GrowthType)

		assert.NotNil(t, module.FinalBakeTimeInMinutes)
		assert.Equal(t, 5, *module.FinalBakeTimeInMinutes)
	})

	t.Run("supports different strategy values", func(t *testing.T) {
		tests := []struct {
			name         string
			duration     int
			growthFactor float64
			bakeTime     int
		}{
			{"fast", 5, 50.0, 0},
			{"slow", 60, 10.0, 15},
			{"instant", 0, 100.0, 0},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				module := NewModule("test_app").WithDeploymentStrategy(tt.duration, tt.growthFactor, tt.bakeTime)

				assert.Equal(t, tt.duration, *module.DeploymentDurationInMinutes)
				assert.Equal(t, tt.growthFactor, *module.GrowthFactor)
				assert.Equal(t, tt.bakeTime, *module.FinalBakeTimeInMinutes)
			})
		}
	})
}

func TestModule_WithValidator(t *testing.T) {
	t.Run("adds JSON schema validator", func(t *testing.T) {
		schema := `{"type": "object"}`
		module := NewModule("test_app")
		result := module.WithValidator("JSON_SCHEMA", schema)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify validator is added
		assert.Len(t, module.ConfigProfileValidator, 1)
		assert.Equal(t, "JSON_SCHEMA", module.ConfigProfileValidator[0].Type)
		assert.Equal(t, schema, module.ConfigProfileValidator[0].Content)
	})

	t.Run("adds Lambda validator", func(t *testing.T) {
		lambdaARN := "arn:aws:lambda:us-east-1:123456789012:function:validator"
		module := NewModule("test_app")
		module.WithValidator("LAMBDA", lambdaARN)

		assert.Len(t, module.ConfigProfileValidator, 1)
		assert.Equal(t, "LAMBDA", module.ConfigProfileValidator[0].Type)
		assert.Equal(t, lambdaARN, module.ConfigProfileValidator[0].Content)
	})

	t.Run("adds multiple validators", func(t *testing.T) {
		module := NewModule("test_app")
		module.WithValidator("JSON_SCHEMA", `{"type": "object"}`)
		module.WithValidator("LAMBDA", "arn:aws:lambda:...")

		assert.Len(t, module.ConfigProfileValidator, 2)
		assert.Equal(t, "JSON_SCHEMA", module.ConfigProfileValidator[0].Type)
		assert.Equal(t, "LAMBDA", module.ConfigProfileValidator[1].Type)
	})
}

func TestModule_WithTags(t *testing.T) {
	t.Run("adds tags to module", func(t *testing.T) {
		module := NewModule("test_app")
		tags := map[string]string{
			"Environment": "production",
			"Team":        "platform",
		}
		result := module.WithTags(tags)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify tags are set
		assert.NotNil(t, module.Tags)
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("merges tags when called multiple times", func(t *testing.T) {
		module := NewModule("test_app")

		tags1 := map[string]string{"Environment": "production"}
		module.WithTags(tags1)

		tags2 := map[string]string{"Team": "platform"}
		module.WithTags(tags2)

		// Verify both sets of tags are present
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("overwrites existing tags with same key", func(t *testing.T) {
		module := NewModule("test_app")

		tags1 := map[string]string{"Environment": "development"}
		module.WithTags(tags1)

		tags2 := map[string]string{"Environment": "production"}
		module.WithTags(tags2)

		// Verify tag was overwritten
		assert.Equal(t, "production", module.Tags["Environment"])
	})
}

func TestModule_LocalName(t *testing.T) {
	t.Run("returns name when set", func(t *testing.T) {
		name := "my_app"
		module := NewModule(name)

		assert.Equal(t, name, module.LocalName())
	})

	t.Run("returns default when name is nil", func(t *testing.T) {
		module := &Module{}

		assert.Equal(t, "appconfig", module.LocalName())
	})

	t.Run("returns empty string when name is empty", func(t *testing.T) {
		emptyName := ""
		module := NewModule(emptyName)

		assert.Equal(t, emptyName, module.LocalName())
	})
}

func TestModule_Configuration(t *testing.T) {
	t.Run("returns empty string and nil error as placeholder", func(t *testing.T) {
		module := NewModule("test_app")

		config, err := module.Configuration()

		// Current implementation is a placeholder
		require.NoError(t, err)
		assert.Empty(t, config)
	})
}

func TestModule_FluentAPI(t *testing.T) {
	t.Run("supports complete fluent configuration", func(t *testing.T) {
		content := `{"flags": {"feature": {"enabled": true}}}`
		module := NewModule("my_app").
			WithFeatureFlags(content).
			WithEnvironment("prod", Environment{Name: "production"}).
			WithDeploymentStrategy(10, 25.0, 5).
			WithTags(map[string]string{"Team": "platform"})

		// Verify all configuration is applied
		assert.NotNil(t, module.Name)
		assert.Equal(t, "my_app", *module.Name)

		assert.NotNil(t, module.ConfigProfileType)
		assert.Equal(t, "AWS.AppConfig.FeatureFlags", *module.ConfigProfileType)

		assert.Len(t, module.Environments, 1)
		assert.Equal(t, "production", module.Environments["prod"].Name)

		assert.NotNil(t, module.DeploymentDurationInMinutes)
		assert.Equal(t, 10, *module.DeploymentDurationInMinutes)

		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("supports freeform configuration fluent API", func(t *testing.T) {
		module := NewModule("config_app").
			WithFreeformConfig(`{"key": "value"}`, "application/json").
			WithValidator("JSON_SCHEMA", `{"type": "object"}`).
			WithTags(map[string]string{"Type": "freeform"})

		assert.Equal(t, "AWS.Freeform", *module.ConfigProfileType)
		assert.Len(t, module.ConfigProfileValidator, 1)
		assert.Equal(t, "freeform", module.Tags["Type"])
	})
}

func TestEnvironment(t *testing.T) {
	t.Run("creates environment with monitors", func(t *testing.T) {
		roleARN := "arn:aws:iam::123456789012:role/AppConfigMonitor"
		env := Environment{
			Name: "production",
			Monitors: []Monitor{
				{
					AlarmARN:     "arn:aws:cloudwatch:us-east-1:123456789012:alarm:HighErrorRate",
					AlarmRoleARN: &roleARN,
				},
			},
		}

		assert.Equal(t, "production", env.Name)
		assert.Len(t, env.Monitors, 1)
		assert.Equal(t, "arn:aws:cloudwatch:us-east-1:123456789012:alarm:HighErrorRate", env.Monitors[0].AlarmARN)
		assert.NotNil(t, env.Monitors[0].AlarmRoleARN)
	})

	t.Run("creates environment with tags", func(t *testing.T) {
		env := Environment{
			Name: "staging",
			Tags: map[string]string{
				"Environment": "staging",
				"CostCenter":  "engineering",
			},
		}

		assert.Equal(t, "staging", env.Name)
		assert.Equal(t, "staging", env.Tags["Environment"])
		assert.Equal(t, "engineering", env.Tags["CostCenter"])
	})
}

func TestValidator(t *testing.T) {
	t.Run("creates JSON schema validator", func(t *testing.T) {
		v := Validator{
			Type:    "JSON_SCHEMA",
			Content: `{"type": "object", "required": ["name"]}`,
		}

		assert.Equal(t, "JSON_SCHEMA", v.Type)
		assert.Contains(t, v.Content, "required")
	})

	t.Run("creates Lambda validator", func(t *testing.T) {
		v := Validator{
			Type:    "LAMBDA",
			Content: "arn:aws:lambda:us-east-1:123456789012:function:validator",
		}

		assert.Equal(t, "LAMBDA", v.Type)
		assert.Contains(t, v.Content, "arn:aws:lambda")
	})
}

func TestExtensionAction(t *testing.T) {
	t.Run("creates Lambda extension action", func(t *testing.T) {
		roleARN := "arn:aws:iam::123456789012:role/AppConfigExtension"
		desc := "Sends notifications on deployment"
		action := ExtensionAction{
			Name:        "notify-deployment",
			URI:         "arn:aws:lambda:us-east-1:123456789012:function:notify",
			RoleARN:     &roleARN,
			Description: &desc,
		}

		assert.Equal(t, "notify-deployment", action.Name)
		assert.Contains(t, action.URI, "arn:aws:lambda")
		assert.NotNil(t, action.RoleARN)
		assert.Equal(t, "Sends notifications on deployment", *action.Description)
	})

	t.Run("creates SNS extension action", func(t *testing.T) {
		action := ExtensionAction{
			Name: "sns-notify",
			URI:  "arn:aws:sns:us-east-1:123456789012:appconfig-events",
		}

		assert.Equal(t, "sns-notify", action.Name)
		assert.Contains(t, action.URI, "arn:aws:sns")
	})
}

func TestExtensionParameter(t *testing.T) {
	t.Run("creates required parameter", func(t *testing.T) {
		required := true
		desc := "API endpoint URL"
		param := ExtensionParameter{
			Required:    &required,
			Description: &desc,
		}

		assert.NotNil(t, param.Required)
		assert.True(t, *param.Required)
		assert.Equal(t, "API endpoint URL", *param.Description)
	})

	t.Run("creates optional parameter", func(t *testing.T) {
		required := false
		param := ExtensionParameter{
			Required: &required,
		}

		assert.NotNil(t, param.Required)
		assert.False(t, *param.Required)
	})
}

// BenchmarkNewModule benchmarks module creation
func BenchmarkNewModule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_app")
	}
}

// BenchmarkFluentAPI benchmarks fluent API calls
func BenchmarkFluentAPI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_app").
			WithFeatureFlags(`{"flags": {}}`).
			WithEnvironment("prod", Environment{Name: "production"}).
			WithTags(map[string]string{"Environment": "production"})
	}
}

// BenchmarkWithEnvironment benchmarks environment addition
func BenchmarkWithEnvironment(b *testing.B) {
	module := NewModule("bench_app")
	env := Environment{Name: "production"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		module.WithEnvironment("prod", env)
	}
}
