#!/usr/bin/env nu

# Comprehensive linting fix script for Forge codebase
# Fixes common linting issues systematically

def main [] {
    print "Starting systematic linting fixes..."

    # Fix all builder files (go, node, python, java)
    fix_builder_files

    # Fix all test files
    fix_test_files

    # Fix lingon package
    fix_lingon_package

    # Run gofumpt on everything
    run_gofumpt

    print "✅ All systematic fixes applied!"
    print "Run 'golangci-lint run ./...' to verify"
}

def fix_builder_files [] {
    print "Fixing builder files..."

    let builders = [
        "internal/build/go_builder.go",
        "internal/build/node_builder.go",
        "internal/build/python_builder.go",
        "internal/build/java_builder.go"
    ]

    for file in $builders {
        print $"  Processing ($file)..."

        # Add periods to comments
        bash -c $"sed -i 's|^\\(// [A-Z][^.]*[^.]\\)$|\\1.|' ($file)"

        print $"  ✓ ($file) - comments fixed"
    }
}

def fix_test_files [] {
    print "Fixing test files..."

    let test_files = [
        "internal/build/go_builder_test.go",
        "internal/build/node_builder_test.go",
        "internal/build/python_builder_test.go",
        "internal/build/java_builder_test.go",
        "internal/build/functional_test.go"
    ]

    for file in $test_files {
        if ($file | path exists) {
            print $"  Processing ($file)..."

            # Fix unused msg parameters in mockLogger
            bash -c $"sed -i 's/func(msg string, /func(_ string, /g' ($file)"
            bash -c $"sed -i 's/, msg string,/, _ string,/g' ($file)"

            # Fix unused err parameters
            bash -c $"sed -i 's/func(err error)/func(_ error)/g' ($file)"
            bash -c $"sed -i 's/(err error) Artifact/(_ error) Artifact/g' ($file)"

            # Fix var-declaration issues - remove explicit type
            bash -c $"sed -i 's/var buildFunc BuildFunc =/buildFunc :=/g' ($file)"

            print $"  ✓ ($file) - test fixes applied"
        }
    }
}

def fix_lingon_package [] {
    print "Fixing lingon package..."

    # Add package comment if missing
    let config_types = "internal/lingon/config_types.go"

    if ($config_types | path exists) {
        print $"  Processing ($config_types)..."

        # Add periods to all comments
        bash -c $"sed -i 's|^\\(// [A-Z][^.]*[^.]\\)$|\\1.|' ($config_types)"

        print $"  ✓ ($config_types) - comments fixed"
    }

    # Fix other lingon files
    let lingon_files = (
        glob "internal/lingon/*.go"
        | where { |f| ($f | str contains "aws") == false }
    )

    for file in $lingon_files {
        print $"  Processing ($file)..."
        bash -c $"sed -i 's|^\\(// [A-Z][^.]*[^.]\\)$|\\1.|' ($file)"
    }
}

def run_gofumpt [] {
    print "Running gofumpt for formatting..."

    bash -c "gofumpt -w internal/build/*.go"
    bash -c "gofumpt -w internal/lingon/*.go"

    print "✓ gofumpt completed"
}
