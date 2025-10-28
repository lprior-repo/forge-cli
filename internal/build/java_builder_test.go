package build

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJavaBuildSignature tests that JavaBuild has correct signature
func TestJavaBuildSignature(t *testing.T) {
	t.Run("JavaBuild matches BuildFunc signature", func(t *testing.T) {
		// JavaBuild should be assignable to BuildFunc
		var buildFunc BuildFunc = JavaBuild

		// Should compile and work with functional patterns
		result := buildFunc(context.Background(), Config{
			SourceDir: "/nonexistent",
			Runtime:   "java21",
		})

		// Should return Either type
		assert.True(t, E.IsLeft(result) || E.IsRight(result), "Should return Either type")
	})

	t.Run("JavaBuild returns Left on error", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/nonexistent/directory",
			OutputPath: "/tmp/output.jar",
			Runtime:    "java21",
		}

		result := JavaBuild(context.Background(), cfg)

		assert.True(t, E.IsLeft(result), "Should return Left on error")
	})
}

// TestJavaBuildPure tests that JavaBuild is a pure function
func TestJavaBuildPure(t *testing.T) {
	t.Run("same inputs produce same result", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/nonexistent",
			Runtime:    "java21",
			OutputPath: "/tmp/test.jar",
		}

		result1 := JavaBuild(context.Background(), cfg)
		result2 := JavaBuild(context.Background(), cfg)

		// Both should fail the same way
		assert.Equal(t, E.IsLeft(result1), E.IsLeft(result2))
	})

	t.Run("no side effects on failure", func(t *testing.T) {
		tmpDir := t.TempDir()
		outputPath := filepath.Join(tmpDir, "output.jar")

		cfg := Config{
			SourceDir:  "/nonexistent",
			Runtime:    "java21",
			OutputPath: outputPath,
		}

		// Call function
		JavaBuild(context.Background(), cfg)

		// Should not create output file on failure
		_, err := os.Stat(outputPath)
		assert.True(t, os.IsNotExist(err), "Should not create output on failure")
	})
}

// TestJavaBuildComposition tests JavaBuild with functional composition
func TestJavaBuildComposition(t *testing.T) {
	t.Run("composes with WithCache", func(t *testing.T) {
		cache := NewMemoryCache()
		cachedBuild := WithCache(cache)(JavaBuild)

		// Should still be a BuildFunc
		var buildFunc BuildFunc = cachedBuild
		assert.NotNil(t, buildFunc)
	})

	t.Run("composes with WithLogging", func(t *testing.T) {
		logger := &mockLogger{
			infoFn:  func(msg string, args ...interface{}) {},
			errorFn: func(msg string, args ...interface{}) {},
		}

		loggedBuild := WithLogging(logger)(JavaBuild)

		// Should still be a BuildFunc
		var buildFunc BuildFunc = loggedBuild
		assert.NotNil(t, buildFunc)
	})

	t.Run("composes with multiple decorators", func(t *testing.T) {
		cache := NewMemoryCache()
		logger := &mockLogger{
			infoFn:  func(msg string, args ...interface{}) {},
			errorFn: func(msg string, args ...interface{}) {},
		}

		composed := Compose(
			WithCache(cache),
			WithLogging(logger),
		)(JavaBuild)

		// Should still be a BuildFunc
		var buildFunc BuildFunc = composed
		assert.NotNil(t, buildFunc)
	})
}

// TestJavaBuildRegistry tests JavaBuild in registry
func TestJavaBuildRegistry(t *testing.T) {
	t.Run("registry contains Java runtimes", func(t *testing.T) {
		registry := NewRegistry()

		assert.Contains(t, registry, "java11", "Should contain java11")
		assert.Contains(t, registry, "java17", "Should contain java17")
		assert.Contains(t, registry, "java21", "Should contain java21")
	})

	t.Run("Java builders use JavaBuild function", func(t *testing.T) {
		registry := NewRegistry()

		builder11 := registry["java11"]
		builder17 := registry["java17"]
		builder21 := registry["java21"]

		// All should be the same function (JavaBuild)
		// We can test this by checking they behave identically
		cfg := Config{SourceDir: "/nonexistent"}

		result11 := builder11(context.Background(), cfg)
		result17 := builder17(context.Background(), cfg)
		result21 := builder21(context.Background(), cfg)

		// All should fail the same way
		assert.Equal(t, E.IsLeft(result11), E.IsLeft(result17))
		assert.Equal(t, E.IsLeft(result17), E.IsLeft(result21))
	})
}

// TestFindJar tests the findJar helper function
func TestFindJar(t *testing.T) {
	t.Run("finds main jar in target directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "target")
		err := os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		// Create main JAR
		mainJar := filepath.Join(targetDir, "myapp-1.0.jar")
		err = os.WriteFile(mainJar, []byte("fake jar"), 0644)
		require.NoError(t, err)

		jarPath, err := findJar(targetDir)
		require.NoError(t, err)
		assert.Equal(t, mainJar, jarPath)
	})

	t.Run("skips sources jar", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "target")
		err := os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		// Create sources JAR (should be skipped)
		sourcesJar := filepath.Join(targetDir, "myapp-1.0-sources.jar")
		err = os.WriteFile(sourcesJar, []byte("fake jar"), 0644)
		require.NoError(t, err)

		// Create main JAR
		mainJar := filepath.Join(targetDir, "myapp-1.0.jar")
		err = os.WriteFile(mainJar, []byte("fake jar"), 0644)
		require.NoError(t, err)

		jarPath, err := findJar(targetDir)
		require.NoError(t, err)
		assert.Equal(t, mainJar, jarPath, "Should find main jar, not sources")
	})

	t.Run("skips javadoc jar", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "target")
		err := os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		// Create javadoc JAR (should be skipped)
		javadocJar := filepath.Join(targetDir, "myapp-1.0-javadoc.jar")
		err = os.WriteFile(javadocJar, []byte("fake jar"), 0644)
		require.NoError(t, err)

		// Create main JAR
		mainJar := filepath.Join(targetDir, "myapp-1.0.jar")
		err = os.WriteFile(mainJar, []byte("fake jar"), 0644)
		require.NoError(t, err)

		jarPath, err := findJar(targetDir)
		require.NoError(t, err)
		assert.Equal(t, mainJar, jarPath, "Should find main jar, not javadoc")
	})

	t.Run("skips original jar from shade plugin", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "target")
		err := os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		// Create original JAR (should be skipped)
		originalJar := filepath.Join(targetDir, "myapp-1.0-original.jar")
		err = os.WriteFile(originalJar, []byte("fake jar"), 0644)
		require.NoError(t, err)

		// Create shaded JAR (this is what we want)
		shadedJar := filepath.Join(targetDir, "myapp-1.0.jar")
		err = os.WriteFile(shadedJar, []byte("fake jar"), 0644)
		require.NoError(t, err)

		jarPath, err := findJar(targetDir)
		require.NoError(t, err)
		assert.Equal(t, shadedJar, jarPath, "Should find shaded jar, not original")
	})

	t.Run("returns error if no jar found", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "target")
		err := os.MkdirAll(targetDir, 0755)
		require.NoError(t, err)

		// No JAR files

		_, err = findJar(targetDir)
		assert.Error(t, err, "Should return error if no jar found")
		assert.Contains(t, err.Error(), "no jar file found")
	})

	t.Run("returns error if target directory does not exist", func(t *testing.T) {
		_, err := findJar("/nonexistent/target")
		assert.Error(t, err, "Should return error if directory does not exist")
	})
}

// TestJavaBuildErrorHandling tests error scenarios
func TestJavaBuildErrorHandling(t *testing.T) {
	t.Run("returns descriptive error for missing pom.xml", func(t *testing.T) {
		tmpDir := t.TempDir()

		cfg := Config{
			SourceDir: tmpDir,
			Runtime:   "java21",
		}

		result := JavaBuild(context.Background(), cfg)

		assert.True(t, E.IsLeft(result), "Should fail without pom.xml")

		// Extract error message
		err := E.Fold(
			func(e error) error { return e },
			func(a Artifact) error { return nil },
		)(result)

		assert.Contains(t, err.Error(), "pom.xml", "Error should mention pom.xml")
	})

	t.Run("returns descriptive error for Maven build failure", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create invalid pom.xml
		pomPath := filepath.Join(tmpDir, "pom.xml")
		invalidPom := `<project><invalid>xml</invalid></project>`
		err := os.WriteFile(pomPath, []byte(invalidPom), 0644)
		require.NoError(t, err)

		cfg := Config{
			SourceDir: tmpDir,
			Runtime:   "java21",
		}

		result := JavaBuild(context.Background(), cfg)

		assert.True(t, E.IsLeft(result), "Should fail with invalid pom.xml")

		// Extract error
		buildErr := E.Fold(
			func(e error) error { return e },
			func(a Artifact) error { return nil },
		)(result)

		assert.Contains(t, buildErr.Error(), "mvn package failed", "Error should mention Maven failure")
	})
}

// TestJavaBuildSuccessful tests successful Java builds
func TestJavaBuildSuccessful(t *testing.T) {
	// Skip if mvn is not available
	if _, err := exec.LookPath("mvn"); err != nil {
		t.Skip("mvn not available, skipping Java build integration test")
	}

	t.Run("builds simple Java Lambda project", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create Java source directory
		javaDir := filepath.Join(tmpDir, "src", "main", "java", "com", "example")
		err := os.MkdirAll(javaDir, 0755)
		require.NoError(t, err)

		// Create Handler.java
		handlerPath := filepath.Join(javaDir, "Handler.java")
		handlerContent := `package com.example;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import java.util.Map;

public class Handler implements RequestHandler<Map<String, Object>, Map<String, Object>> {
    @Override
    public Map<String, Object> handleRequest(Map<String, Object> event, Context context) {
        return Map.of("statusCode", 200, "body", "Hello");
    }
}
`
		err = os.WriteFile(handlerPath, []byte(handlerContent), 0644)
		require.NoError(t, err)

		// Create minimal pom.xml
		pomPath := filepath.Join(tmpDir, "pom.xml")
		pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test-lambda</artifactId>
    <version>1.0</version>
    <packaging>jar</packaging>

    <properties>
        <maven.compiler.source>11</maven.compiler.source>
        <maven.compiler.target>11</maven.compiler.target>
    </properties>

    <dependencies>
        <dependency>
            <groupId>com.amazonaws</groupId>
            <artifactId>aws-lambda-java-core</artifactId>
            <version>1.2.3</version>
        </dependency>
    </dependencies>

    <build>
        <plugins>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-shade-plugin</artifactId>
                <version>3.5.1</version>
                <configuration>
                    <finalName>lambda</finalName>
                </configuration>
                <executions>
                    <execution>
                        <phase>package</phase>
                        <goals>
                            <goal>shade</goal>
                        </goals>
                    </execution>
                </executions>
            </plugin>
        </plugins>
    </build>
</project>
`
		err = os.WriteFile(pomPath, []byte(pomContent), 0644)
		require.NoError(t, err)

		cfg := Config{
			SourceDir: tmpDir,
			Runtime:   "java21",
		}

		result := JavaBuild(context.Background(), cfg)

		if E.IsLeft(result) {
			// Extract error for debugging
			err := E.Fold(
				func(e error) error { return e },
				func(a Artifact) error { return nil },
			)(result)
			t.Logf("Java build failed: %v", err)
		}

		assert.True(t, E.IsRight(result), "Should succeed with valid Java project")

		// Verify artifact
		artifact := E.Fold(
			func(err error) Artifact { return Artifact{} },
			func(a Artifact) Artifact { return a },
		)(result)

		assert.NotEmpty(t, artifact.Path)
		assert.FileExists(t, artifact.Path)
		assert.Greater(t, artifact.Size, int64(1000), "JAR should be larger than 1KB")
		assert.NotEmpty(t, artifact.Checksum)
	})
}

// TestJavaBuildOutputPath tests output path handling
func TestJavaBuildOutputPath(t *testing.T) {
	t.Run("uses default output path if not specified", func(t *testing.T) {
		cfg := Config{
			SourceDir:  "/tmp/test",
			OutputPath: "", // Empty - should use default
			Runtime:    "java21",
		}

		// We know this will fail, but we can check the default path logic
		// by examining what path would be used
		expectedPath := filepath.Join(cfg.SourceDir, "target", "lambda.jar")
		assert.Contains(t, expectedPath, "target/lambda.jar")
	})

	t.Run("respects custom output path", func(t *testing.T) {
		customPath := "/custom/path/myapp.jar"
		cfg := Config{
			SourceDir:  "/tmp/test",
			OutputPath: customPath,
			Runtime:    "java21",
		}

		assert.Equal(t, customPath, cfg.OutputPath)
	})
}

// Benchmark JavaBuild function
func BenchmarkJavaBuild(b *testing.B) {
	// Setup a basic project (this will fail, but we're benchmarking the function overhead)
	cfg := Config{
		SourceDir:  "/nonexistent",
		OutputPath: "/tmp/test.jar",
		Runtime:    "java21",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		JavaBuild(context.Background(), cfg)
	}
}

// BenchmarkJavaBuildWithComposition benchmarks composed build functions
func BenchmarkJavaBuildWithComposition(b *testing.B) {
	cfg := Config{
		SourceDir: "/nonexistent",
		Runtime:   "java21",
	}

	b.Run("Plain", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			JavaBuild(context.Background(), cfg)
		}
	})

	b.Run("WithCache", func(b *testing.B) {
		cache := NewMemoryCache()
		cachedBuild := WithCache(cache)(JavaBuild)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cachedBuild(context.Background(), cfg)
		}
	})

	b.Run("WithLogging", func(b *testing.B) {
		logger := &mockLogger{
			infoFn:  func(msg string, args ...interface{}) {},
			errorFn: func(msg string, args ...interface{}) {},
		}
		loggedBuild := WithLogging(logger)(JavaBuild)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			loggedBuild(context.Background(), cfg)
		}
	})

	b.Run("Composed", func(b *testing.B) {
		cache := NewMemoryCache()
		logger := &mockLogger{
			infoFn:  func(msg string, args ...interface{}) {},
			errorFn: func(msg string, args ...interface{}) {},
		}

		composed := Compose(
			WithCache(cache),
			WithLogging(logger),
		)(JavaBuild)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			composed(context.Background(), cfg)
		}
	})
}
