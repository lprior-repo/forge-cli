package eventbridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	t.Run("creates module with sensible defaults", func(t *testing.T) {
		name := "test_bus"
		module := NewModule(name)

		require.NotNil(t, module)
		assert.Equal(t, "terraform-aws-modules/eventbridge/aws", module.Source)
		assert.Equal(t, "~> 3.0", module.Version)
		assert.NotNil(t, module.BusName)
		assert.Equal(t, name, *module.BusName)

		// Verify sensible defaults
		assert.NotNil(t, module.Create)
		assert.True(t, *module.Create)

		assert.NotNil(t, module.CreateBus)
		assert.True(t, *module.CreateBus)

		assert.NotNil(t, module.CreateRules)
		assert.True(t, *module.CreateRules)

		assert.NotNil(t, module.CreateTargets)
		assert.True(t, *module.CreateTargets)

		assert.NotNil(t, module.CreateRole)
		assert.True(t, *module.CreateRole)

		assert.NotNil(t, module.AppendRulePostfix)
		assert.True(t, *module.AppendRulePostfix)
	})

	t.Run("creates module with different names", func(t *testing.T) {
		names := []string{"orders", "events", "app-events"}
		for _, name := range names {
			module := NewModule(name)
			assert.NotNil(t, module.BusName)
			assert.Equal(t, name, *module.BusName)
		}
	})
}

func TestModule_WithRule(t *testing.T) {
	t.Run("adds a rule", func(t *testing.T) {
		name := "daily-trigger"
		desc := "Runs daily at midnight"
		scheduleExpr := "rate(1 day)"

		rule := Rule{
			Name:               &name,
			Description:        &desc,
			ScheduleExpression: &scheduleExpr,
		}

		module := NewModule("test_bus")
		result := module.WithRule("daily", rule)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Rules)
		assert.Len(t, module.Rules, 1)
		assert.Equal(t, name, *module.Rules["daily"].Name)
		assert.Equal(t, scheduleExpr, *module.Rules["daily"].ScheduleExpression)
	})

	t.Run("adds event pattern rule", func(t *testing.T) {
		eventPattern := `{"source": ["aws.ec2"]}`
		rule := Rule{
			EventPattern: &eventPattern,
		}

		module := NewModule("test_bus")
		module.WithRule("ec2_events", rule)

		assert.NotNil(t, module.Rules["ec2_events"].EventPattern)
		assert.Equal(t, eventPattern, *module.Rules["ec2_events"].EventPattern)
	})

	t.Run("adds multiple rules", func(t *testing.T) {
		module := NewModule("test_bus")

		module.WithRule("rule1", Rule{})
		module.WithRule("rule2", Rule{})

		assert.Len(t, module.Rules, 2)
	})
}

func TestModule_WithTarget(t *testing.T) {
	t.Run("adds a target to a rule", func(t *testing.T) {
		target := Target{
			ARN: "arn:aws:lambda:us-east-1:123456789012:function:processor",
		}

		module := NewModule("test_bus")
		result := module.WithTarget("rule1", target)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Targets)
		assert.Len(t, module.Targets["rule1"], 1)
		assert.Equal(t, target.ARN, module.Targets["rule1"][0].ARN)
	})

	t.Run("adds multiple targets to same rule", func(t *testing.T) {
		module := NewModule("test_bus")

		target1 := Target{ARN: "arn1"}
		module.WithTarget("rule1", target1)

		target2 := Target{ARN: "arn2"}
		module.WithTarget("rule1", target2)

		assert.Len(t, module.Targets["rule1"], 2)
	})

	t.Run("adds targets to different rules", func(t *testing.T) {
		module := NewModule("test_bus")

		module.WithTarget("rule1", Target{ARN: "arn1"})
		module.WithTarget("rule2", Target{ARN: "arn2"})

		assert.Len(t, module.Targets, 2)
		assert.Contains(t, module.Targets, "rule1")
		assert.Contains(t, module.Targets, "rule2")
	})
}

func TestModule_WithSchedule(t *testing.T) {
	t.Run("adds a schedule", func(t *testing.T) {
		schedule := Schedule{
			Name:               "hourly-task",
			ScheduleExpression: "rate(1 hour)",
		}

		module := NewModule("test_bus")
		result := module.WithSchedule("hourly", schedule)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.CreateSchedules)
		assert.True(t, *module.CreateSchedules)
		assert.NotNil(t, module.Schedules)
		assert.Len(t, module.Schedules, 1)
		assert.Equal(t, "hourly-task", module.Schedules["hourly"].Name)
	})

	t.Run("adds schedule with cron expression", func(t *testing.T) {
		schedule := Schedule{
			Name:               "midnight-task",
			ScheduleExpression: "cron(0 0 * * ? *)",
		}

		module := NewModule("test_bus")
		module.WithSchedule("midnight", schedule)

		assert.Equal(t, "cron(0 0 * * ? *)", module.Schedules["midnight"].ScheduleExpression)
	})
}

func TestModule_WithPipe(t *testing.T) {
	t.Run("adds a pipe", func(t *testing.T) {
		pipe := Pipe{
			Name:    "sqs-to-lambda",
			Source:  "arn:aws:sqs:us-east-1:123456789012:queue",
			Target:  "arn:aws:lambda:us-east-1:123456789012:function:processor",
			RoleARN: "arn:aws:iam::123456789012:role/pipe-role",
		}

		module := NewModule("test_bus")
		result := module.WithPipe("pipe1", pipe)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.CreatePipes)
		assert.True(t, *module.CreatePipes)
		assert.NotNil(t, module.Pipes)
		assert.Len(t, module.Pipes, 1)
		assert.Equal(t, "sqs-to-lambda", module.Pipes["pipe1"].Name)
	})

	t.Run("adds pipe with enrichment", func(t *testing.T) {
		enrichment := "arn:aws:lambda:us-east-1:123456789012:function:enricher"
		pipe := Pipe{
			Name:       "enriched-pipe",
			Source:     "source-arn",
			Target:     "target-arn",
			RoleARN:    "role-arn",
			Enrichment: &enrichment,
		}

		module := NewModule("test_bus")
		module.WithPipe("enriched", pipe)

		assert.NotNil(t, module.Pipes["enriched"].Enrichment)
		assert.Equal(t, enrichment, *module.Pipes["enriched"].Enrichment)
	})
}

func TestModule_WithTags(t *testing.T) {
	t.Run("adds tags to the event bus", func(t *testing.T) {
		tags := map[string]string{
			"Environment": "production",
			"Team":        "platform",
		}

		module := NewModule("test_bus")
		result := module.WithTags(tags)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Tags)
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("merges tags when called multiple times", func(t *testing.T) {
		module := NewModule("test_bus")

		module.WithTags(map[string]string{"Key1": "value1"})
		module.WithTags(map[string]string{"Key2": "value2"})

		assert.Equal(t, "value1", module.Tags["Key1"])
		assert.Equal(t, "value2", module.Tags["Key2"])
	})
}

func TestModule_LocalName(t *testing.T) {
	t.Run("returns bus name when set", func(t *testing.T) {
		name := "my_bus"
		module := NewModule(name)

		assert.Equal(t, name, module.LocalName())
	})

	t.Run("returns default when name is nil", func(t *testing.T) {
		module := &Module{}

		assert.Equal(t, "eventbridge", module.LocalName())
	})
}

func TestModule_Configuration(t *testing.T) {
	t.Run("returns empty string and nil error as placeholder", func(t *testing.T) {
		module := NewModule("test_bus")

		config, err := module.Configuration()

		require.NoError(t, err)
		assert.Empty(t, config)
	})
}

func TestModule_FluentAPI(t *testing.T) {
	t.Run("supports complete fluent configuration", func(t *testing.T) {
		rule := Rule{
			ScheduleExpression: ptr("rate(1 hour)"),
		}

		target := Target{
			ARN: "arn:aws:lambda:us-east-1:123456789012:function:processor",
		}

		module := NewModule("app-events").
			WithRule("hourly", rule).
			WithTarget("hourly", target).
			WithTags(map[string]string{"Team": "platform"})

		assert.NotNil(t, module.BusName)
		assert.Equal(t, "app-events", *module.BusName)
		assert.Len(t, module.Rules, 1)
		assert.Len(t, module.Targets["hourly"], 1)
		assert.Equal(t, "platform", module.Tags["Team"])
	})
}

func TestRule(t *testing.T) {
	t.Run("creates schedule expression rule", func(t *testing.T) {
		name := "daily-backup"
		desc := "Daily backup job"
		expr := "rate(1 day)"
		enabled := true

		rule := Rule{
			Name:               &name,
			Description:        &desc,
			ScheduleExpression: &expr,
			Enabled:            &enabled,
		}

		assert.Equal(t, "daily-backup", *rule.Name)
		assert.Equal(t, "rate(1 day)", *rule.ScheduleExpression)
		assert.True(t, *rule.Enabled)
	})

	t.Run("creates event pattern rule", func(t *testing.T) {
		pattern := `{"source": ["aws.s3"], "detail-type": ["Object Created"]}`
		rule := Rule{
			EventPattern: &pattern,
		}

		assert.NotNil(t, rule.EventPattern)
		assert.Contains(t, *rule.EventPattern, "aws.s3")
	})
}

func TestTarget(t *testing.T) {
	t.Run("creates target with input transformation", func(t *testing.T) {
		transformer := &InputTransformer{
			InputPaths: map[string]string{
				"time":   "$.time",
				"detail": "$.detail",
			},
			InputTemplate: `{"timestamp": <time>, "data": <detail>}`,
		}

		target := Target{
			ARN:              "arn:aws:lambda:us-east-1:123456789012:function:processor",
			InputTransformer: transformer,
		}

		assert.NotNil(t, target.InputTransformer)
		assert.Len(t, target.InputTransformer.InputPaths, 2)
	})

	t.Run("creates target with retry policy", func(t *testing.T) {
		maxAge := 3600
		maxRetries := 2

		retry := &RetryPolicy{
			MaximumEventAge:      &maxAge,
			MaximumRetryAttempts: &maxRetries,
		}

		target := Target{
			ARN:         "arn",
			RetryPolicy: retry,
		}

		assert.NotNil(t, target.RetryPolicy)
		assert.Equal(t, 3600, *target.RetryPolicy.MaximumEventAge)
		assert.Equal(t, 2, *target.RetryPolicy.MaximumRetryAttempts)
	})

	t.Run("creates target with DLQ", func(t *testing.T) {
		dlqARN := "arn:aws:sqs:us-east-1:123456789012:dlq"
		target := Target{
			ARN:           "arn",
			DeadLetterARN: &dlqARN,
		}

		assert.NotNil(t, target.DeadLetterARN)
		assert.Equal(t, dlqARN, *target.DeadLetterARN)
	})
}

func TestArchive(t *testing.T) {
	t.Run("creates archive with retention", func(t *testing.T) {
		desc := "Archive all events"
		retention := 7

		archive := Archive{
			Name:          "weekly-archive",
			Description:   &desc,
			RetentionDays: &retention,
		}

		assert.Equal(t, "weekly-archive", archive.Name)
		assert.Equal(t, 7, *archive.RetentionDays)
	})

	t.Run("creates archive with event pattern", func(t *testing.T) {
		pattern := `{"source": ["custom.app"]}`
		archive := Archive{
			Name:         "filtered-archive",
			EventPattern: &pattern,
		}

		assert.NotNil(t, archive.EventPattern)
	})
}

func TestPermission(t *testing.T) {
	t.Run("creates permission for account", func(t *testing.T) {
		action := "events:PutEvents"
		perm := Permission{
			Principal:   "123456789012",
			StatementID: "AllowAccountAccess",
			Action:      &action,
		}

		assert.Equal(t, "123456789012", perm.Principal)
		assert.Equal(t, "events:PutEvents", *perm.Action)
	})

	t.Run("creates permission with condition", func(t *testing.T) {
		condition := &PermissionCondition{
			Type:  "StringEquals",
			Key:   "aws:SourceAccount",
			Value: "123456789012",
		}

		perm := Permission{
			Principal:   "service.amazonaws.com",
			StatementID: "AllowService",
			Condition:   condition,
		}

		assert.NotNil(t, perm.Condition)
		assert.Equal(t, "StringEquals", perm.Condition.Type)
	})
}

func TestSchedule(t *testing.T) {
	t.Run("creates schedule with flexible time window", func(t *testing.T) {
		maxWindow := 15
		window := &FlexibleTimeWindow{
			Mode:                   "FLEXIBLE",
			MaximumWindowInMinutes: &maxWindow,
		}

		schedule := Schedule{
			Name:               "flexible-task",
			ScheduleExpression: "rate(1 hour)",
			FlexibleTimeWindow: window,
		}

		assert.NotNil(t, schedule.FlexibleTimeWindow)
		assert.Equal(t, "FLEXIBLE", schedule.FlexibleTimeWindow.Mode)
		assert.Equal(t, 15, *schedule.FlexibleTimeWindow.MaximumWindowInMinutes)
	})

	t.Run("creates schedule with timezone", func(t *testing.T) {
		tz := "America/New_York"
		schedule := Schedule{
			Name:                       "daily-task",
			ScheduleExpression:         "cron(0 9 * * ? *)",
			ScheduleExpressionTimezone: &tz,
		}

		assert.Equal(t, "America/New_York", *schedule.ScheduleExpressionTimezone)
	})
}

func TestPipe(t *testing.T) {
	t.Run("creates pipe with source parameters", func(t *testing.T) {
		sourceParams := map[string]interface{}{
			"FilterCriteria": map[string]interface{}{
				"Filters": []map[string]interface{}{
					{"Pattern": `{"body": {"type": ["order"]}}`},
				},
			},
		}

		pipe := Pipe{
			Name:             "filtered-pipe",
			Source:           "sqs-arn",
			Target:           "lambda-arn",
			RoleARN:          "role-arn",
			SourceParameters: sourceParams,
		}

		assert.NotNil(t, pipe.SourceParameters)
	})
}

// Helper function to create pointer to string
func ptr(s string) *string {
	return &s
}

// BenchmarkNewModule benchmarks module creation
func BenchmarkNewModule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_bus")
	}
}

// BenchmarkFluentAPI benchmarks fluent API calls
func BenchmarkFluentAPI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		rule := Rule{ScheduleExpression: ptr("rate(1 hour)")}
		target := Target{ARN: "arn:aws:lambda:us-east-1:123456789012:function:test"}

		_ = NewModule("bench_bus").
			WithRule("test", rule).
			WithTarget("test", target).
			WithTags(map[string]string{"Environment": "production"})
	}
}
