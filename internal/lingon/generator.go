package lingon

import (
	"context"
	"fmt"

	E "github.com/IBM/fp-go/either"
	O "github.com/IBM/fp-go/option"
	"github.com/samber/lo"
)

// GeneratorFunc is a function that generates Terraform code from ForgeConfig
type GeneratorFunc func(ctx context.Context, config ForgeConfig) E.Either[error, []byte]

// Generator provides Terraform generation from Forge configuration
type Generator struct {
	Generate GeneratorFunc
}

// NewGenerator creates a new Terraform generator using Lingon
func NewGenerator() Generator {
	return Generator{
		Generate: makeGenerateFunc(),
	}
}

// makeGenerateFunc creates the main generation function
func makeGenerateFunc() GeneratorFunc {
	return func(ctx context.Context, config ForgeConfig) E.Either[error, []byte] {
		// Validate configuration
		if err := validateConfig(config); err != nil {
			return E.Left[[]byte](fmt.Errorf("invalid configuration: %w", err))
		}

		// Generate Lingon stack
		stack, err := generateStack(config)
		if err != nil {
			return E.Left[[]byte](fmt.Errorf("failed to generate stack: %w", err))
		}

		// Export to Terraform
		terraformCode, err := exportToTerraform(stack)
		if err != nil {
			return E.Left[[]byte](fmt.Errorf("failed to export terraform: %w", err))
		}

		return E.Right[error](terraformCode)
	}
}

// validateConfig validates the ForgeConfig
func validateConfig(config ForgeConfig) error {
	if config.Service == "" {
		return fmt.Errorf("service name is required")
	}

	if config.Provider.Region == "" {
		return fmt.Errorf("provider region is required")
	}

	if len(config.Functions) == 0 {
		return fmt.Errorf("at least one function is required")
	}

	// Validate each function
	for name, fn := range config.Functions {
		if err := validateFunction(name, fn); err != nil {
			return fmt.Errorf("function %s: %w", name, err)
		}
	}

	return nil
}

// validateFunction validates a single function configuration
func validateFunction(name string, fn FunctionConfig) error {
	if fn.Handler == "" {
		return fmt.Errorf("handler is required")
	}

	if fn.Runtime == "" {
		return fmt.Errorf("runtime is required")
	}

	if fn.Source.Path == "" && fn.Source.S3Bucket == "" && fn.Source.Filename == "" {
		return fmt.Errorf("source path, S3 location, or filename is required")
	}

	// Validate runtime
	validRuntimes := []string{
		"nodejs18.x", "nodejs20.x",
		"python3.9", "python3.10", "python3.11", "python3.12",
		"go1.x",
		"java11", "java17", "java21",
		"dotnet6", "dotnet7", "dotnet8",
		"ruby3.2", "ruby3.3",
		"provided.al2", "provided.al2023",
	}

	if !lo.Contains(validRuntimes, fn.Runtime) {
		return fmt.Errorf("unsupported runtime: %s", fn.Runtime)
	}

	return nil
}

// Stack represents a Lingon stack containing all resources
type Stack struct {
	Service   string
	Provider  ProviderConfig
	Functions map[string]*LambdaFunction
	APIGateway O.Option[*APIGateway]
	Tables    map[string]*DynamoDBTable
	EventBridgeRules map[string]*EventBridgeRule
	StateMachines map[string]*StepFunctionsStateMachine
	Topics    map[string]*SNSTopic
	Queues    map[string]*SQSQueue
	Buckets   map[string]*S3Bucket
	Alarms    map[string]*CloudWatchAlarm
}

// LambdaFunction represents a Lingon Lambda function resource
type LambdaFunction struct {
	Name           string
	Config         FunctionConfig
	Role           *IAMRole
	LogGroup       O.Option[*CloudWatchLogGroup]
	FunctionURL    O.Option[*LambdaFunctionURL]
	EventSources   []EventSourceMapping
}

// IAMRole represents a Lingon IAM role
type IAMRole struct {
	Name                 string
	AssumeRolePolicy     string
	ManagedPolicyArns    []string
	InlinePolicies       []InlinePolicy
	PermissionsBoundary  O.Option[string]
	MaxSessionDuration   O.Option[int]
}

// CloudWatchLogGroup represents a CloudWatch log group
type CloudWatchLogGroup struct {
	Name              string
	RetentionInDays   O.Option[int]
	KMSKeyId          O.Option[string]
}

// LambdaFunctionURL represents a Lambda function URL
type LambdaFunctionURL struct {
	FunctionName      string
	AuthorizationType string
	CORS              O.Option[CORSConfig]
	InvokeMode        O.Option[string]
}

// EventSourceMapping represents a Lambda event source mapping
type EventSourceMapping struct {
	FunctionName      string
	EventSourceArn    string
	Config            EventSourceMappingConfig
}

// APIGateway represents a Lingon API Gateway
type APIGateway struct {
	Name        string
	Config      APIGatewayConfig
	Integrations map[string]*APIGatewayIntegration
	Routes      map[string]*APIGatewayRoute
	Stages      map[string]*APIGatewayStage
	Domain      O.Option[*APIGatewayDomain]
}

// APIGatewayIntegration represents an API Gateway integration
type APIGatewayIntegration struct {
	APIId             string
	IntegrationType   string
	IntegrationURI    string
	PayloadFormatVersion string
}

// APIGatewayRoute represents an API Gateway route
type APIGatewayRoute struct {
	APIId             string
	RouteKey          string
	Target            string
	AuthorizationType O.Option[string]
	AuthorizerId      O.Option[string]
}

// APIGatewayStage represents an API Gateway stage
type APIGatewayStage struct {
	APIId      string
	Name       string
	Config     StageConfig
}

// APIGatewayDomain represents an API Gateway custom domain
type APIGatewayDomain struct {
	DomainName     string
	CertificateArn string
	Config         DomainConfig
}

// DynamoDBTable represents a Lingon DynamoDB table
type DynamoDBTable struct {
	Name   string
	Config TableConfig
}

// EventBridgeRule represents an EventBridge rule
type EventBridgeRule struct {
	Name   string
	Config EventBridgeConfig
}

// StepFunctionsStateMachine represents a Step Functions state machine
type StepFunctionsStateMachine struct {
	Name   string
	Config StateMachineConfig
}

// SNSTopic represents an SNS topic
type SNSTopic struct {
	Name   string
	Config TopicConfig
}

// SQSQueue represents an SQS queue
type SQSQueue struct {
	Name   string
	Config QueueConfig
}

// S3Bucket represents an S3 bucket
type S3Bucket struct {
	Name   string
	Config BucketConfig
}

// CloudWatchAlarm represents a CloudWatch alarm
type CloudWatchAlarm struct {
	Name   string
	Config AlarmConfig
}

// generateStack generates a Lingon stack from ForgeConfig
func generateStack(config ForgeConfig) (*Stack, error) {
	stack := &Stack{
		Service:   config.Service,
		Provider:  config.Provider,
		Functions: make(map[string]*LambdaFunction),
		Tables:    make(map[string]*DynamoDBTable),
		EventBridgeRules: make(map[string]*EventBridgeRule),
		StateMachines: make(map[string]*StepFunctionsStateMachine),
		Topics:    make(map[string]*SNSTopic),
		Queues:    make(map[string]*SQSQueue),
		Buckets:   make(map[string]*S3Bucket),
		Alarms:    make(map[string]*CloudWatchAlarm),
	}

	// Generate Lambda functions
	for name, fnConfig := range config.Functions {
		// Validate function first
		if err := validateFunction(name, fnConfig); err != nil {
			return nil, fmt.Errorf("function %s: %w", name, err)
		}

		lambdaFn, err := generateLambdaFunction(config.Service, name, fnConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to generate function %s: %w", name, err)
		}
		stack.Functions[name] = lambdaFn
	}

	// Generate API Gateway if configured
	if config.APIGateway != nil {
		apiGateway, err := generateAPIGateway(config.Service, *config.APIGateway, config.Functions)
		if err != nil {
			return nil, fmt.Errorf("failed to generate API Gateway: %w", err)
		}
		stack.APIGateway = O.Some(apiGateway)
	} else {
		stack.APIGateway = O.None[*APIGateway]()
	}

	// Generate DynamoDB tables
	for name, tableConfig := range config.Tables {
		table := generateDynamoDBTable(config.Service, name, tableConfig)
		stack.Tables[name] = table
	}

	// Generate EventBridge rules
	for name, eventConfig := range config.EventBridge {
		rule := generateEventBridgeRule(config.Service, name, eventConfig)
		stack.EventBridgeRules[name] = rule
	}

	// Generate Step Functions state machines
	for name, smConfig := range config.StateMachines {
		stateMachine := generateStateMachine(config.Service, name, smConfig)
		stack.StateMachines[name] = stateMachine
	}

	// Generate SNS topics
	for name, topicConfig := range config.Topics {
		topic := generateSNSTopic(config.Service, name, topicConfig)
		stack.Topics[name] = topic
	}

	// Generate SQS queues
	for name, queueConfig := range config.Queues {
		queue := generateSQSQueue(config.Service, name, queueConfig)
		stack.Queues[name] = queue
	}

	// Generate S3 buckets
	for name, bucketConfig := range config.Buckets {
		bucket := generateS3Bucket(config.Service, name, bucketConfig)
		stack.Buckets[name] = bucket
	}

	// Generate CloudWatch alarms
	for name, alarmConfig := range config.Alarms {
		alarm := generateCloudWatchAlarm(config.Service, name, alarmConfig)
		stack.Alarms[name] = alarm
	}

	return stack, nil
}

// generateLambdaFunction generates a Lambda function resource
func generateLambdaFunction(service, name string, config FunctionConfig) (*LambdaFunction, error) {
	// Generate IAM role
	role := generateIAMRole(service, name, config.IAM)

	// Generate CloudWatch log group
	logGroup := O.None[*CloudWatchLogGroup]()
	if config.Logs.RetentionInDays > 0 || config.Logs.LogGroupName != "" {
		retentionOpt := O.None[int]()
		if config.Logs.RetentionInDays > 0 {
			retentionOpt = O.Some(config.Logs.RetentionInDays)
		}

		kmsKeyOpt := O.None[string]()
		if config.Logs.KMSKeyId != "" {
			kmsKeyOpt = O.Some(config.Logs.KMSKeyId)
		}

		lg := &CloudWatchLogGroup{
			Name:            lo.Ternary(config.Logs.LogGroupName != "", config.Logs.LogGroupName, fmt.Sprintf("/aws/lambda/%s-%s", service, name)),
			RetentionInDays: retentionOpt,
			KMSKeyId:        kmsKeyOpt,
		}
		logGroup = O.Some(lg)
	}

	// Generate function URL if configured
	functionURL := O.None[*LambdaFunctionURL]()
	if config.FunctionURL != nil {
		corsOpt := O.None[CORSConfig]()
		if config.FunctionURL.CORS != nil {
			corsOpt = O.Some(*config.FunctionURL.CORS)
		}

		invokeModeOpt := O.None[string]()
		if config.FunctionURL.InvokeMode != "" {
			invokeModeOpt = O.Some(config.FunctionURL.InvokeMode)
		}

		furl := &LambdaFunctionURL{
			FunctionName:      fmt.Sprintf("%s-%s", service, name),
			AuthorizationType: config.FunctionURL.AuthorizationType,
			CORS:              corsOpt,
			InvokeMode:        invokeModeOpt,
		}
		functionURL = O.Some(furl)
	}

	// Generate event source mappings
	eventSources := lo.Map(config.EventSourceMappings, func(esm EventSourceMappingConfig, _ int) EventSourceMapping {
		return EventSourceMapping{
			FunctionName:   fmt.Sprintf("%s-%s", service, name),
			EventSourceArn: esm.EventSourceArn,
			Config:         esm,
		}
	})

	return &LambdaFunction{
		Name:         fmt.Sprintf("%s-%s", service, name),
		Config:       config,
		Role:         role,
		LogGroup:     logGroup,
		FunctionURL:  functionURL,
		EventSources: eventSources,
	}, nil
}

// generateIAMRole generates an IAM role for a Lambda function
func generateIAMRole(service, functionName string, config IAMConfig) *IAMRole {
	// Default assume role policy for Lambda
	defaultAssumeRolePolicy := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"Service": "lambda.amazonaws.com"},
			"Action": "sts:AssumeRole"
		}]
	}`

	assumeRolePolicy := lo.Ternary(
		config.AssumeRolePolicy != "",
		config.AssumeRolePolicy,
		defaultAssumeRolePolicy,
	)

	// Add basic Lambda execution role if no policies specified
	managedPolicies := config.ManagedPolicyArns
	if len(managedPolicies) == 0 {
		managedPolicies = []string{
			"arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole",
		}
	}

	roleName := lo.Ternary(
		config.RoleName != "",
		config.RoleName,
		fmt.Sprintf("%s-%s-role", service, functionName),
	)

	permissionsBoundaryOpt := O.None[string]()
	if config.PermissionsBoundary != "" {
		permissionsBoundaryOpt = O.Some(config.PermissionsBoundary)
	}

	maxSessionDurationOpt := O.None[int]()
	if config.MaxSessionDuration > 0 {
		maxSessionDurationOpt = O.Some(config.MaxSessionDuration)
	}

	return &IAMRole{
		Name:                roleName,
		AssumeRolePolicy:    assumeRolePolicy,
		ManagedPolicyArns:   managedPolicies,
		InlinePolicies:      config.InlinePolicies,
		PermissionsBoundary: permissionsBoundaryOpt,
		MaxSessionDuration:  maxSessionDurationOpt,
	}
}

// generateAPIGateway generates an API Gateway resource
func generateAPIGateway(service string, config APIGatewayConfig, functions map[string]FunctionConfig) (*APIGateway, error) {
	apiName := lo.Ternary(config.Name != "", config.Name, fmt.Sprintf("%s-api", service))

	api := &APIGateway{
		Name:         apiName,
		Config:       config,
		Integrations: make(map[string]*APIGatewayIntegration),
		Routes:       make(map[string]*APIGatewayRoute),
		Stages:       make(map[string]*APIGatewayStage),
	}

	// Generate integrations and routes for functions with HTTP routing
	for fnName, fnConfig := range functions {
		if fnConfig.HTTPRouting != nil {
			integrationName := fmt.Sprintf("%s-integration", fnName)
			routeName := fmt.Sprintf("%s-route", fnName)

			// Create integration
			integration := &APIGatewayIntegration{
				APIId:                apiName,
				IntegrationType:      "AWS_PROXY",
				IntegrationURI:       fmt.Sprintf("arn:aws:lambda:${var.region}:${var.account_id}:function:%s-%s", service, fnName),
				PayloadFormatVersion: "2.0",
			}
			api.Integrations[integrationName] = integration

			// Create route
			routeKey := fmt.Sprintf("%s %s", fnConfig.HTTPRouting.Method, fnConfig.HTTPRouting.Path)

			authTypeOpt := O.None[string]()
			if fnConfig.HTTPRouting.AuthorizationType != "" {
				authTypeOpt = O.Some(fnConfig.HTTPRouting.AuthorizationType)
			}

			authorizerIdOpt := O.None[string]()
			if fnConfig.HTTPRouting.AuthorizerId != "" {
				authorizerIdOpt = O.Some(fnConfig.HTTPRouting.AuthorizerId)
			}

			route := &APIGatewayRoute{
				APIId:             apiName,
				RouteKey:          routeKey,
				Target:            fmt.Sprintf("integrations/${%s.id}", integrationName),
				AuthorizationType: authTypeOpt,
				AuthorizerId:      authorizerIdOpt,
			}
			api.Routes[routeName] = route
		}
	}

	// Generate stages
	if len(config.Stages) > 0 {
		for stageName, stageConfig := range config.Stages {
			stage := &APIGatewayStage{
				APIId:  apiName,
				Name:   stageName,
				Config: stageConfig,
			}
			api.Stages[stageName] = stage
		}
	} else {
		// Default stage
		defaultStage := &APIGatewayStage{
			APIId: apiName,
			Name:  "$default",
			Config: StageConfig{
				Name:       "$default",
				AutoDeploy: true,
			},
		}
		api.Stages["default"] = defaultStage
	}

	// Generate custom domain if configured
	if config.Domain != nil {
		domain := &APIGatewayDomain{
			DomainName:     config.Domain.DomainName,
			CertificateArn: config.Domain.CertificateArn,
			Config:         *config.Domain,
		}
		api.Domain = O.Some(domain)
	} else {
		api.Domain = O.None[*APIGatewayDomain]()
	}

	return api, nil
}

// generateDynamoDBTable generates a DynamoDB table resource
func generateDynamoDBTable(service, name string, config TableConfig) *DynamoDBTable {
	tableName := lo.Ternary(config.TableName != "", config.TableName, fmt.Sprintf("%s-%s", service, name))

	tableConfig := config
	tableConfig.TableName = tableName

	return &DynamoDBTable{
		Name:   tableName,
		Config: tableConfig,
	}
}

// generateEventBridgeRule generates an EventBridge rule resource
func generateEventBridgeRule(service, name string, config EventBridgeConfig) *EventBridgeRule {
	ruleName := lo.Ternary(config.Name != "", config.Name, fmt.Sprintf("%s-%s", service, name))

	ruleConfig := config
	ruleConfig.Name = ruleName

	return &EventBridgeRule{
		Name:   ruleName,
		Config: ruleConfig,
	}
}

// generateStateMachine generates a Step Functions state machine resource
func generateStateMachine(service, name string, config StateMachineConfig) *StepFunctionsStateMachine {
	smName := lo.Ternary(config.Name != "", config.Name, fmt.Sprintf("%s-%s", service, name))

	smConfig := config
	smConfig.Name = smName

	return &StepFunctionsStateMachine{
		Name:   smName,
		Config: smConfig,
	}
}

// generateSNSTopic generates an SNS topic resource
func generateSNSTopic(service, name string, config TopicConfig) *SNSTopic {
	topicName := lo.Ternary(config.Name != "", config.Name, fmt.Sprintf("%s-%s", service, name))

	topicConfig := config
	topicConfig.Name = topicName

	return &SNSTopic{
		Name:   topicName,
		Config: topicConfig,
	}
}

// generateSQSQueue generates an SQS queue resource
func generateSQSQueue(service, name string, config QueueConfig) *SQSQueue {
	queueName := lo.Ternary(config.Name != "", config.Name, fmt.Sprintf("%s-%s", service, name))

	queueConfig := config
	queueConfig.Name = queueName

	return &SQSQueue{
		Name:   queueName,
		Config: queueConfig,
	}
}

// generateS3Bucket generates an S3 bucket resource
func generateS3Bucket(service, name string, config BucketConfig) *S3Bucket {
	bucketName := lo.Ternary(config.Name != "", config.Name, fmt.Sprintf("%s-%s", service, name))

	bucketConfig := config
	bucketConfig.Name = bucketName

	return &S3Bucket{
		Name:   bucketName,
		Config: bucketConfig,
	}
}

// generateCloudWatchAlarm generates a CloudWatch alarm resource
func generateCloudWatchAlarm(service, name string, config AlarmConfig) *CloudWatchAlarm {
	alarmName := lo.Ternary(config.Name != "", config.Name, fmt.Sprintf("%s-%s", service, name))

	alarmConfig := config
	alarmConfig.Name = alarmName

	return &CloudWatchAlarm{
		Name:   alarmName,
		Config: alarmConfig,
	}
}

// exportToTerraform exports the stack to Terraform code using Lingon
func exportToTerraform(stack *Stack) ([]byte, error) {
	// Convert our internal Stack to ForgeConfig for Lingon
	config := ForgeConfig{
		Service:  stack.Service,
		Provider: stack.Provider,
		Functions: make(map[string]FunctionConfig),
		Tables:    make(map[string]TableConfig),
	}

	// Convert Functions
	for name, fn := range stack.Functions {
		config.Functions[name] = fn.Config
	}

	// Convert API Gateway
	if O.IsSome(stack.APIGateway) {
		apiGateway := O.Fold(
			func() *APIGateway { return nil },
			func(ag *APIGateway) *APIGateway { return ag },
		)(stack.APIGateway)

		if apiGateway != nil {
			config.APIGateway = &apiGateway.Config
		}
	}

	// Convert Tables
	for name, table := range stack.Tables {
		config.Tables[name] = table.Config
	}

	// Create Lingon stack
	lingonStack, err := NewForgeStack(config)
	if err != nil {
		return nil, fmt.Errorf("creating Lingon stack: %w", err)
	}

	// Export to Terraform HCL
	hcl, err := lingonStack.ExportTerraform()
	if err != nil {
		return nil, fmt.Errorf("exporting Terraform: %w", err)
	}

	return hcl, nil
}
