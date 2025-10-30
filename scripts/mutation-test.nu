#!/usr/bin/env nu

# Mutation Testing Script for Forge
# Tests the quality of test suites by introducing mutations (bugs) and checking if tests catch them
#
# This script follows production-ready Nushell patterns:
# - Pure functional data transformations
# - Explicit type signatures
# - Railway-oriented error handling
# - Streaming pipeline architecture
# - Parallel execution with par-each
# - Comprehensive timeout and error handling

# ============================================================================
# Type Definitions (for documentation)
# ============================================================================

# MutationResult represents the outcome of mutation testing for a single package
# record<
#   package: string,      # Package path (e.g., internal/build)
#   score: float,         # Mutation score (0.0 - 1.0)
#   passed: int,          # Number of mutations killed
#   failed: int,          # Number of mutations that survived
#   total: int,           # Total mutations generated
#   skipped: int,         # Number of mutations skipped
#   duplicated: int,      # Number of duplicate mutations
#   error: string         # Error type: "" | "timeout" | "parse_error" | "no_mutations" | "execution_error"
# >

# ============================================================================
# Pure Functions (Calculations)
# ============================================================================

# Parse mutation testing output to extract results
# Type: string -> Result<MutationResult, string>
def parse_mutation_output [
    pkg: string          # Package being tested
    stdout: string       # Standard output from go-mutesting
]: nothing -> record {
    let lines = ($stdout | lines)

    # Look for summary line: "The mutation score is 0.857143 (12 passed, 2 failed, 0 duplicated, 0 skipped, total is 14)"
    let summary = ($lines | where { |line| $line =~ "mutation score" })

    if ($summary | is-empty) {
        return {
            package: $pkg
            score: 0.0
            passed: 0
            failed: 0
            total: 0
            skipped: 0
            duplicated: 0
            error: "no_mutations"
        }
    }

    let summary_line = ($summary | first)

    # Parse score with error handling
    let score = (
        try {
            $summary_line
            | parse "mutation score is {score}"
            | get score
            | first
            | into float
        } catch {
            0.0
        }
    )

    # Parse counts with error handling
    let passed = (
        try {
            $summary_line
            | parse "{passed} passed"
            | get passed
            | first
            | str trim
            | into int
        } catch {
            0
        }
    )

    let failed = (
        try {
            $summary_line
            | parse "{failed} failed"
            | get failed
            | first
            | str trim
            | into int
        } catch {
            0
        }
    )

    let duplicated = (
        try {
            $summary_line
            | parse "{duplicated} duplicated"
            | get duplicated
            | first
            | str trim
            | into int
        } catch {
            0
        }
    )

    let skipped = (
        try {
            $summary_line
            | parse "{skipped} skipped"
            | get skipped
            | first
            | str trim
            | into int
        } catch {
            0
        }
    )

    let total = (
        try {
            $summary_line
            | parse "total is {total}"
            | get total
            | first
            | str trim
            | into int
        } catch {
            0
        }
    )

    {
        package: $pkg
        score: $score
        passed: $passed
        failed: $failed
        total: $total
        skipped: $skipped
        duplicated: $duplicated
        error: ""
    }
}

# Calculate overall mutation score from package results
# Type: list<MutationResult> -> float
def calculate_overall_score []: list -> float {
    let results = $in

    # Filter out error results
    let valid_results = ($results | where error == "")

    if ($valid_results | is-empty) {
        return 0.0
    }

    let total_mutations = ($valid_results | reduce -f 0 { |it, acc| $acc + $it.total })
    let killed_mutations = ($valid_results | reduce -f 0 { |it, acc| $acc + $it.passed })

    if $total_mutations > 0 {
        $killed_mutations / $total_mutations
    } else {
        0.0
    }
}

# Determine status emoji based on mutation score
# Type: float -> string
def get_status_emoji []: float -> string {
    let score = $in

    if $score >= 0.90 {
        "✅"
    } else if $score >= 0.80 {
        "⚠️"
    } else if $score >= 0.70 {
        "⚠️"
    } else {
        "❌"
    }
}

# Format score as percentage with 1 decimal place
# Type: float -> string
def format_score []: float -> string {
    let score = $in
    $score * 100 | math round -p 1 | $"($in)%"
}

# Generate progress bar for score visualization
# Type: float -> string
def generate_progress_bar [
    width: int = 50     # Width of progress bar in characters
]: float -> string {
    let score = $in
    let bar_length = ($score * $width | math round)
    let filled = (seq 1 $bar_length | each { "█" } | str join)
    let empty_length = ($width - $bar_length)
    let empty = (seq 1 $empty_length | each { "░" } | str join)
    $filled + $empty
}

# ============================================================================
# Actions (I/O Operations)
# ============================================================================

# Run mutation testing on a single package with timeout
# Type: string -> int -> bool -> MutationResult
def run_package_mutation_test [
    pkg: string          # Package path
    timeout_secs: int    # Timeout in seconds
    verbose: bool        # Verbose output flag
]: nothing -> record {
    if $verbose {
        print $"[($pkg)] Starting mutation testing..."
    }

    # Run go-mutesting with timeout
    let output = (
        ^timeout $timeout_secs go-mutesting --do-not-remove-tmp-folder $pkg
        | complete
    )

    # Handle timeout (exit code 124 from timeout command)
    if $output.exit_code == 124 {
        let hours = ($timeout_secs / 3600 | math round -p 1)
        print $"[($pkg)] ⏱️  TIMEOUT after ($timeout_secs)s \(($hours)h\)"
        return {
            package: $pkg
            score: 0.0
            passed: 0
            failed: 0
            total: 0
            skipped: 0
            duplicated: 0
            error: "timeout"
        }
    }

    # Handle execution errors
    if $output.exit_code != 0 {
        if $verbose {
            print $"[($pkg)] ❌ ERROR: ($output.stderr)"
        } else {
            print $"[($pkg)] ❌ Execution failed"
        }
        return {
            package: $pkg
            score: 0.0
            passed: 0
            failed: 0
            total: 0
            skipped: 0
            duplicated: 0
            error: "execution_error"
        }
    }

    # Parse output
    let result = (parse_mutation_output $pkg $output.stdout)

    # Print immediate feedback
    if $result.error == "no_mutations" {
        print $"[($pkg)] ⚠️  No mutations generated"
    } else if $result.error == "" {
        let emoji = ($result.score | get_status_emoji)
        let score_str = ($result.score | format_score)
        print $"[($pkg)] ($emoji) Score: ($score_str) \(($result.passed)/($result.total) mutations killed\)"
    }

    $result
}

# ============================================================================
# Report Generation (Pure Presentation Logic)
# ============================================================================

# Print summary header
def print_summary_header []: nothing -> nothing {
    print "=================================="
    print "📊 Overall Mutation Test Results"
    print "=================================="
    print ""
}

# Print overall statistics
def print_overall_stats [
    total_mutations: int
    killed_mutations: int
    overall_score: float
]: nothing -> nothing {
    print $"Total Mutations:  ($total_mutations)"
    print $"Killed Mutations: ($killed_mutations)"
    print $"Overall Score:    (($overall_score | format_score))"
    print ""
}

# Print per-package breakdown table
def print_package_breakdown []: list -> nothing {
    let results = $in

    print "Per-Package Breakdown:"
    print ""

    # Sort by score descending for better readability
    let sorted_results = ($results | sort-by score -r)

    for result in $sorted_results {
        let bar = ($result.score | generate_progress_bar 40)
        let score_str = ($result.score | format_score)
        let emoji = ($result.score | get_status_emoji)

        # Handle error cases
        if $result.error != "" {
            let error_label = match $result.error {
                "timeout" => "TIMEOUT"
                "no_mutations" => "NO MUTATIONS"
                "execution_error" => "ERROR"
                "parse_error" => "PARSE ERROR"
                _ => "UNKNOWN"
            }
            print $"  ($emoji) ($result.package | fill -a right -w 35) [($error_label)]"
        } else {
            print $"  ($emoji) ($result.package | fill -a right -w 35) ($bar) ($score_str)"
        }
    }

    print ""
}

# Print final verdict and exit with appropriate code
def print_verdict_and_exit [
    overall_score: float
]: nothing -> nothing {
    if $overall_score >= 0.85 {
        print "✅ EXCELLENT: Test suite catches 85%+ of mutations"
        exit 0
    } else if $overall_score >= 0.75 {
        print "⚠️  GOOD: Test suite catches 75%+ of mutations (aim for 85%)"
        exit 0
    } else if $overall_score >= 0.65 {
        print "⚠️  FAIR: Test suite catches 65%+ of mutations (aim for 85%)"
        exit 1
    } else {
        print "❌ NEEDS IMPROVEMENT: Test suite catches <65% of mutations"
        exit 1
    }
}

# ============================================================================
# Main Entry Point
# ============================================================================

def main [
    --package (-p): string = ""     # Specific package to test (e.g., internal/build)
    --verbose (-v)                  # Verbose output (show detailed mutation results)
    --parallel (-j): int = 4        # Number of parallel jobs (default 4)
    --timeout (-t): int = 28800     # Timeout per package in seconds (default 8 hours)
]: nothing -> nothing {

    # Print header
    print "🧬 Mutation Testing for Forge"
    print "=================================="
    print $"Parallel jobs: ($parallel)"
    let timeout_hours = ($timeout / 3600 | math round -p 1)
    print $"Timeout:       ($timeout)s \(($timeout_hours) hours per package\)"
    print ""

    # Define all packages to test (39 total)
    let all_packages = [
        # Core build system
        "internal/build"

        # CLI layer
        "internal/cli"

        # Configuration
        "internal/config"

        # Discovery and scaffolding
        "internal/discovery"
        "internal/scaffold"

        # Pipeline orchestration
        "internal/pipeline"

        # Terraform integration
        "internal/terraform"

        # State management
        "internal/state"

        # Generators
        "internal/generators"
        "internal/generators/dynamodb"
        "internal/generators/python"
        "internal/generators/s3"
        "internal/generators/sns"
        "internal/generators/sqs"

        # Lingon integration (exclude AWS generated code)
        "internal/lingon"

        # Type-safe Terraform modules (24 modules)
        "internal/tfmodules"
        "internal/tfmodules/apigatewayv2"
        "internal/tfmodules/appconfig"
        "internal/tfmodules/appsync"
        "internal/tfmodules/cloudfront"
        "internal/tfmodules/dynamodb"
        "internal/tfmodules/eventbridge"
        "internal/tfmodules/lambda"
        "internal/tfmodules/s3"
        "internal/tfmodules/secretsmanager"
        "internal/tfmodules/sns"
        "internal/tfmodules/sqs"
        "internal/tfmodules/ssm"
        "internal/tfmodules/stepfunctions"

        # UI components
        "internal/ui"
    ]

    # Select packages to test
    let packages = if $package != "" {
        [$package]
    } else {
        $all_packages
    }

    print $"Testing ($packages | length) packages in parallel..."
    print ""

    # Run mutation tests in parallel
    # Using par-each for true parallel execution with streaming results
    let results = (
        $packages
        | par-each -t $parallel { |pkg|
            run_package_mutation_test $pkg $timeout $verbose
        }
    )

    # Calculate overall metrics (pure functional reduction)
    let overall_score = ($results | calculate_overall_score)

    let valid_results = ($results | where error == "")
    let total_mutations = ($valid_results | reduce -f 0 { |it, acc| $acc + $it.total })
    let killed_mutations = ($valid_results | reduce -f 0 { |it, acc| $acc + $it.passed })

    # Print summary report
    print ""
    print_summary_header
    print_overall_stats $total_mutations $killed_mutations $overall_score
    ($results | print_package_breakdown)

    # Exit with appropriate code based on score
    print_verdict_and_exit $overall_score
}
