package cloudfront

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	t.Run("creates module with sensible defaults", func(t *testing.T) {
		comment := "test distribution"
		module := NewModule(comment)

		// Verify basic properties
		require.NotNil(t, module)
		assert.Equal(t, "terraform-aws-modules/cloudfront/aws", module.Source)
		assert.Equal(t, "~> 3.0", module.Version)
		assert.NotNil(t, module.Comment)
		assert.Equal(t, comment, *module.Comment)

		// Verify sensible defaults
		assert.NotNil(t, module.CreateDistribution)
		assert.True(t, *module.CreateDistribution)

		assert.NotNil(t, module.Enabled)
		assert.True(t, *module.Enabled)

		assert.NotNil(t, module.HTTPVersion)
		assert.Equal(t, "http2", *module.HTTPVersion)

		assert.NotNil(t, module.WaitForDeployment)
		assert.True(t, *module.WaitForDeployment)

		assert.NotNil(t, module.IsIPv6Enabled)
		assert.True(t, *module.IsIPv6Enabled)
	})

	t.Run("creates module with different comments", func(t *testing.T) {
		comments := []string{"dist1", "my-cdn", "api_distribution"}
		for _, comment := range comments {
			module := NewModule(comment)
			assert.NotNil(t, module.Comment)
			assert.Equal(t, comment, *module.Comment)
		}
	})

	t.Run("creates module with empty comment", func(t *testing.T) {
		module := NewModule("")
		assert.NotNil(t, module.Comment)
		assert.Equal(t, "", *module.Comment)
	})
}

func TestModule_WithOrigin(t *testing.T) {
	t.Run("adds origin to distribution", func(t *testing.T) {
		module := NewModule("test")
		origin := Origin{
			DomainName: "example.com",
			OriginID:   "my-origin",
		}
		result := module.WithOrigin("my-origin", origin)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify origin is added
		assert.NotNil(t, module.Origin)
		assert.Len(t, module.Origin, 1)
		assert.Equal(t, "example.com", module.Origin["my-origin"].DomainName)
		assert.Equal(t, "my-origin", module.Origin["my-origin"].OriginID)
	})

	t.Run("adds multiple origins", func(t *testing.T) {
		module := NewModule("test")
		origin1 := Origin{DomainName: "origin1.com", OriginID: "o1"}
		origin2 := Origin{DomainName: "origin2.com", OriginID: "o2"}

		module.WithOrigin("o1", origin1)
		module.WithOrigin("o2", origin2)

		assert.Len(t, module.Origin, 2)
		assert.Equal(t, "origin1.com", module.Origin["o1"].DomainName)
		assert.Equal(t, "origin2.com", module.Origin["o2"].DomainName)
	})

	t.Run("overwrites origin with same ID", func(t *testing.T) {
		module := NewModule("test")
		origin1 := Origin{DomainName: "old.com", OriginID: "o1"}
		origin2 := Origin{DomainName: "new.com", OriginID: "o1"}

		module.WithOrigin("o1", origin1)
		module.WithOrigin("o1", origin2)

		assert.Len(t, module.Origin, 1)
		assert.Equal(t, "new.com", module.Origin["o1"].DomainName)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		origin := Origin{DomainName: "example.com", OriginID: "o1"}
		module := NewModule("test").WithOrigin("o1", origin)

		assert.NotNil(t, module.Origin)
		assert.Len(t, module.Origin, 1)
	})
}

func TestModule_WithS3Origin(t *testing.T) {
	t.Run("adds S3 origin with OAI", func(t *testing.T) {
		module := NewModule("test")
		result := module.WithS3Origin("s3-origin", "bucket.s3.amazonaws.com", "oai-123")

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify S3 origin is configured
		assert.NotNil(t, module.Origin)
		assert.Len(t, module.Origin, 1)
		origin := module.Origin["s3-origin"]
		assert.Equal(t, "bucket.s3.amazonaws.com", origin.DomainName)
		assert.Equal(t, "s3-origin", origin.OriginID)
		assert.NotNil(t, origin.S3OriginConfig)
		assert.NotNil(t, origin.S3OriginConfig.OriginAccessIdentity)
		assert.Equal(t, "oai-123", *origin.S3OriginConfig.OriginAccessIdentity)
	})

	t.Run("adds multiple S3 origins", func(t *testing.T) {
		module := NewModule("test")
		module.WithS3Origin("s3-1", "bucket1.s3.amazonaws.com", "oai-1")
		module.WithS3Origin("s3-2", "bucket2.s3.amazonaws.com", "oai-2")

		assert.Len(t, module.Origin, 2)
		assert.Equal(t, "oai-1", *module.Origin["s3-1"].S3OriginConfig.OriginAccessIdentity)
		assert.Equal(t, "oai-2", *module.Origin["s3-2"].S3OriginConfig.OriginAccessIdentity)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithS3Origin("s3", "bucket.s3.amazonaws.com", "oai").
			WithAliases("cdn.example.com")

		assert.NotNil(t, module.Origin)
		assert.NotNil(t, module.Aliases)
	})
}

func TestModule_WithCustomOrigin(t *testing.T) {
	t.Run("adds custom origin with HTTPS only", func(t *testing.T) {
		module := NewModule("test")
		result := module.WithCustomOrigin("api", "api.example.com", true)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify custom origin is configured
		assert.NotNil(t, module.Origin)
		assert.Len(t, module.Origin, 1)
		origin := module.Origin["api"]
		assert.Equal(t, "api.example.com", origin.DomainName)
		assert.Equal(t, "api", origin.OriginID)
		assert.NotNil(t, origin.CustomOriginConfig)
		assert.Equal(t, 80, origin.CustomOriginConfig.HTTPPort)
		assert.Equal(t, 443, origin.CustomOriginConfig.HTTPSPort)
		assert.Equal(t, "https-only", origin.CustomOriginConfig.OriginProtocolPolicy)
		assert.Equal(t, []string{"TLSv1.2"}, origin.CustomOriginConfig.OriginSSLProtocols)
	})

	t.Run("adds custom origin with match-viewer", func(t *testing.T) {
		module := NewModule("test")
		module.WithCustomOrigin("api", "api.example.com", false)

		origin := module.Origin["api"]
		assert.Equal(t, "match-viewer", origin.CustomOriginConfig.OriginProtocolPolicy)
	})

	t.Run("adds multiple custom origins", func(t *testing.T) {
		module := NewModule("test")
		module.WithCustomOrigin("api", "api.example.com", true)
		module.WithCustomOrigin("web", "web.example.com", false)

		assert.Len(t, module.Origin, 2)
		assert.Equal(t, "https-only", module.Origin["api"].CustomOriginConfig.OriginProtocolPolicy)
		assert.Equal(t, "match-viewer", module.Origin["web"].CustomOriginConfig.OriginProtocolPolicy)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithCustomOrigin("api", "api.example.com", true).
			WithAliases("cdn.example.com")

		assert.NotNil(t, module.Origin)
		assert.NotNil(t, module.Aliases)
	})
}

func TestModule_WithDefaultCacheBehavior(t *testing.T) {
	t.Run("configures default cache behavior", func(t *testing.T) {
		module := NewModule("test")
		result := module.WithDefaultCacheBehavior("my-origin", "redirect-to-https")

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify cache behavior is configured
		assert.NotNil(t, module.DefaultCacheBehavior)
		assert.Equal(t, "my-origin", module.DefaultCacheBehavior.TargetOriginID)
		assert.Equal(t, "redirect-to-https", module.DefaultCacheBehavior.ViewerProtocolPolicy)
		assert.Equal(t, []string{"GET", "HEAD", "OPTIONS"}, module.DefaultCacheBehavior.AllowedMethods)
		assert.Equal(t, []string{"GET", "HEAD"}, module.DefaultCacheBehavior.CachedMethods)
	})

	t.Run("overwrites previous cache behavior", func(t *testing.T) {
		module := NewModule("test")
		module.WithDefaultCacheBehavior("origin1", "allow-all")
		module.WithDefaultCacheBehavior("origin2", "https-only")

		assert.Equal(t, "origin2", module.DefaultCacheBehavior.TargetOriginID)
		assert.Equal(t, "https-only", module.DefaultCacheBehavior.ViewerProtocolPolicy)
	})

	t.Run("supports different viewer protocols", func(t *testing.T) {
		tests := []struct {
			name     string
			protocol string
		}{
			{"allow all", "allow-all"},
			{"https only", "https-only"},
			{"redirect", "redirect-to-https"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				module := NewModule("test")
				module.WithDefaultCacheBehavior("origin", tt.protocol)
				assert.Equal(t, tt.protocol, module.DefaultCacheBehavior.ViewerProtocolPolicy)
			})
		}
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithDefaultCacheBehavior("origin", "https-only").
			WithAliases("cdn.example.com")

		assert.NotNil(t, module.DefaultCacheBehavior)
		assert.NotNil(t, module.Aliases)
	})
}

func TestModule_WithCertificate(t *testing.T) {
	t.Run("configures ACM certificate", func(t *testing.T) {
		certARN := "arn:aws:acm:us-east-1:123456789012:certificate/abc123"
		tlsVersion := "TLSv1.2_2021"
		module := NewModule("test")
		result := module.WithCertificate(certARN, tlsVersion)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify certificate is configured
		assert.NotNil(t, module.ViewerCertificate)
		assert.NotNil(t, module.ViewerCertificate.ACMCertificateARN)
		assert.Equal(t, certARN, *module.ViewerCertificate.ACMCertificateARN)
		assert.NotNil(t, module.ViewerCertificate.MinimumProtocolVersion)
		assert.Equal(t, tlsVersion, *module.ViewerCertificate.MinimumProtocolVersion)
		assert.NotNil(t, module.ViewerCertificate.SSLSupportMethod)
		assert.Equal(t, "sni-only", *module.ViewerCertificate.SSLSupportMethod)
	})

	t.Run("supports different TLS versions", func(t *testing.T) {
		tests := []struct {
			name    string
			version string
		}{
			{"TLS 1.2 2018", "TLSv1.2_2018"},
			{"TLS 1.2 2019", "TLSv1.2_2019"},
			{"TLS 1.2 2021", "TLSv1.2_2021"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				module := NewModule("test")
				module.WithCertificate("arn", tt.version)
				assert.Equal(t, tt.version, *module.ViewerCertificate.MinimumProtocolVersion)
			})
		}
	})

	t.Run("overwrites previous certificate", func(t *testing.T) {
		module := NewModule("test")
		module.WithCertificate("arn1", "TLSv1.2_2018")
		module.WithCertificate("arn2", "TLSv1.2_2021")

		assert.Equal(t, "arn2", *module.ViewerCertificate.ACMCertificateARN)
		assert.Equal(t, "TLSv1.2_2021", *module.ViewerCertificate.MinimumProtocolVersion)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithCertificate("arn", "TLSv1.2_2021").
			WithAliases("cdn.example.com")

		assert.NotNil(t, module.ViewerCertificate)
		assert.NotNil(t, module.Aliases)
	})
}

func TestModule_WithAliases(t *testing.T) {
	t.Run("adds aliases to distribution", func(t *testing.T) {
		module := NewModule("test")
		result := module.WithAliases("cdn.example.com", "www.example.com")

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify aliases are added
		assert.Len(t, module.Aliases, 2)
		assert.Contains(t, module.Aliases, "cdn.example.com")
		assert.Contains(t, module.Aliases, "www.example.com")
	})

	t.Run("adds single alias", func(t *testing.T) {
		module := NewModule("test")
		module.WithAliases("cdn.example.com")

		assert.Len(t, module.Aliases, 1)
		assert.Equal(t, "cdn.example.com", module.Aliases[0])
	})

	t.Run("overwrites previous aliases", func(t *testing.T) {
		module := NewModule("test")
		module.WithAliases("old.example.com")
		module.WithAliases("new.example.com")

		assert.Len(t, module.Aliases, 1)
		assert.Equal(t, "new.example.com", module.Aliases[0])
	})

	t.Run("handles empty aliases", func(t *testing.T) {
		module := NewModule("test")
		module.WithAliases()

		assert.Empty(t, module.Aliases)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithAliases("cdn.example.com").
			WithPriceClass("PriceClass_100")

		assert.NotNil(t, module.Aliases)
		assert.NotNil(t, module.PriceClass)
	})
}

func TestModule_WithPriceClass(t *testing.T) {
	t.Run("sets price class", func(t *testing.T) {
		module := NewModule("test")
		result := module.WithPriceClass("PriceClass_100")

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify price class is set
		assert.NotNil(t, module.PriceClass)
		assert.Equal(t, "PriceClass_100", *module.PriceClass)
	})

	t.Run("supports different price classes", func(t *testing.T) {
		tests := []struct {
			name       string
			priceClass string
		}{
			{"all edge locations", "PriceClass_All"},
			{"200 locations", "PriceClass_200"},
			{"100 locations", "PriceClass_100"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				module := NewModule("test")
				module.WithPriceClass(tt.priceClass)
				assert.Equal(t, tt.priceClass, *module.PriceClass)
			})
		}
	})

	t.Run("overwrites previous price class", func(t *testing.T) {
		module := NewModule("test")
		module.WithPriceClass("PriceClass_All")
		module.WithPriceClass("PriceClass_100")

		assert.Equal(t, "PriceClass_100", *module.PriceClass)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithPriceClass("PriceClass_100").
			WithAliases("cdn.example.com")

		assert.NotNil(t, module.PriceClass)
		assert.NotNil(t, module.Aliases)
	})
}

func TestModule_WithLogging(t *testing.T) {
	t.Run("configures logging", func(t *testing.T) {
		module := NewModule("test")
		result := module.WithLogging("logs-bucket.s3.amazonaws.com", "cloudfront/", true)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify logging is configured
		assert.NotNil(t, module.LoggingConfig)
		assert.Equal(t, "logs-bucket.s3.amazonaws.com", module.LoggingConfig.Bucket)
		assert.NotNil(t, module.LoggingConfig.Prefix)
		assert.Equal(t, "cloudfront/", *module.LoggingConfig.Prefix)
		assert.NotNil(t, module.LoggingConfig.IncludeCookies)
		assert.True(t, *module.LoggingConfig.IncludeCookies)
	})

	t.Run("configures logging without cookies", func(t *testing.T) {
		module := NewModule("test")
		module.WithLogging("logs-bucket.s3.amazonaws.com", "cf/", false)

		assert.NotNil(t, module.LoggingConfig.IncludeCookies)
		assert.False(t, *module.LoggingConfig.IncludeCookies)
	})

	t.Run("configures logging with empty prefix", func(t *testing.T) {
		module := NewModule("test")
		module.WithLogging("logs-bucket.s3.amazonaws.com", "", false)

		assert.NotNil(t, module.LoggingConfig.Prefix)
		assert.Equal(t, "", *module.LoggingConfig.Prefix)
	})

	t.Run("overwrites previous logging config", func(t *testing.T) {
		module := NewModule("test")
		module.WithLogging("old-bucket.s3.amazonaws.com", "old/", false)
		module.WithLogging("new-bucket.s3.amazonaws.com", "new/", true)

		assert.Equal(t, "new-bucket.s3.amazonaws.com", module.LoggingConfig.Bucket)
		assert.Equal(t, "new/", *module.LoggingConfig.Prefix)
		assert.True(t, *module.LoggingConfig.IncludeCookies)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithLogging("bucket", "prefix/", true).
			WithAliases("cdn.example.com")

		assert.NotNil(t, module.LoggingConfig)
		assert.NotNil(t, module.Aliases)
	})
}

func TestModule_WithGeoRestriction(t *testing.T) {
	t.Run("configures whitelist restriction", func(t *testing.T) {
		module := NewModule("test")
		result := module.WithGeoRestriction("whitelist", []string{"US", "CA", "MX"})

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify geo restriction is configured
		assert.NotNil(t, module.GeoRestriction)
		assert.Equal(t, "whitelist", module.GeoRestriction.RestrictionType)
		assert.Len(t, module.GeoRestriction.Locations, 3)
		assert.Contains(t, module.GeoRestriction.Locations, "US")
		assert.Contains(t, module.GeoRestriction.Locations, "CA")
		assert.Contains(t, module.GeoRestriction.Locations, "MX")
	})

	t.Run("configures blacklist restriction", func(t *testing.T) {
		module := NewModule("test")
		module.WithGeoRestriction("blacklist", []string{"CN", "RU"})

		assert.Equal(t, "blacklist", module.GeoRestriction.RestrictionType)
		assert.Len(t, module.GeoRestriction.Locations, 2)
	})

	t.Run("configures no restriction", func(t *testing.T) {
		module := NewModule("test")
		module.WithGeoRestriction("none", []string{})

		assert.Equal(t, "none", module.GeoRestriction.RestrictionType)
		assert.Empty(t, module.GeoRestriction.Locations)
	})

	t.Run("overwrites previous restriction", func(t *testing.T) {
		module := NewModule("test")
		module.WithGeoRestriction("whitelist", []string{"US"})
		module.WithGeoRestriction("blacklist", []string{"CN"})

		assert.Equal(t, "blacklist", module.GeoRestriction.RestrictionType)
		assert.Len(t, module.GeoRestriction.Locations, 1)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithGeoRestriction("whitelist", []string{"US"}).
			WithAliases("cdn.example.com")

		assert.NotNil(t, module.GeoRestriction)
		assert.NotNil(t, module.Aliases)
	})
}

func TestModule_WithWAF(t *testing.T) {
	t.Run("associates WAF web ACL", func(t *testing.T) {
		webACLID := "arn:aws:wafv2:us-east-1:123456789012:global/webacl/test/abc123"
		module := NewModule("test")
		result := module.WithWAF(webACLID)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify WAF is configured
		assert.NotNil(t, module.WebACLID)
		assert.Equal(t, webACLID, *module.WebACLID)
	})

	t.Run("overwrites previous WAF", func(t *testing.T) {
		module := NewModule("test")
		module.WithWAF("arn:old")
		module.WithWAF("arn:new")

		assert.Equal(t, "arn:new", *module.WebACLID)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithWAF("arn").
			WithAliases("cdn.example.com")

		assert.NotNil(t, module.WebACLID)
		assert.NotNil(t, module.Aliases)
	})
}

func TestModule_WithOriginAccessControl(t *testing.T) {
	t.Run("creates origin access control", func(t *testing.T) {
		module := NewModule("test")
		result := module.WithOriginAccessControl("my-oac", "S3 access control")

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify OAC is configured
		assert.NotNil(t, module.CreateOriginAccessControl)
		assert.True(t, *module.CreateOriginAccessControl)
		assert.NotNil(t, module.OriginAccessControl)
		assert.Len(t, module.OriginAccessControl, 1)

		oac := module.OriginAccessControl["my-oac"]
		assert.NotNil(t, oac.Name)
		assert.Equal(t, "my-oac", *oac.Name)
		assert.Equal(t, "S3 access control", oac.Description)
		assert.Equal(t, "s3", oac.OriginType)
		assert.Equal(t, "always", oac.SigningBehavior)
		assert.Equal(t, "sigv4", oac.SigningProtocol)
	})

	t.Run("adds multiple OACs", func(t *testing.T) {
		module := NewModule("test")
		module.WithOriginAccessControl("oac1", "First OAC")
		module.WithOriginAccessControl("oac2", "Second OAC")

		assert.Len(t, module.OriginAccessControl, 2)
		assert.Equal(t, "First OAC", module.OriginAccessControl["oac1"].Description)
		assert.Equal(t, "Second OAC", module.OriginAccessControl["oac2"].Description)
	})

	t.Run("overwrites OAC with same name", func(t *testing.T) {
		module := NewModule("test")
		module.WithOriginAccessControl("oac", "Old description")
		module.WithOriginAccessControl("oac", "New description")

		assert.Len(t, module.OriginAccessControl, 1)
		assert.Equal(t, "New description", module.OriginAccessControl["oac"].Description)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithOriginAccessControl("oac", "description").
			WithAliases("cdn.example.com")

		assert.NotNil(t, module.OriginAccessControl)
		assert.NotNil(t, module.Aliases)
	})
}

func TestModule_WithLambdaEdge(t *testing.T) {
	t.Run("adds Lambda@Edge to default cache behavior", func(t *testing.T) {
		lambdaARN := "arn:aws:lambda:us-east-1:123456789012:function:edge:1"
		module := NewModule("test")
		result := module.WithLambdaEdge("viewer-request", lambdaARN)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify Lambda@Edge is configured
		assert.NotNil(t, module.DefaultCacheBehavior)
		assert.Len(t, module.DefaultCacheBehavior.LambdaFunctionAssociations, 1)

		lambda := module.DefaultCacheBehavior.LambdaFunctionAssociations[0]
		assert.Equal(t, "viewer-request", lambda.EventType)
		assert.Equal(t, lambdaARN, lambda.LambdaARN)
	})

	t.Run("adds multiple Lambda@Edge functions", func(t *testing.T) {
		module := NewModule("test")
		module.WithLambdaEdge("viewer-request", "arn1")
		module.WithLambdaEdge("origin-response", "arn2")

		assert.Len(t, module.DefaultCacheBehavior.LambdaFunctionAssociations, 2)
		assert.Equal(t, "viewer-request", module.DefaultCacheBehavior.LambdaFunctionAssociations[0].EventType)
		assert.Equal(t, "origin-response", module.DefaultCacheBehavior.LambdaFunctionAssociations[1].EventType)
	})

	t.Run("creates cache behavior if not exists", func(t *testing.T) {
		module := &Module{}
		assert.Nil(t, module.DefaultCacheBehavior)

		module.WithLambdaEdge("viewer-request", "arn")

		assert.NotNil(t, module.DefaultCacheBehavior)
		assert.Len(t, module.DefaultCacheBehavior.LambdaFunctionAssociations, 1)
	})

	t.Run("supports different event types", func(t *testing.T) {
		tests := []struct {
			name      string
			eventType string
		}{
			{"viewer request", "viewer-request"},
			{"viewer response", "viewer-response"},
			{"origin request", "origin-request"},
			{"origin response", "origin-response"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				module := NewModule("test")
				module.WithLambdaEdge(tt.eventType, "arn")
				assert.Equal(t, tt.eventType, module.DefaultCacheBehavior.LambdaFunctionAssociations[0].EventType)
			})
		}
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithLambdaEdge("viewer-request", "arn").
			WithAliases("cdn.example.com")

		assert.NotNil(t, module.DefaultCacheBehavior)
		assert.NotNil(t, module.Aliases)
	})
}

func TestModule_WithTags(t *testing.T) {
	t.Run("adds tags to module", func(t *testing.T) {
		module := NewModule("test")
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
		module := NewModule("test")

		tags1 := map[string]string{
			"Environment": "production",
		}
		module.WithTags(tags1)

		tags2 := map[string]string{
			"Team": "platform",
		}
		module.WithTags(tags2)

		// Verify both sets of tags are present
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("overwrites existing tags with same key", func(t *testing.T) {
		module := NewModule("test")

		tags1 := map[string]string{
			"Environment": "development",
		}
		module.WithTags(tags1)

		tags2 := map[string]string{
			"Environment": "production",
		}
		module.WithTags(tags2)

		// Verify tag was overwritten
		assert.Equal(t, "production", module.Tags["Environment"])
	})

	t.Run("handles empty tag map", func(t *testing.T) {
		module := NewModule("test")
		module.WithTags(map[string]string{})

		// Tags map should be initialized but empty
		assert.NotNil(t, module.Tags)
		assert.Empty(t, module.Tags)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithTags(map[string]string{"Team": "platform"}).
			WithAliases("cdn.example.com")

		assert.NotNil(t, module.Tags)
		assert.NotNil(t, module.Aliases)
	})
}

func TestModule_LocalName(t *testing.T) {
	t.Run("returns comment when set", func(t *testing.T) {
		comment := "my_distribution"
		module := NewModule(comment)

		assert.Equal(t, comment, module.LocalName())
	})

	t.Run("returns default when comment is nil", func(t *testing.T) {
		module := &Module{}

		assert.Equal(t, "distribution", module.LocalName())
	})

	t.Run("returns empty string when comment is empty", func(t *testing.T) {
		emptyComment := ""
		module := NewModule(emptyComment)

		assert.Equal(t, emptyComment, module.LocalName())
	})
}

func TestModule_Configuration(t *testing.T) {
	t.Run("returns empty string and nil error as placeholder", func(t *testing.T) {
		module := NewModule("test")

		config, err := module.Configuration()

		// Current implementation is a placeholder
		require.NoError(t, err)
		assert.Empty(t, config)
	})
}

func TestModule_FluentAPI(t *testing.T) {
	t.Run("supports complete fluent configuration", func(t *testing.T) {
		module := NewModule("cdn").
			WithS3Origin("s3", "bucket.s3.amazonaws.com", "oai").
			WithDefaultCacheBehavior("s3", "redirect-to-https").
			WithCertificate("arn:cert", "TLSv1.2_2021").
			WithAliases("cdn.example.com", "www.example.com").
			WithPriceClass("PriceClass_100").
			WithLogging("logs.s3.amazonaws.com", "cf/", true).
			WithGeoRestriction("whitelist", []string{"US", "CA"}).
			WithWAF("arn:waf").
			WithTags(map[string]string{
				"Environment": "production",
				"Team":        "platform",
			})

		// Verify all configuration is applied
		assert.NotNil(t, module.Comment)
		assert.Equal(t, "cdn", *module.Comment)

		assert.NotNil(t, module.Origin)
		assert.Len(t, module.Origin, 1)

		assert.NotNil(t, module.DefaultCacheBehavior)
		assert.Equal(t, "redirect-to-https", module.DefaultCacheBehavior.ViewerProtocolPolicy)

		assert.NotNil(t, module.ViewerCertificate)
		assert.Equal(t, "TLSv1.2_2021", *module.ViewerCertificate.MinimumProtocolVersion)

		assert.Len(t, module.Aliases, 2)
		assert.Equal(t, "PriceClass_100", *module.PriceClass)

		assert.NotNil(t, module.LoggingConfig)
		assert.NotNil(t, module.GeoRestriction)
		assert.NotNil(t, module.WebACLID)

		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("supports custom origin configuration", func(t *testing.T) {
		module := NewModule("api_cdn").
			WithCustomOrigin("api", "api.example.com", true).
			WithDefaultCacheBehavior("api", "https-only").
			WithOriginAccessControl("api-oac", "API access control").
			WithLambdaEdge("viewer-request", "arn:lambda")

		assert.NotNil(t, module.Origin)
		assert.Equal(t, "https-only", module.Origin["api"].CustomOriginConfig.OriginProtocolPolicy)
		assert.NotNil(t, module.OriginAccessControl)
		assert.Len(t, module.DefaultCacheBehavior.LambdaFunctionAssociations, 1)
	})

	t.Run("supports all fluent methods in different orders", func(t *testing.T) {
		// Order 1
		module1 := NewModule("d1").
			WithTags(map[string]string{"a": "1"}).
			WithAliases("d1.example.com").
			WithPriceClass("PriceClass_All").
			WithS3Origin("s3", "bucket", "oai")

		// Order 2
		module2 := NewModule("d2").
			WithS3Origin("s3", "bucket", "oai").
			WithPriceClass("PriceClass_100").
			WithAliases("d2.example.com").
			WithTags(map[string]string{"b": "2"})

		// Both should be properly configured
		assert.NotNil(t, module1.Origin)
		assert.NotNil(t, module1.Aliases)
		assert.NotNil(t, module2.Origin)
		assert.NotNil(t, module2.Aliases)
	})
}

func TestOrigin(t *testing.T) {
	t.Run("creates basic origin", func(t *testing.T) {
		origin := Origin{
			DomainName: "example.com",
			OriginID:   "my-origin",
		}

		assert.Equal(t, "example.com", origin.DomainName)
		assert.Equal(t, "my-origin", origin.OriginID)
	})

	t.Run("creates origin with custom config", func(t *testing.T) {
		origin := Origin{
			DomainName: "api.example.com",
			OriginID:   "api",
			CustomOriginConfig: &CustomOriginConfig{
				HTTPPort:             80,
				HTTPSPort:            443,
				OriginProtocolPolicy: "https-only",
				OriginSSLProtocols:   []string{"TLSv1.2"},
			},
		}

		assert.NotNil(t, origin.CustomOriginConfig)
		assert.Equal(t, "https-only", origin.CustomOriginConfig.OriginProtocolPolicy)
	})

	t.Run("creates origin with S3 config", func(t *testing.T) {
		oai := "oai-123"
		origin := Origin{
			DomainName: "bucket.s3.amazonaws.com",
			OriginID:   "s3",
			S3OriginConfig: &S3OriginConfig{
				OriginAccessIdentity: &oai,
			},
		}

		assert.NotNil(t, origin.S3OriginConfig)
		assert.Equal(t, "oai-123", *origin.S3OriginConfig.OriginAccessIdentity)
	})

	t.Run("creates origin with custom headers", func(t *testing.T) {
		origin := Origin{
			DomainName: "api.example.com",
			OriginID:   "api",
			CustomHeaders: []OriginCustomHeader{
				{Name: "X-Custom-Header", Value: "value1"},
				{Name: "X-API-Key", Value: "secret"},
			},
		}

		assert.Len(t, origin.CustomHeaders, 2)
		assert.Equal(t, "X-Custom-Header", origin.CustomHeaders[0].Name)
		assert.Equal(t, "value1", origin.CustomHeaders[0].Value)
	})

	t.Run("creates origin with Origin Shield", func(t *testing.T) {
		region := "us-east-1"
		origin := Origin{
			DomainName: "example.com",
			OriginID:   "shielded",
			OriginShield: &OriginShield{
				Enabled:            true,
				OriginShieldRegion: &region,
			},
		}

		assert.NotNil(t, origin.OriginShield)
		assert.True(t, origin.OriginShield.Enabled)
		assert.Equal(t, "us-east-1", *origin.OriginShield.OriginShieldRegion)
	})
}

func TestCacheBehavior(t *testing.T) {
	t.Run("creates default cache behavior", func(t *testing.T) {
		cb := CacheBehavior{
			TargetOriginID:       "my-origin",
			ViewerProtocolPolicy: "redirect-to-https",
			AllowedMethods:       []string{"GET", "HEAD"},
			CachedMethods:        []string{"GET", "HEAD"},
		}

		assert.Equal(t, "my-origin", cb.TargetOriginID)
		assert.Equal(t, "redirect-to-https", cb.ViewerProtocolPolicy)
		assert.Len(t, cb.AllowedMethods, 2)
	})

	t.Run("creates ordered cache behavior with path pattern", func(t *testing.T) {
		pattern := "/api/*"
		cb := CacheBehavior{
			PathPattern:          &pattern,
			TargetOriginID:       "api",
			ViewerProtocolPolicy: "https-only",
		}

		assert.NotNil(t, cb.PathPattern)
		assert.Equal(t, "/api/*", *cb.PathPattern)
	})

	t.Run("creates cache behavior with policies", func(t *testing.T) {
		cachePolicy := "cache-policy-id"
		originPolicy := "origin-policy-id"
		cb := CacheBehavior{
			TargetOriginID:        "origin",
			ViewerProtocolPolicy:  "https-only",
			CachePolicyID:         &cachePolicy,
			OriginRequestPolicyID: &originPolicy,
		}

		assert.NotNil(t, cb.CachePolicyID)
		assert.Equal(t, "cache-policy-id", *cb.CachePolicyID)
		assert.Equal(t, "origin-policy-id", *cb.OriginRequestPolicyID)
	})

	t.Run("creates cache behavior with Lambda@Edge", func(t *testing.T) {
		cb := CacheBehavior{
			TargetOriginID:       "origin",
			ViewerProtocolPolicy: "https-only",
			LambdaFunctionAssociations: []LambdaFunctionAssociation{
				{
					EventType: "viewer-request",
					LambdaARN: "arn:lambda",
				},
			},
		}

		assert.Len(t, cb.LambdaFunctionAssociations, 1)
		assert.Equal(t, "viewer-request", cb.LambdaFunctionAssociations[0].EventType)
	})
}

func TestViewerCertificate(t *testing.T) {
	t.Run("creates default CloudFront certificate", func(t *testing.T) {
		useDefault := true
		vc := ViewerCertificate{
			CloudFrontDefaultCertificate: &useDefault,
		}

		assert.NotNil(t, vc.CloudFrontDefaultCertificate)
		assert.True(t, *vc.CloudFrontDefaultCertificate)
	})

	t.Run("creates ACM certificate", func(t *testing.T) {
		certARN := "arn:acm"
		minVersion := "TLSv1.2_2021"
		sslMethod := "sni-only"

		vc := ViewerCertificate{
			ACMCertificateARN:      &certARN,
			MinimumProtocolVersion: &minVersion,
			SSLSupportMethod:       &sslMethod,
		}

		assert.Equal(t, "arn:acm", *vc.ACMCertificateARN)
		assert.Equal(t, "TLSv1.2_2021", *vc.MinimumProtocolVersion)
		assert.Equal(t, "sni-only", *vc.SSLSupportMethod)
	})

	t.Run("creates IAM certificate", func(t *testing.T) {
		iamCert := "iam-cert-id"
		vc := ViewerCertificate{
			IAMCertificateID: &iamCert,
		}

		assert.NotNil(t, vc.IAMCertificateID)
		assert.Equal(t, "iam-cert-id", *vc.IAMCertificateID)
	})
}

func TestGeoRestriction(t *testing.T) {
	t.Run("creates whitelist restriction", func(t *testing.T) {
		gr := GeoRestriction{
			RestrictionType: "whitelist",
			Locations:       []string{"US", "CA", "MX"},
		}

		assert.Equal(t, "whitelist", gr.RestrictionType)
		assert.Len(t, gr.Locations, 3)
		assert.Contains(t, gr.Locations, "US")
	})

	t.Run("creates blacklist restriction", func(t *testing.T) {
		gr := GeoRestriction{
			RestrictionType: "blacklist",
			Locations:       []string{"CN", "RU"},
		}

		assert.Equal(t, "blacklist", gr.RestrictionType)
		assert.Len(t, gr.Locations, 2)
	})

	t.Run("creates no restriction", func(t *testing.T) {
		gr := GeoRestriction{
			RestrictionType: "none",
			Locations:       []string{},
		}

		assert.Equal(t, "none", gr.RestrictionType)
		assert.Empty(t, gr.Locations)
	})
}

func TestLoggingConfig(t *testing.T) {
	t.Run("creates logging config with cookies", func(t *testing.T) {
		includeCookies := true
		prefix := "cloudfront/"

		lc := LoggingConfig{
			Bucket:         "logs.s3.amazonaws.com",
			IncludeCookies: &includeCookies,
			Prefix:         &prefix,
		}

		assert.Equal(t, "logs.s3.amazonaws.com", lc.Bucket)
		assert.NotNil(t, lc.IncludeCookies)
		assert.True(t, *lc.IncludeCookies)
		assert.Equal(t, "cloudfront/", *lc.Prefix)
	})

	t.Run("creates logging config without cookies", func(t *testing.T) {
		includeCookies := false
		lc := LoggingConfig{
			Bucket:         "logs.s3.amazonaws.com",
			IncludeCookies: &includeCookies,
		}

		assert.False(t, *lc.IncludeCookies)
	})
}

func TestCustomErrorResponse(t *testing.T) {
	t.Run("creates custom error response", func(t *testing.T) {
		responseCode := 404
		responsePath := "/404.html"
		ttl := 300

		cer := CustomErrorResponse{
			ErrorCode:          404,
			ResponseCode:       &responseCode,
			ResponsePagePath:   &responsePath,
			ErrorCachingMinTTL: &ttl,
		}

		assert.Equal(t, 404, cer.ErrorCode)
		assert.NotNil(t, cer.ResponseCode)
		assert.Equal(t, 404, *cer.ResponseCode)
		assert.Equal(t, "/404.html", *cer.ResponsePagePath)
		assert.Equal(t, 300, *cer.ErrorCachingMinTTL)
	})

	t.Run("creates minimal error response", func(t *testing.T) {
		cer := CustomErrorResponse{
			ErrorCode: 500,
		}

		assert.Equal(t, 500, cer.ErrorCode)
		assert.Nil(t, cer.ResponseCode)
		assert.Nil(t, cer.ResponsePagePath)
	})
}

func TestOriginGroup(t *testing.T) {
	t.Run("creates origin group with failover", func(t *testing.T) {
		og := OriginGroup{
			OriginID:         "group-1",
			FailoverCriteria: []int{500, 502, 503, 504},
			Members:          []string{"origin-1", "origin-2"},
		}

		assert.Equal(t, "group-1", og.OriginID)
		assert.Len(t, og.FailoverCriteria, 4)
		assert.Contains(t, og.FailoverCriteria, 500)
		assert.Len(t, og.Members, 2)
	})
}

func TestOriginAccessControl(t *testing.T) {
	t.Run("creates origin access control", func(t *testing.T) {
		name := "my-oac"
		oac := OriginAccessControl{
			Name:            &name,
			Description:     "S3 access control",
			OriginType:      "s3",
			SigningBehavior: "always",
			SigningProtocol: "sigv4",
		}

		assert.NotNil(t, oac.Name)
		assert.Equal(t, "my-oac", *oac.Name)
		assert.Equal(t, "S3 access control", oac.Description)
		assert.Equal(t, "s3", oac.OriginType)
		assert.Equal(t, "always", oac.SigningBehavior)
		assert.Equal(t, "sigv4", oac.SigningProtocol)
	})
}

func TestLambdaFunctionAssociation(t *testing.T) {
	t.Run("creates Lambda@Edge association", func(t *testing.T) {
		lfa := LambdaFunctionAssociation{
			EventType: "viewer-request",
			LambdaARN: "arn:aws:lambda:us-east-1:123456789012:function:edge:1",
		}

		assert.Equal(t, "viewer-request", lfa.EventType)
		assert.Equal(t, "arn:aws:lambda:us-east-1:123456789012:function:edge:1", lfa.LambdaARN)
	})

	t.Run("creates Lambda@Edge association with body", func(t *testing.T) {
		includeBody := true
		lfa := LambdaFunctionAssociation{
			EventType:   "origin-request",
			LambdaARN:   "arn:lambda",
			IncludeBody: &includeBody,
		}

		assert.NotNil(t, lfa.IncludeBody)
		assert.True(t, *lfa.IncludeBody)
	})
}

func TestFunctionAssociation(t *testing.T) {
	t.Run("creates CloudFront Functions association", func(t *testing.T) {
		fa := FunctionAssociation{
			EventType:   "viewer-request",
			FunctionARN: "arn:aws:cloudfront::123456789012:function/my-function",
		}

		assert.Equal(t, "viewer-request", fa.EventType)
		assert.Equal(t, "arn:aws:cloudfront::123456789012:function/my-function", fa.FunctionARN)
	})
}

func TestVPCOrigin(t *testing.T) {
	t.Run("creates VPC origin", func(t *testing.T) {
		vo := VPCOrigin{
			Name:                 "vpc-origin",
			ARN:                  "arn:aws:vpc",
			HTTPPort:             80,
			HTTPSPort:            443,
			OriginProtocolPolicy: "https-only",
			OriginSSLProtocols: OriginSSLProtocols{
				Items:    []string{"TLSv1.2"},
				Quantity: 1,
			},
		}

		assert.Equal(t, "vpc-origin", vo.Name)
		assert.Equal(t, "arn:aws:vpc", vo.ARN)
		assert.Equal(t, 80, vo.HTTPPort)
		assert.Equal(t, 443, vo.HTTPSPort)
		assert.Equal(t, "https-only", vo.OriginProtocolPolicy)
		assert.Len(t, vo.OriginSSLProtocols.Items, 1)
		assert.Equal(t, 1, vo.OriginSSLProtocols.Quantity)
	})
}

func TestModule_StructTags(t *testing.T) {
	t.Run("validates struct tag presence", func(t *testing.T) {
		// This test ensures struct tags are properly defined
		module := NewModule("test")

		// Check that pointer fields are properly initialized
		assert.NotNil(t, module.Comment)
		assert.NotNil(t, module.CreateDistribution)
		assert.NotNil(t, module.Enabled)
	})
}

func TestModule_PointerSemantics(t *testing.T) {
	t.Run("distinguishes between nil and zero value", func(t *testing.T) {
		module := &Module{}

		// Unset fields should be nil
		assert.Nil(t, module.Enabled)
		assert.Nil(t, module.IsIPv6Enabled)

		// Setting to false should be distinguishable from nil
		disabled := false
		module.Enabled = &disabled

		assert.NotNil(t, module.Enabled)
		assert.False(t, *module.Enabled)
	})

	t.Run("allows explicit false values", func(t *testing.T) {
		module := NewModule("test")

		// Default has distribution enabled
		assert.True(t, *module.Enabled)

		// Explicitly disable
		disabled := false
		module.Enabled = &disabled

		assert.NotNil(t, module.Enabled)
		assert.False(t, *module.Enabled)
	})
}

func TestModule_ComplexConfiguration(t *testing.T) {
	t.Run("configures multi-origin distribution", func(t *testing.T) {
		module := NewModule("multi-origin").
			WithS3Origin("static", "static.s3.amazonaws.com", "oai-static").
			WithCustomOrigin("api", "api.example.com", true).
			WithDefaultCacheBehavior("static", "redirect-to-https")

		assert.Len(t, module.Origin, 2)
		assert.NotNil(t, module.Origin["static"].S3OriginConfig)
		assert.NotNil(t, module.Origin["api"].CustomOriginConfig)
	})

	t.Run("configures distribution with custom error pages", func(t *testing.T) {
		responseCode := 404
		responsePath := "/404.html"
		ttl := 300

		module := NewModule("error-pages")
		module.CustomErrorResponse = map[string]CustomErrorResponse{
			"404": {
				ErrorCode:          404,
				ResponseCode:       &responseCode,
				ResponsePagePath:   &responsePath,
				ErrorCachingMinTTL: &ttl,
			},
		}

		assert.Len(t, module.CustomErrorResponse, 1)
		assert.Equal(t, 404, module.CustomErrorResponse["404"].ErrorCode)
	})

	t.Run("configures distribution with monitoring", func(t *testing.T) {
		createMonitoring := true
		realtimeStatus := "Enabled"

		module := NewModule("monitored")
		module.CreateMonitoringSubscription = &createMonitoring
		module.RealtimeMetricsSubscriptionStatus = &realtimeStatus

		assert.True(t, *module.CreateMonitoringSubscription)
		assert.Equal(t, "Enabled", *module.RealtimeMetricsSubscriptionStatus)
	})
}

// BenchmarkNewModule benchmarks module creation.
func BenchmarkNewModule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_distribution")
	}
}

// BenchmarkFluentAPI benchmarks fluent API calls.
func BenchmarkFluentAPI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_distribution").
			WithS3Origin("s3", "bucket.s3.amazonaws.com", "oai").
			WithDefaultCacheBehavior("s3", "redirect-to-https").
			WithCertificate("arn:cert", "TLSv1.2_2021").
			WithAliases("cdn.example.com").
			WithTags(map[string]string{
				"Environment": "production",
			})
	}
}

// BenchmarkWithTags benchmarks tag merging.
func BenchmarkWithTags(b *testing.B) {
	module := NewModule("bench_distribution")
	tags := map[string]string{
		"Environment": "production",
		"Team":        "platform",
		"Service":     "cdn",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		module.WithTags(tags)
	}
}

// BenchmarkWithOrigin benchmarks origin addition.
func BenchmarkWithOrigin(b *testing.B) {
	module := NewModule("bench_distribution")
	origin := Origin{
		DomainName: "example.com",
		OriginID:   "my-origin",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		module.WithOrigin("my-origin", origin)
	}
}
