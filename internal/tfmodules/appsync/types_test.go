package appsync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewModule tests module creation with defaults.
func TestNewModule(t *testing.T) {
	t.Run("creates module with sensible defaults", func(t *testing.T) {
		name := "my-graphql-api"
		m := NewModule(name)

		require.NotNil(t, m)
		assert.Equal(t, "terraform-aws-modules/appsync/aws", m.Source)
		assert.Equal(t, "~> 2.0", m.Version)
		assert.Equal(t, name, *m.Name)
		assert.True(t, *m.CreateGraphQLAPI)
		assert.Equal(t, "API_KEY", *m.AuthenticationType)
		assert.True(t, *m.CreateLogsRole)
		assert.Equal(t, "FULL_REQUEST_CACHING", *m.CachingBehavior)
		assert.Equal(t, "SMALL", *m.CacheType)
	})

	t.Run("name is set correctly", func(t *testing.T) {
		tests := []struct {
			name     string
			expected string
		}{
			{"api", "api"},
			{"my-graphql-api", "my-graphql-api"},
			{"production-api", "production-api"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				m := NewModule(tt.name)
				assert.Equal(t, tt.expected, *m.Name)
			})
		}
	})
}

// TestModuleWithSchema tests schema configuration.
func TestModuleWithSchema(t *testing.T) {
	t.Run("sets schema", func(t *testing.T) {
		m := NewModule("api")
		schema := `
			type Query {
				hello: String
			}
		`

		result := m.WithSchema(schema)

		assert.Equal(t, m, result, "should return same instance for chaining")
		require.NotNil(t, m.Schema)
		assert.Equal(t, schema, *m.Schema)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		m := NewModule("api").
			WithSchema("type Query { hello: String }")

		require.NotNil(t, m.Schema)
		assert.Contains(t, *m.Schema, "Query")
	})
}

// TestModuleAuthConfiguration tests authentication methods.
func TestModuleAuthConfiguration(t *testing.T) {
	t.Run("WithCognitoAuth configures Cognito", func(t *testing.T) {
		m := NewModule("api")
		userPoolID := "us-east-1_ABC123"
		region := "us-east-1"

		result := m.WithCognitoAuth(userPoolID, region)

		assert.Equal(t, m, result, "should return same instance")
		assert.Equal(t, "AMAZON_COGNITO_USER_POOLS", *m.AuthenticationType)
		require.NotNil(t, m.UserPoolConfig)
		assert.Equal(t, userPoolID, m.UserPoolConfig["user_pool_id"])
		assert.Equal(t, region, m.UserPoolConfig["aws_region"])
		assert.Equal(t, "ALLOW", m.UserPoolConfig["default_action"])
	})

	t.Run("WithIAMAuth configures IAM", func(t *testing.T) {
		m := NewModule("api")

		result := m.WithIAMAuth()

		assert.Equal(t, m, result, "should return same instance")
		assert.Equal(t, "AWS_IAM", *m.AuthenticationType)
	})

	t.Run("WithLambdaAuth configures Lambda authorizer", func(t *testing.T) {
		m := NewModule("api")
		authURI := "arn:aws:lambda:us-east-1:123456789012:function:authorizer"
		ttl := 300

		result := m.WithLambdaAuth(authURI, ttl)

		assert.Equal(t, m, result, "should return same instance")
		assert.Equal(t, "AWS_LAMBDA", *m.AuthenticationType)
		require.NotNil(t, m.LambdaAuthorizerConfig)
		assert.Equal(t, authURI, m.LambdaAuthorizerConfig["authorizer_uri"])
		assert.Equal(t, "300", m.LambdaAuthorizerConfig["authorizer_result_ttl_in_seconds"])
	})

	t.Run("auth methods can override each other", func(t *testing.T) {
		m := NewModule("api").
			WithCognitoAuth("pool-1", "us-east-1").
			WithIAMAuth()

		// Last auth method wins
		assert.Equal(t, "AWS_IAM", *m.AuthenticationType)
	})
}

// TestModuleWithLogging tests logging configuration.
func TestModuleWithLogging(t *testing.T) {
	t.Run("enables logging with specified level", func(t *testing.T) {
		m := NewModule("api")

		result := m.WithLogging("ALL", true)

		assert.Equal(t, m, result, "should return same instance")
		assert.True(t, *m.LoggingEnabled)
		assert.Equal(t, "ALL", *m.LogFieldLogLevel)
		assert.True(t, *m.LogExcludeVerboseContent)
	})

	t.Run("supports different log levels", func(t *testing.T) {
		tests := []struct {
			level           string
			excludeVerbose  bool
		}{
			{"ALL", true},
			{"ERROR", false},
			{"NONE", true},
		}

		for _, tt := range tests {
			t.Run(tt.level, func(t *testing.T) {
				m := NewModule("api").WithLogging(tt.level, tt.excludeVerbose)

				assert.Equal(t, tt.level, *m.LogFieldLogLevel)
				assert.Equal(t, tt.excludeVerbose, *m.LogExcludeVerboseContent)
			})
		}
	})
}

// TestModuleWithXRayTracing tests X-Ray configuration.
func TestModuleWithXRayTracing(t *testing.T) {
	t.Run("enables X-Ray tracing", func(t *testing.T) {
		m := NewModule("api")

		result := m.WithXRayTracing()

		assert.Equal(t, m, result, "should return same instance")
		assert.True(t, *m.XRayEnabled)
	})
}

// TestModuleWithCaching tests caching configuration.
func TestModuleWithCaching(t *testing.T) {
	t.Run("configures caching with all parameters", func(t *testing.T) {
		m := NewModule("api")
		cacheType := "MEDIUM"
		ttl := 3600
		atRest := true
		transit := true

		result := m.WithCaching(cacheType, ttl, atRest, transit)

		assert.Equal(t, m, result, "should return same instance")
		assert.True(t, *m.CachingEnabled)
		assert.Equal(t, cacheType, *m.CacheType)
		assert.Equal(t, ttl, *m.CacheTTL)
		assert.True(t, *m.CacheAtRestEncryptionEnabled)
		assert.True(t, *m.CacheTransitEncryptionEnabled)
	})

	t.Run("supports different cache types", func(t *testing.T) {
		types := []string{"SMALL", "MEDIUM", "LARGE", "XLARGE"}

		for _, cacheType := range types {
			t.Run(cacheType, func(t *testing.T) {
				m := NewModule("api").WithCaching(cacheType, 600, false, false)

				assert.Equal(t, cacheType, *m.CacheType)
			})
		}
	})
}

// TestModuleWithDomainName tests domain name configuration.
func TestModuleWithDomainName(t *testing.T) {
	t.Run("configures custom domain", func(t *testing.T) {
		m := NewModule("api")
		domain := "api.example.com"
		certARN := "arn:aws:acm:us-east-1:123456789012:certificate/abc123"

		result := m.WithDomainName(domain, certARN)

		assert.Equal(t, m, result, "should return same instance")
		assert.True(t, *m.DomainNameAssociationEnabled)
		assert.Equal(t, domain, *m.DomainName)
		assert.Equal(t, certARN, *m.CertificateARN)
	})
}

// TestModuleWithAPIKey tests API key configuration.
func TestModuleWithAPIKey(t *testing.T) {
	t.Run("adds single API key", func(t *testing.T) {
		m := NewModule("api")
		name := "default"
		description := "Default API key"

		result := m.WithAPIKey(name, description)

		assert.Equal(t, m, result, "should return same instance")
		require.NotNil(t, m.APIKeys)
		assert.Equal(t, description, m.APIKeys[name])
	})

	t.Run("adds multiple API keys", func(t *testing.T) {
		m := NewModule("api").
			WithAPIKey("dev", "Development key").
			WithAPIKey("prod", "Production key")

		require.NotNil(t, m.APIKeys)
		assert.Len(t, m.APIKeys, 2)
		assert.Equal(t, "Development key", m.APIKeys["dev"])
		assert.Equal(t, "Production key", m.APIKeys["prod"])
	})

	t.Run("initializes APIKeys map if nil", func(t *testing.T) {
		m := &Module{}
		assert.Nil(t, m.APIKeys)

		m.WithAPIKey("key1", "First key")

		assert.NotNil(t, m.APIKeys)
	})
}

// TestModuleWithDataSource tests data source configuration.
func TestModuleWithDataSource(t *testing.T) {
	t.Run("adds generic data source", func(t *testing.T) {
		m := NewModule("api")
		name := "custom"
		ds := DataSource{
			Type: "HTTP",
			HTTPConfig: &HTTPConfig{
				Endpoint: "https://api.example.com",
			},
		}

		result := m.WithDataSource(name, ds)

		assert.Equal(t, m, result, "should return same instance")
		require.NotNil(t, m.DataSources)
		assert.Equal(t, ds, m.DataSources[name])
	})

	t.Run("initializes DataSources map if nil", func(t *testing.T) {
		m := &Module{}
		assert.Nil(t, m.DataSources)

		m.WithDataSource("ds1", DataSource{Type: "NONE"})

		assert.NotNil(t, m.DataSources)
	})
}

// TestModuleWithLambdaDataSource tests Lambda data source helper.
func TestModuleWithLambdaDataSource(t *testing.T) {
	t.Run("adds Lambda data source", func(t *testing.T) {
		m := NewModule("api")
		name := "handler"
		arn := "arn:aws:lambda:us-east-1:123456789012:function:handler"

		result := m.WithLambdaDataSource(name, arn)

		assert.Equal(t, m, result, "should return same instance")
		require.NotNil(t, m.DataSources)

		ds := m.DataSources[name]
		assert.Equal(t, "AWS_LAMBDA", ds.Type)
		require.NotNil(t, ds.LambdaConfig)
		assert.Equal(t, arn, ds.LambdaConfig.FunctionARN)
	})

	t.Run("supports method chaining with multiple data sources", func(t *testing.T) {
		m := NewModule("api").
			WithLambdaDataSource("func1", "arn:1").
			WithLambdaDataSource("func2", "arn:2")

		assert.Len(t, m.DataSources, 2)
	})
}

// TestModuleWithDynamoDBDataSource tests DynamoDB data source helper.
func TestModuleWithDynamoDBDataSource(t *testing.T) {
	t.Run("adds DynamoDB data source", func(t *testing.T) {
		m := NewModule("api")
		name := "users_table"
		tableName := "users"

		result := m.WithDynamoDBDataSource(name, tableName)

		assert.Equal(t, m, result, "should return same instance")
		require.NotNil(t, m.DataSources)

		ds := m.DataSources[name]
		assert.Equal(t, "AMAZON_DYNAMODB", ds.Type)
		require.NotNil(t, ds.DynamoDBConfig)
		assert.Equal(t, tableName, ds.DynamoDBConfig.TableName)
	})
}

// TestModuleWithResolver tests resolver configuration.
func TestModuleWithResolver(t *testing.T) {
	t.Run("adds resolver", func(t *testing.T) {
		m := NewModule("api")
		name := "get_user"
		resolver := Resolver{
			Type:  "Query",
			Field: "getUser",
		}
		dataSource := "users_table"
		resolver.DataSource = &dataSource

		result := m.WithResolver(name, resolver)

		assert.Equal(t, m, result, "should return same instance")
		require.NotNil(t, m.Resolvers)
		assert.Equal(t, resolver, m.Resolvers[name])
	})

	t.Run("adds multiple resolvers", func(t *testing.T) {
		m := NewModule("api").
			WithResolver("query1", Resolver{Type: "Query", Field: "query1"}).
			WithResolver("mutation1", Resolver{Type: "Mutation", Field: "mutation1"})

		assert.Len(t, m.Resolvers, 2)
	})

	t.Run("initializes Resolvers map if nil", func(t *testing.T) {
		m := &Module{}
		assert.Nil(t, m.Resolvers)

		m.WithResolver("r1", Resolver{Type: "Query", Field: "test"})

		assert.NotNil(t, m.Resolvers)
	})
}

// TestModuleWithFunction tests pipeline function configuration.
func TestModuleWithFunction(t *testing.T) {
	t.Run("adds pipeline function", func(t *testing.T) {
		m := NewModule("api")
		name := "transform"
		fn := Function{
			DataSource: "lambda_ds",
		}
		desc := "Transform data"
		fn.Description = &desc

		result := m.WithFunction(name, fn)

		assert.Equal(t, m, result, "should return same instance")
		require.NotNil(t, m.Functions)
		assert.Equal(t, fn, m.Functions[name])
	})

	t.Run("initializes Functions map if nil", func(t *testing.T) {
		m := &Module{}
		assert.Nil(t, m.Functions)

		m.WithFunction("f1", Function{DataSource: "ds1"})

		assert.NotNil(t, m.Functions)
	})
}

// TestModuleWithTags tests tag configuration.
func TestModuleWithTags(t *testing.T) {
	t.Run("adds tags", func(t *testing.T) {
		m := NewModule("api")
		tags := map[string]string{
			"Environment": "production",
			"Team":        "platform",
		}

		result := m.WithTags(tags)

		assert.Equal(t, m, result, "should return same instance")
		require.NotNil(t, m.Tags)
		assert.Equal(t, "production", m.Tags["Environment"])
		assert.Equal(t, "platform", m.Tags["Team"])
	})

	t.Run("merges tags on multiple calls", func(t *testing.T) {
		m := NewModule("api").
			WithTags(map[string]string{"Key1": "Value1"}).
			WithTags(map[string]string{"Key2": "Value2"})

		assert.Len(t, m.Tags, 2)
		assert.Equal(t, "Value1", m.Tags["Key1"])
		assert.Equal(t, "Value2", m.Tags["Key2"])
	})

	t.Run("initializes Tags map if nil", func(t *testing.T) {
		m := &Module{}
		assert.Nil(t, m.Tags)

		m.WithTags(map[string]string{"Key": "Value"})

		assert.NotNil(t, m.Tags)
	})

	t.Run("overwrites existing tags with same key", func(t *testing.T) {
		m := NewModule("api").
			WithTags(map[string]string{"Env": "dev"}).
			WithTags(map[string]string{"Env": "prod"})

		assert.Equal(t, "prod", m.Tags["Env"])
	})
}

// TestModuleLocalName tests local name generation.
func TestModuleLocalName(t *testing.T) {
	t.Run("returns name if set", func(t *testing.T) {
		name := "my-custom-api"
		m := NewModule(name)

		assert.Equal(t, name, m.LocalName())
	})

	t.Run("returns default name if not set", func(t *testing.T) {
		m := &Module{}

		assert.Equal(t, "graphql_api", m.LocalName())
	})
}

// TestModuleConfiguration tests HCL configuration generation.
func TestModuleConfiguration(t *testing.T) {
	t.Run("returns empty string currently", func(t *testing.T) {
		m := NewModule("api")

		config, err := m.Configuration()

		assert.NoError(t, err)
		assert.Empty(t, config)
	})
}

// TestModuleFluentBuilder tests method chaining.
func TestModuleFluentBuilder(t *testing.T) {
	t.Run("supports complex fluent builder pattern", func(t *testing.T) {
		m := NewModule("production-api").
			WithSchema("type Query { hello: String }").
			WithIAMAuth().
			WithLogging("ALL", true).
			WithXRayTracing().
			WithCaching("MEDIUM", 3600, true, true).
			WithAPIKey("default", "Default key").
			WithLambdaDataSource("handler", "arn:aws:lambda:us-east-1:123456789012:function:handler").
			WithTags(map[string]string{"Environment": "production"})

		// Verify all configurations were applied
		assert.Equal(t, "production-api", *m.Name)
		assert.NotNil(t, m.Schema)
		assert.Equal(t, "AWS_IAM", *m.AuthenticationType)
		assert.True(t, *m.LoggingEnabled)
		assert.True(t, *m.XRayEnabled)
		assert.True(t, *m.CachingEnabled)
		assert.Contains(t, m.APIKeys, "default")
		assert.Contains(t, m.DataSources, "handler")
		assert.Equal(t, "production", m.Tags["Environment"])
	})

	t.Run("all With methods return module for chaining", func(t *testing.T) {
		m := NewModule("api")

		// Each method should return the same instance
		assert.Equal(t, m, m.WithSchema("schema"))
		assert.Equal(t, m, m.WithCognitoAuth("pool", "region"))
		assert.Equal(t, m, m.WithIAMAuth())
		assert.Equal(t, m, m.WithLambdaAuth("uri", 300))
		assert.Equal(t, m, m.WithLogging("ALL", true))
		assert.Equal(t, m, m.WithXRayTracing())
		assert.Equal(t, m, m.WithCaching("SMALL", 600, false, false))
		assert.Equal(t, m, m.WithDomainName("domain", "cert"))
		assert.Equal(t, m, m.WithAPIKey("key", "desc"))
		assert.Equal(t, m, m.WithDataSource("ds", DataSource{Type: "NONE"}))
		assert.Equal(t, m, m.WithLambdaDataSource("lambda", "arn"))
		assert.Equal(t, m, m.WithDynamoDBDataSource("ddb", "table"))
		assert.Equal(t, m, m.WithResolver("resolver", Resolver{Type: "Query", Field: "test"}))
		assert.Equal(t, m, m.WithFunction("fn", Function{DataSource: "ds"}))
		assert.Equal(t, m, m.WithTags(map[string]string{}))
	})
}

// TestDataSourceTypes tests data source type creation.
func TestDataSourceTypes(t *testing.T) {
	t.Run("Lambda data source", func(t *testing.T) {
		ds := DataSource{
			Type: "AWS_LAMBDA",
			LambdaConfig: &LambdaConfig{
				FunctionARN: "arn:aws:lambda:us-east-1:123456789012:function:test",
			},
		}

		assert.Equal(t, "AWS_LAMBDA", ds.Type)
		require.NotNil(t, ds.LambdaConfig)
		assert.Contains(t, ds.LambdaConfig.FunctionARN, "function:test")
	})

	t.Run("DynamoDB data source with delta sync", func(t *testing.T) {
		ds := DataSource{
			Type: "AMAZON_DYNAMODB",
			DynamoDBConfig: &DynamoDBConfig{
				TableName: "users",
				DeltaSyncConfig: &DeltaSyncConfig{
					DeltaSyncTableName: "users-delta",
				},
			},
		}

		assert.Equal(t, "AMAZON_DYNAMODB", ds.Type)
		require.NotNil(t, ds.DynamoDBConfig)
		require.NotNil(t, ds.DynamoDBConfig.DeltaSyncConfig)
		assert.Equal(t, "users-delta", ds.DynamoDBConfig.DeltaSyncConfig.DeltaSyncTableName)
	})

	t.Run("HTTP data source with IAM auth", func(t *testing.T) {
		ds := DataSource{
			Type: "HTTP",
			HTTPConfig: &HTTPConfig{
				Endpoint: "https://api.example.com",
				AuthorizationConfig: &AuthorizationConfig{
					AuthorizationType: "AWS_IAM",
					AWSIAMConfig: &AWSIAMConfig{
						SigningRegion:      "us-east-1",
						SigningServiceName: "execute-api",
					},
				},
			},
		}

		assert.Equal(t, "HTTP", ds.Type)
		require.NotNil(t, ds.HTTPConfig)
		require.NotNil(t, ds.HTTPConfig.AuthorizationConfig)
		assert.Equal(t, "AWS_IAM", ds.HTTPConfig.AuthorizationConfig.AuthorizationType)
	})
}

// TestResolverTypes tests resolver configurations.
func TestResolverTypes(t *testing.T) {
	t.Run("unit resolver", func(t *testing.T) {
		kind := "UNIT"
		ds := "users_table"
		resolver := Resolver{
			Type:       "Query",
			Field:      "getUser",
			Kind:       &kind,
			DataSource: &ds,
		}

		assert.Equal(t, "Query", resolver.Type)
		assert.Equal(t, "getUser", resolver.Field)
		assert.Equal(t, "UNIT", *resolver.Kind)
		assert.Equal(t, "users_table", *resolver.DataSource)
	})

	t.Run("pipeline resolver", func(t *testing.T) {
		kind := "PIPELINE"
		resolver := Resolver{
			Type:  "Mutation",
			Field: "createUser",
			Kind:  &kind,
			PipelineConfig: &PipelineConfig{
				Functions: []string{"validate", "transform", "save"},
			},
		}

		assert.Equal(t, "PIPELINE", *resolver.Kind)
		require.NotNil(t, resolver.PipelineConfig)
		assert.Len(t, resolver.PipelineConfig.Functions, 3)
	})

	t.Run("resolver with caching", func(t *testing.T) {
		ttl := 300
		resolver := Resolver{
			Type:  "Query",
			Field: "listUsers",
			CachingConfig: &CachingConfig{
				TTL:         &ttl,
				CachingKeys: []string{"$context.identity.sub"},
			},
		}

		require.NotNil(t, resolver.CachingConfig)
		assert.Equal(t, 300, *resolver.CachingConfig.TTL)
		assert.Len(t, resolver.CachingConfig.CachingKeys, 1)
	})
}

// BenchmarkModuleCreation benchmarks module creation.
func BenchmarkModuleCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewModule("api")
	}
}

// BenchmarkFluentBuilder benchmarks fluent builder pattern.
func BenchmarkFluentBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewModule("api").
			WithSchema("type Query { hello: String }").
			WithIAMAuth().
			WithLogging("ALL", true).
			WithXRayTracing().
			WithLambdaDataSource("handler", "arn").
			WithTags(map[string]string{"Env": "prod"})
	}
}
