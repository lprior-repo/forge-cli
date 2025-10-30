package apigatewayv2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	t.Run("creates module with sensible defaults", func(t *testing.T) {
		name := "test_api"
		module := NewModule(name)

		require.NotNil(t, module)
		assert.Equal(t, "terraform-aws-modules/apigateway-v2/aws", module.Source)
		assert.Equal(t, "~> 5.0", module.Version)
		assert.NotNil(t, module.Name)
		assert.Equal(t, name, *module.Name)

		// Verify sensible defaults
		assert.NotNil(t, module.Create)
		assert.True(t, *module.Create)

		assert.NotNil(t, module.ProtocolType)
		assert.Equal(t, "HTTP", *module.ProtocolType)

		assert.NotNil(t, module.CreateStage)
		assert.True(t, *module.CreateStage)

		assert.NotNil(t, module.StageName)
		assert.Equal(t, "default", *module.StageName)

		assert.NotNil(t, module.AutoDeploy)
		assert.True(t, *module.AutoDeploy)
	})

	t.Run("creates module with different names", func(t *testing.T) {
		names := []string{"api1", "my-api", "api_gateway"}
		for _, name := range names {
			module := NewModule(name)
			assert.NotNil(t, module.Name)
			assert.Equal(t, name, *module.Name)
		}
	})
}

func TestModule_WithCORS(t *testing.T) {
	t.Run("configures CORS settings", func(t *testing.T) {
		allowOrigins := []string{"https://example.com", "https://app.example.com"}
		allowMethods := []string{"GET", "POST", "PUT"}
		allowHeaders := []string{"Content-Type", "Authorization"}

		module := NewModule("test_api")
		result := module.WithCORS(allowOrigins, allowMethods, allowHeaders)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.CORSConfiguration)
		assert.Equal(t, allowOrigins, module.CORSConfiguration.AllowOrigins)
		assert.Equal(t, allowMethods, module.CORSConfiguration.AllowMethods)
		assert.Equal(t, allowHeaders, module.CORSConfiguration.AllowHeaders)
	})

	t.Run("supports wildcard origins", func(t *testing.T) {
		module := NewModule("test_api")
		module.WithCORS([]string{"*"}, []string{"*"}, []string{"*"})

		assert.Equal(t, []string{"*"}, module.CORSConfiguration.AllowOrigins)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test_api").
			WithCORS([]string{"*"}, []string{"GET"}, []string{"Content-Type"})

		assert.NotNil(t, module.CORSConfiguration)
	})
}

func TestModule_WithDomainName(t *testing.T) {
	t.Run("configures custom domain", func(t *testing.T) {
		domain := "api.example.com"
		certARN := "arn:aws:acm:us-east-1:123456789012:certificate/12345"

		module := NewModule("test_api")
		result := module.WithDomainName(domain, certARN)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.CreateDomainName)
		assert.True(t, *module.CreateDomainName)
		assert.NotNil(t, module.DomainName)
		assert.Equal(t, domain, *module.DomainName)
		assert.NotNil(t, module.DomainNameCertificateARN)
		assert.Equal(t, certARN, *module.DomainNameCertificateARN)
	})

	t.Run("supports wildcard domains", func(t *testing.T) {
		module := NewModule("test_api")
		module.WithDomainName("*.example.com", "cert-arn")

		assert.Equal(t, "*.example.com", *module.DomainName)
	})
}

func TestModule_WithJWTAuthorizer(t *testing.T) {
	t.Run("adds JWT authorizer", func(t *testing.T) {
		name := "cognito-auth"
		issuer := "https://cognito-idp.us-east-1.amazonaws.com/us-east-1_XXXXXX"
		audience := []string{"app-client-id"}

		module := NewModule("test_api")
		result := module.WithJWTAuthorizer(name, issuer, audience)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Authorizers)
		assert.Len(t, module.Authorizers, 1)

		auth := module.Authorizers[name]
		assert.NotNil(t, auth.Name)
		assert.Equal(t, name, *auth.Name)
		assert.NotNil(t, auth.AuthorizerType)
		assert.Equal(t, "JWT", *auth.AuthorizerType)
		assert.NotNil(t, auth.JWTConfiguration)
		assert.Equal(t, issuer, *auth.JWTConfiguration.Issuer)
		assert.Equal(t, audience, auth.JWTConfiguration.Audience)
	})

	t.Run("adds multiple JWT authorizers", func(t *testing.T) {
		module := NewModule("test_api")

		module.WithJWTAuthorizer("auth1", "issuer1", []string{"aud1"})
		module.WithJWTAuthorizer("auth2", "issuer2", []string{"aud2"})

		assert.Len(t, module.Authorizers, 2)
		assert.Contains(t, module.Authorizers, "auth1")
		assert.Contains(t, module.Authorizers, "auth2")
	})
}

func TestModule_WithLambdaAuthorizer(t *testing.T) {
	t.Run("adds Lambda authorizer", func(t *testing.T) {
		name := "custom-auth"
		lambdaURI := "arn:aws:lambda:us-east-1:123456789012:function:authorizer"
		identitySources := []string{"$request.header.Authorization"}

		module := NewModule("test_api")
		result := module.WithLambdaAuthorizer(name, lambdaURI, identitySources)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Authorizers)

		auth := module.Authorizers[name]
		assert.NotNil(t, auth.AuthorizerType)
		assert.Equal(t, "REQUEST", *auth.AuthorizerType)
		assert.NotNil(t, auth.AuthorizerURI)
		assert.Equal(t, lambdaURI, *auth.AuthorizerURI)
		assert.Equal(t, identitySources, auth.IdentitySources)
	})

	t.Run("supports multiple identity sources", func(t *testing.T) {
		module := NewModule("test_api")
		sources := []string{
			"$request.header.Authorization",
			"$context.identity.sourceIp",
		}
		module.WithLambdaAuthorizer("auth", "lambda-arn", sources)

		auth := module.Authorizers["auth"]
		assert.Len(t, auth.IdentitySources, 2)
	})
}

func TestModule_WithRoute(t *testing.T) {
	t.Run("adds a route", func(t *testing.T) {
		route := Route{
			RouteKey: "GET /users",
		}

		module := NewModule("test_api")
		result := module.WithRoute("get_users", route)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Routes)
		assert.Len(t, module.Routes, 1)
		assert.Equal(t, "GET /users", module.Routes["get_users"].RouteKey)
	})

	t.Run("adds multiple routes", func(t *testing.T) {
		module := NewModule("test_api")

		module.WithRoute("get", Route{RouteKey: "GET /users"})
		module.WithRoute("post", Route{RouteKey: "POST /users"})

		assert.Len(t, module.Routes, 2)
	})

	t.Run("supports ANY method", func(t *testing.T) {
		module := NewModule("test_api")
		module.WithRoute("catch_all", Route{RouteKey: "ANY /proxy"})

		assert.Equal(t, "ANY /proxy", module.Routes["catch_all"].RouteKey)
	})
}

func TestModule_WithIntegration(t *testing.T) {
	t.Run("adds Lambda integration", func(t *testing.T) {
		integration := Integration{
			IntegrationType: "AWS_PROXY",
			IntegrationURI:  ptr("arn:aws:lambda:us-east-1:123456789012:function:api"),
		}

		module := NewModule("test_api")
		result := module.WithIntegration("lambda", integration)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Integrations)
		assert.Len(t, module.Integrations, 1)
		assert.Equal(t, "AWS_PROXY", module.Integrations["lambda"].IntegrationType)
	})

	t.Run("adds HTTP integration", func(t *testing.T) {
		uri := "https://api.example.com"
		integration := Integration{
			IntegrationType: "HTTP_PROXY",
			IntegrationURI:  &uri,
		}

		module := NewModule("test_api")
		module.WithIntegration("http", integration)

		assert.Equal(t, "HTTP_PROXY", module.Integrations["http"].IntegrationType)
	})
}

func TestModule_WithTags(t *testing.T) {
	t.Run("adds tags to the API", func(t *testing.T) {
		tags := map[string]string{
			"Environment": "production",
			"Team":        "platform",
		}

		module := NewModule("test_api")
		result := module.WithTags(tags)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Tags)
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("merges tags when called multiple times", func(t *testing.T) {
		module := NewModule("test_api")

		module.WithTags(map[string]string{"Key1": "value1"})
		module.WithTags(map[string]string{"Key2": "value2"})

		assert.Equal(t, "value1", module.Tags["Key1"])
		assert.Equal(t, "value2", module.Tags["Key2"])
	})
}

func TestModule_LocalName(t *testing.T) {
	t.Run("returns name when set", func(t *testing.T) {
		name := "my_api"
		module := NewModule(name)

		assert.Equal(t, name, module.LocalName())
	})

	t.Run("returns default when name is nil", func(t *testing.T) {
		module := &Module{}

		assert.Equal(t, "api_gateway", module.LocalName())
	})
}

func TestModule_Configuration(t *testing.T) {
	t.Run("returns empty string and nil error as placeholder", func(t *testing.T) {
		module := NewModule("test_api")

		config, err := module.Configuration()

		require.NoError(t, err)
		assert.Empty(t, config)
	})
}

func TestModule_FluentAPI(t *testing.T) {
	t.Run("supports complete fluent configuration", func(t *testing.T) {
		module := NewModule("api").
			WithCORS([]string{"*"}, []string{"GET", "POST"}, []string{"Content-Type"}).
			WithDomainName("api.example.com", "cert-arn").
			WithJWTAuthorizer("auth", "issuer", []string{"aud"}).
			WithRoute("get", Route{RouteKey: "GET /users"}).
			WithTags(map[string]string{"Team": "platform"})

		assert.NotNil(t, module.Name)
		assert.NotNil(t, module.CORSConfiguration)
		assert.True(t, *module.CreateDomainName)
		assert.Len(t, module.Authorizers, 1)
		assert.Len(t, module.Routes, 1)
		assert.Equal(t, "platform", module.Tags["Team"])
	})
}

func TestCORSConfiguration(t *testing.T) {
	t.Run("creates complete CORS configuration", func(t *testing.T) {
		allowCreds := true
		maxAge := 3600
		cors := CORSConfiguration{
			AllowCredentials: &allowCreds,
			AllowOrigins:     []string{"https://example.com"},
			AllowMethods:     []string{"GET", "POST"},
			AllowHeaders:     []string{"Content-Type"},
			ExposeHeaders:    []string{"X-Custom-Header"},
			MaxAge:           &maxAge,
		}

		assert.True(t, *cors.AllowCredentials)
		assert.Equal(t, 3600, *cors.MaxAge)
	})
}

func TestRoute(t *testing.T) {
	t.Run("creates route with authorization", func(t *testing.T) {
		integrationKey := "lambda"
		authKey := "jwt-auth"
		authType := "JWT"
		opName := "GetUsers"

		route := Route{
			RouteKey:          "GET /users",
			IntegrationKey:    &integrationKey,
			AuthorizationKey:  &authKey,
			AuthorizationType: &authType,
			OperationName:     &opName,
		}

		assert.Equal(t, "GET /users", route.RouteKey)
		assert.Equal(t, "lambda", *route.IntegrationKey)
		assert.Equal(t, "JWT", *route.AuthorizationType)
	})
}

func TestIntegration(t *testing.T) {
	t.Run("creates Lambda integration with timeout", func(t *testing.T) {
		uri := "arn:aws:lambda:us-east-1:123456789012:function:api"
		timeout := 5000

		integration := Integration{
			IntegrationType:     "AWS_PROXY",
			IntegrationURI:      &uri,
			TimeoutMilliseconds: &timeout,
		}

		assert.Equal(t, "AWS_PROXY", integration.IntegrationType)
		assert.Equal(t, 5000, *integration.TimeoutMilliseconds)
	})
}

// Helper function to create pointer to string.
func ptr(s string) *string {
	return &s
}

// BenchmarkNewModule benchmarks module creation.
func BenchmarkNewModule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_api")
	}
}

// BenchmarkFluentAPI benchmarks fluent API calls.
func BenchmarkFluentAPI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_api").
			WithCORS([]string{"*"}, []string{"GET"}, []string{"Content-Type"}).
			WithRoute("test", Route{RouteKey: "GET /test"}).
			WithTags(map[string]string{"Environment": "production"})
	}
}
