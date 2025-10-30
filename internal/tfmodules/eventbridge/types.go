// Package eventbridge provides type-safe Terraform module definitions for terraform-aws-modules/eventbridge/aws.
// Generated from https://github.com/terraform-aws-modules/terraform-aws-eventbridge v3.0
package eventbridge

// Module represents the terraform-aws-modules/eventbridge/aws module.
// All fields use pointers to distinguish between "not set" (nil) and "set to zero value".
//
// Rule represents an EventBridge rule.
//
// Target represents an EventBridge target.
//
// InputTransformer represents input transformation configuration.
//
// RetryPolicy represents retry policy configuration.
//
// ECSTarget represents ECS task configuration for EventBridge target.
//
// NetworkConfiguration represents network configuration for ECS tasks.
//
// AWSVPCConfiguration represents VPC configuration.
//
// Archive represents an EventBridge archive.
//
// Permission represents an EventBridge permission.
//
// PermissionCondition represents a permission condition.
//
// Connection represents an EventBridge connection.
//
// APIDestination represents an EventBridge API destination.
//
// ScheduleGroup represents an EventBridge schedule group.
//
// Schedule represents an EventBridge schedule.
//
// FlexibleTimeWindow represents flexible time window configuration.
//
// ScheduleTarget represents a schedule target.
//
// Pipe represents an EventBridge pipe.
type (
	Module struct {
	// Source is the Terraform module source
	Source string `json:"source" hcl:"source,attr"`

	// Version is the module version constraint
	Version string `json:"version,omitempty" hcl:"version,attr"`

	// Create controls whether resources should be created
	Create *bool `json:"create,omitempty" hcl:"create,attr"`

	// CreateRole controls whether IAM roles should be created
	CreateRole *bool `json:"create_role,omitempty" hcl:"create_role,attr"`

	// Region where the resource(s) will be managed
	Region *string `json:"region,omitempty" hcl:"region,attr"`

	// Tags to assign to resources
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`

	// ================================
	// Event Bus
	// ================================

	// CreateBus controls whether EventBridge Bus should be created
	CreateBus *bool `json:"create_bus,omitempty" hcl:"create_bus,attr"`

	// BusName is a unique name for your EventBridge Bus
	BusName *string `json:"bus_name,omitempty" hcl:"bus_name,attr"`

	// BusDescription is the event bus description
	BusDescription *string `json:"bus_description,omitempty" hcl:"bus_description,attr"`

	// EventSourceName is the partner event source name
	EventSourceName *string `json:"event_source_name,omitempty" hcl:"event_source_name,attr"`

	// ================================
	// Rules
	// ================================

	// CreateRules controls whether EventBridge Rule resources should be created
	CreateRules *bool `json:"create_rules,omitempty" hcl:"create_rules,attr"`

	// AppendRulePostfix controls whether to append '-rule' to the rule name
	AppendRulePostfix *bool `json:"append_rule_postfix,omitempty" hcl:"append_rule_postfix,attr"`

	// Rules is a map of rule configurations
	Rules map[string]Rule `json:"rules,omitempty" hcl:"rules,attr"`

	// ================================
	// Targets
	// ================================

	// CreateTargets controls whether EventBridge Target resources should be created
	CreateTargets *bool `json:"create_targets,omitempty" hcl:"create_targets,attr"`

	// Targets is a map of target configurations
	Targets map[string][]Target `json:"targets,omitempty" hcl:"targets,attr"`

	// ================================
	// Archives
	// ================================

	// CreateArchives controls whether EventBridge Archive resources should be created
	CreateArchives *bool `json:"create_archives,omitempty" hcl:"create_archives,attr"`

	// Archives is a map of archive configurations
	Archives map[string]Archive `json:"archives,omitempty" hcl:"archives,attr"`

	// ================================
	// Permissions
	// ================================

	// CreatePermissions controls whether EventBridge Permission resources should be created
	CreatePermissions *bool `json:"create_permissions,omitempty" hcl:"create_permissions,attr"`

	// Permissions is a map of permission configurations
	Permissions map[string]Permission `json:"permissions,omitempty" hcl:"permissions,attr"`

	// ================================
	// API Destinations
	// ================================

	// CreateConnections controls whether EventBridge Connection resources should be created
	CreateConnections *bool `json:"create_connections,omitempty" hcl:"create_connections,attr"`

	// Connections is a map of connection configurations
	Connections map[string]Connection `json:"connections,omitempty" hcl:"connections,attr"`

	// CreateAPIDestinations controls whether EventBridge Destination resources should be created
	CreateAPIDestinations *bool `json:"create_api_destinations,omitempty" hcl:"create_api_destinations,attr"`

	// APIDestinations is a map of API destination configurations
	APIDestinations map[string]APIDestination `json:"api_destinations,omitempty" hcl:"api_destinations,attr"`

	// ================================
	// Schedules
	// ================================

	// CreateScheduleGroups controls whether EventBridge Schedule Group resources should be created
	CreateScheduleGroups *bool `json:"create_schedule_groups,omitempty" hcl:"create_schedule_groups,attr"`

	// ScheduleGroups is a map of schedule group configurations
	ScheduleGroups map[string]ScheduleGroup `json:"schedule_groups,omitempty" hcl:"schedule_groups,attr"`

	// CreateSchedules controls whether EventBridge Schedule resources should be created
	CreateSchedules *bool `json:"create_schedules,omitempty" hcl:"create_schedules,attr"`

	// Schedules is a map of schedule configurations
	Schedules map[string]Schedule `json:"schedules,omitempty" hcl:"schedules,attr"`

	// ================================
	// Pipes
	// ================================

	// CreatePipes controls whether EventBridge Pipes resources should be created
	CreatePipes *bool `json:"create_pipes,omitempty" hcl:"create_pipes,attr"`

	// Pipes is a map of pipe configurations
	Pipes map[string]Pipe `json:"pipes,omitempty" hcl:"pipes,attr"`

	// ================================
	// Schema Discovery
	// ================================

	// CreateSchemasDiscoverer controls whether default schemas discoverer should be created
	CreateSchemasDiscoverer *bool `json:"create_schemas_discoverer,omitempty" hcl:"create_schemas_discoverer,attr"`
	}

	Rule struct {
	// Name is the rule name (optional, auto-generated if not set)
	Name *string `json:"name,omitempty" hcl:"name,attr"`

	// Description of the rule
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// EventPattern is the event pattern JSON
	EventPattern *string `json:"event_pattern,omitempty" hcl:"event_pattern,attr"`

	// ScheduleExpression is the schedule expression (cron or rate)
	ScheduleExpression *string `json:"schedule_expression,omitempty" hcl:"schedule_expression,attr"`

	// EventBusName is the event bus to associate with this rule
	EventBusName *string `json:"event_bus_name,omitempty" hcl:"event_bus_name,attr"`

	// Enabled indicates if the rule is enabled
	Enabled *bool `json:"enabled,omitempty" hcl:"enabled,attr"`

	// RoleARN is the IAM role ARN for the rule
	RoleARN *string `json:"role_arn,omitempty" hcl:"role_arn,attr"`
	}

	Target struct {
	// Name is the target name
	Name *string `json:"name,omitempty" hcl:"name,attr"`

	// ARN is the target ARN
	ARN string `json:"arn" hcl:"arn,attr"`

	// RoleARN is the IAM role ARN for the target
	RoleARN *string `json:"role_arn,omitempty" hcl:"role_arn,attr"`

	// Input is the input JSON
	Input *string `json:"input,omitempty" hcl:"input,attr"`

	// InputPath is the JSONPath to select part of the event
	InputPath *string `json:"input_path,omitempty" hcl:"input_path,attr"`

	// InputTransformer transforms the event input
	InputTransformer *InputTransformer `json:"input_transformer,omitempty" hcl:"input_transformer,attr"`

	// DeadLetterARN is the dead letter queue ARN
	DeadLetterARN *string `json:"dead_letter_arn,omitempty" hcl:"dead_letter_arn,attr"`

	// RetryPolicy configures retry behavior
	RetryPolicy *RetryPolicy `json:"retry_policy,omitempty" hcl:"retry_policy,attr"`

	// ECSTarget configures ECS task parameters
	ECSTarget *ECSTarget `json:"ecs_target,omitempty" hcl:"ecs_target,attr"`
}

// InputTransformer represents input transformation configuration
type InputTransformer struct {
	// InputPaths maps JSONPath expressions to variables
	InputPaths map[string]string `json:"input_paths,omitempty" hcl:"input_paths,attr"`

	// InputTemplate is the template for transformation
	InputTemplate string `json:"input_template" hcl:"input_template,attr"`
}

// RetryPolicy represents retry policy configuration
type RetryPolicy struct {
	// MaximumEventAge is the maximum age in seconds (60-86400)
	MaximumEventAge *int `json:"maximum_event_age,omitempty" hcl:"maximum_event_age,attr"`

	// MaximumRetryAttempts is the maximum number of retries (0-185)
	MaximumRetryAttempts *int `json:"maximum_retry_attempts,omitempty" hcl:"maximum_retry_attempts,attr"`
}

// ECSTarget represents ECS task configuration for EventBridge target
type ECSTarget struct {
	// TaskDefinitionARN is the ECS task definition ARN
	TaskDefinitionARN string `json:"task_definition_arn" hcl:"task_definition_arn,attr"`

	// TaskCount is the number of tasks to create (default 1)
	TaskCount *int `json:"task_count,omitempty" hcl:"task_count,attr"`

	// LaunchType is the launch type (EC2 or FARGATE)
	LaunchType *string `json:"launch_type,omitempty" hcl:"launch_type,attr"`

	// PlatformVersion is the Fargate platform version
	PlatformVersion *string `json:"platform_version,omitempty" hcl:"platform_version,attr"`

	// NetworkConfiguration for Fargate tasks
	NetworkConfiguration *NetworkConfiguration `json:"network_configuration,omitempty" hcl:"network_configuration,attr"`
}

// NetworkConfiguration represents network configuration for ECS tasks
type NetworkConfiguration struct {
	// AWSVPCConfiguration for VPC settings
	AWSVPCConfiguration *AWSVPCConfiguration `json:"awsvpc_configuration,omitempty" hcl:"awsvpc_configuration,attr"`
}

// AWSVPCConfiguration represents VPC configuration
type AWSVPCConfiguration struct {
	// Subnets is the list of subnet IDs
	Subnets []string `json:"subnets" hcl:"subnets,attr"`

	// SecurityGroups is the list of security group IDs
	SecurityGroups []string `json:"security_groups,omitempty" hcl:"security_groups,attr"`

	// AssignPublicIP controls public IP assignment (ENABLED or DISABLED)
	AssignPublicIP *string `json:"assign_public_ip,omitempty" hcl:"assign_public_ip,attr"`
}

// Archive represents an EventBridge archive
type Archive struct {
	// Name is the archive name
	Name string `json:"name" hcl:"name,attr"`

	// Description of the archive
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// EventPattern is the event pattern JSON for filtering
	EventPattern *string `json:"event_pattern,omitempty" hcl:"event_pattern,attr"`

	// RetentionDays is the retention period in days (0 for indefinite)
	RetentionDays *int `json:"retention_days,omitempty" hcl:"retention_days,attr"`
}

// Permission represents an EventBridge permission
type Permission struct {
	// Principal is the AWS principal
	Principal string `json:"principal" hcl:"principal,attr"`

	// StatementID is the statement ID
	StatementID string `json:"statement_id" hcl:"statement_id,attr"`

	// Action is the action (e.g., "events:PutEvents")
	Action *string `json:"action,omitempty" hcl:"action,attr"`

	// Condition is the condition for the permission
	Condition *PermissionCondition `json:"condition,omitempty" hcl:"condition,attr"`
}

// PermissionCondition represents a permission condition
type PermissionCondition struct {
	// Type is the condition type
	Type string `json:"type" hcl:"type,attr"`

	// Key is the condition key
	Key string `json:"key" hcl:"key,attr"`

	// Value is the condition value
	Value string `json:"value" hcl:"value,attr"`
}

// Connection represents an EventBridge connection
type Connection struct {
	// Name is the connection name
	Name string `json:"name" hcl:"name,attr"`

	// Description of the connection
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// AuthorizationType is the authorization type
	// Valid values: "API_KEY" | "BASIC" | "OAUTH_CLIENT_CREDENTIALS" | "INVOCATION_HTTP_PARAMETERS"
	AuthorizationType string `json:"authorization_type" hcl:"authorization_type,attr"`

	// AuthParameters is the authorization parameters
	AuthParameters map[string]interface{} `json:"auth_parameters,omitempty" hcl:"auth_parameters,attr"`
}

// APIDestination represents an EventBridge API destination
type APIDestination struct {
	// Name is the API destination name
	Name string `json:"name" hcl:"name,attr"`

	// Description of the API destination
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// InvocationEndpoint is the HTTP endpoint URL
	InvocationEndpoint string `json:"invocation_endpoint" hcl:"invocation_endpoint,attr"`

	// HTTPMethod is the HTTP method
	// Valid values: "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "HEAD" | "OPTIONS"
	HTTPMethod string `json:"http_method" hcl:"http_method,attr"`

	// InvocationRateLimitPerSecond is the rate limit (1-300)
	InvocationRateLimitPerSecond *int `json:"invocation_rate_limit_per_second,omitempty" hcl:"invocation_rate_limit_per_second,attr"`

	// ConnectionARN is the connection ARN
	ConnectionARN string `json:"connection_arn" hcl:"connection_arn,attr"`
}

// ScheduleGroup represents an EventBridge schedule group
type ScheduleGroup struct {
	// Name is the schedule group name
	Name string `json:"name" hcl:"name,attr"`

	// Description of the schedule group
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// Tags for the schedule group
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`
}

// Schedule represents an EventBridge schedule
type Schedule struct {
	// Name is the schedule name
	Name string `json:"name" hcl:"name,attr"`

	// Description of the schedule
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// ScheduleExpression is the schedule expression (rate, cron, at)
	ScheduleExpression string `json:"schedule_expression" hcl:"schedule_expression,attr"`

	// ScheduleExpressionTimezone is the timezone
	ScheduleExpressionTimezone *string `json:"schedule_expression_timezone,omitempty" hcl:"schedule_expression_timezone,attr"`

	// FlexibleTimeWindow configures flexible execution window
	FlexibleTimeWindow *FlexibleTimeWindow `json:"flexible_time_window,omitempty" hcl:"flexible_time_window,attr"`

	// Target is the schedule target configuration
	Target *ScheduleTarget `json:"target,omitempty" hcl:"target,attr"`
}

// FlexibleTimeWindow represents flexible time window configuration
type FlexibleTimeWindow struct {
	// Mode is the flexible time window mode
	// Valid values: "OFF" | "FLEXIBLE"
	Mode string `json:"mode" hcl:"mode,attr"`

	// MaximumWindowInMinutes is the max window in minutes (1-1440)
	MaximumWindowInMinutes *int `json:"maximum_window_in_minutes,omitempty" hcl:"maximum_window_in_minutes,attr"`
}

// ScheduleTarget represents a schedule target
type ScheduleTarget struct {
	// ARN is the target ARN
	ARN string `json:"arn" hcl:"arn,attr"`

	// RoleARN is the IAM role ARN
	RoleARN string `json:"role_arn" hcl:"role_arn,attr"`

	// Input is the input JSON
	Input *string `json:"input,omitempty" hcl:"input,attr"`
}

// Pipe represents an EventBridge pipe
type Pipe struct {
	// Name is the pipe name
	Name string `json:"name" hcl:"name,attr"`

	// Description of the pipe
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// Source is the source ARN (SQS, DynamoDB Streams, Kinesis)
	Source string `json:"source" hcl:"source,attr"`

	// Target is the target ARN
	Target string `json:"target" hcl:"target,attr"`

	// RoleARN is the IAM role ARN
	RoleARN string `json:"role_arn" hcl:"role_arn,attr"`

	// Enrichment is the enrichment ARN (Lambda, API Gateway, etc.)
	Enrichment *string `json:"enrichment,omitempty" hcl:"enrichment,attr"`

	// SourceParameters configures source-specific settings
	SourceParameters map[string]interface{} `json:"source_parameters,omitempty" hcl:"source_parameters,attr"`

	// TargetParameters configures target-specific settings
	TargetParameters map[string]interface{} `json:"target_parameters,omitempty" hcl:"target_parameters,attr"`
}

// NewModule creates a new EventBridge module with sensible defaults
func NewModule(name string) *Module {
	source := "terraform-aws-modules/eventbridge/aws"
	version := "~> 3.0"
	create := true
	createBus := true
	createRules := true
	createTargets := true
	createRole := true
	appendPostfix := true

	return &Module{
		Source:            source,
		Version:           version,
		BusName:           &name,
		Create:            &create,
		CreateBus:         &createBus,
		CreateRules:       &createRules,
		CreateTargets:     &createTargets,
		CreateRole:        &createRole,
		AppendRulePostfix: &appendPostfix,
	}
}

// WithRule adds a rule
func (m *Module) WithRule(key string, rule Rule) *Module {
	if m.Rules == nil {
		m.Rules = make(map[string]Rule)
	}
	m.Rules[key] = rule
	return m
}

// WithTarget adds a target for a rule
func (m *Module) WithTarget(ruleKey string, target Target) *Module {
	if m.Targets == nil {
		m.Targets = make(map[string][]Target)
	}
	m.Targets[ruleKey] = append(m.Targets[ruleKey], target)
	return m
}

// WithSchedule adds a schedule
func (m *Module) WithSchedule(key string, schedule Schedule) *Module {
	createSchedules := true
	m.CreateSchedules = &createSchedules
	if m.Schedules == nil {
		m.Schedules = make(map[string]Schedule)
	}
	m.Schedules[key] = schedule
	return m
}

// WithPipe adds a pipe
func (m *Module) WithPipe(key string, pipe Pipe) *Module {
	createPipes := true
	m.CreatePipes = &createPipes
	if m.Pipes == nil {
		m.Pipes = make(map[string]Pipe)
	}
	m.Pipes[key] = pipe
	return m
}

// WithTags adds tags to the event bus
func (m *Module) WithTags(tags map[string]string) *Module {
	if m.Tags == nil {
		m.Tags = make(map[string]string)
	}
	for k, v := range tags {
		m.Tags[k] = v
	}
	return m
}

// ================================
// Integration Helper Methods
// ================================

// WithLambdaTarget adds a Lambda function as a target for a rule
func (m *Module) WithLambdaTarget(ruleKey, lambdaARN string) *Module {
	target := Target{
		ARN: lambdaARN,
	}
	return m.WithTarget(ruleKey, target)
}

// WithSQSTarget adds an SQS queue as a target for a rule
func (m *Module) WithSQSTarget(ruleKey, queueARN string) *Module {
	target := Target{
		ARN: queueARN,
	}
	return m.WithTarget(ruleKey, target)
}

// WithSNSTarget adds an SNS topic as a target for a rule
func (m *Module) WithSNSTarget(ruleKey, topicARN string) *Module {
	target := Target{
		ARN: topicARN,
	}
	return m.WithTarget(ruleKey, target)
}

// WithStepFunctionsTarget adds a Step Functions state machine as a target for a rule
func (m *Module) WithStepFunctionsTarget(ruleKey, stateMachineARN string) *Module {
	target := Target{
		ARN: stateMachineARN,
	}
	return m.WithTarget(ruleKey, target)
}

// WithKinesisTarget adds a Kinesis stream as a target for a rule
func (m *Module) WithKinesisTarget(ruleKey, streamARN string) *Module {
	target := Target{
		ARN: streamARN,
	}
	return m.WithTarget(ruleKey, target)
}

// WithECSTarget adds an ECS task as a target for a rule
func (m *Module) WithECSTarget(ruleKey, clusterARN, taskDefinitionARN string, subnets []string) *Module {
	target := Target{
		ARN: clusterARN,
		ECSTarget: &ECSTarget{
			TaskDefinitionARN: taskDefinitionARN,
			NetworkConfiguration: &NetworkConfiguration{
				AWSVPCConfiguration: &AWSVPCConfiguration{
					Subnets: subnets,
				},
			},
		},
	}
	return m.WithTarget(ruleKey, target)
}

// WithEventPattern creates a rule with an event pattern
func (m *Module) WithEventPatternRule(key, description, eventPattern string, enabled bool) *Module {
	rule := Rule{
		Description:  &description,
		EventPattern: &eventPattern,
		Enabled:      &enabled,
	}
	return m.WithRule(key, rule)
}

// WithScheduleRule creates a rule with a schedule expression
func (m *Module) WithScheduleRule(key, description, scheduleExpression string, enabled bool) *Module {
	rule := Rule{
		Description:        &description,
		ScheduleExpression: &scheduleExpression,
		Enabled:            &enabled,
	}
	return m.WithRule(key, rule)
}

// WithAPIDestinationTarget adds an API destination as a target for a rule
func (m *Module) WithAPIDestinationTarget(ruleKey, destinationARN string) *Module {
	target := Target{
		ARN: destinationARN,
	}
	return m.WithTarget(ruleKey, target)
}

// LocalName returns the local identifier for this module instance
func (m *Module) LocalName() string {
	if m.BusName != nil {
		return *m.BusName
	}
	return "eventbridge"
}

// Configuration generates the HCL configuration for this module
func (m *Module) Configuration() (string, error) {
	// TODO: Implement full HCL generation using hclwrite or lingon's marshaling
	return "", nil
}
