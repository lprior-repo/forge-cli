// Package python provides E2E test helpers for Python Lambda function testing.

package python

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// createAWSSession creates an AWS session for testing.
func createAWSSession(t *testing.T, region string) *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		t.Fatalf("Failed to create AWS session: %v", err)
	}
	return sess
}

// getAWSRegion returns the AWS region from environment or default.
func getAWSRegion() string {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	return region
}

// getProjectRoot finds the project root directory.
func getProjectRoot(t *testing.T) string {
	// Start from current directory and walk up until we find go.mod
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (no go.mod found)")
		}
		dir = parent
	}
}
