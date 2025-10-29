#!/usr/bin/env nu

# Mutation Testing Script for Forge
# Tests the quality of test suites by introducing mutations (bugs) and checking if tests catch them

def main [
    --package (-p): string = ""     # Specific package to test (e.g., internal/build)
    --verbose (-v)                   # Verbose output
] {
    print "🧬 Running Mutation Testing on Forge"
    print "=================================="
    print ""

    # Packages to test (exclude generated AWS resources)
    let packages = if $package != "" {
        [$package]
    } else {
        [
            "internal/build"
            "internal/config"
            "internal/stack"
            "internal/pipeline"
            "internal/terraform"
            "internal/scaffold"
            "internal/lingon"
        ]
    }

    mut total_mutations = 0
    mut killed_mutations = 0
    mut results = []

    for pkg in $packages {
        print $"Testing package: ($pkg)"
        print "---"

        # Run go-mutesting on the package
        let output = (
            go-mutesting --do-not-remove-tmp-folder $pkg
            | complete
        )

        if $verbose {
            print $output.stdout
        }

        # Parse results from output
        let lines = ($output.stdout | lines)

        # Look for summary line like "The mutation score is 0.857143 (12 passed, 2 failed, 0 duplicated, 0 skipped, total is 14)"
        let summary = ($lines | where { |line| $line =~ "mutation score" } | first)

        if ($summary | is-empty) {
            print $"  ⚠️  No mutations generated for ($pkg)"
            continue
        }

        # Extract score
        let score_match = ($summary | parse "mutation score is {score}" | get score | first)
        let score = ($score_match | into float)

        # Extract counts
        let passed = ($summary | parse "{passed} passed" | get passed | first | into int)
        let failed = ($summary | parse "{failed} failed" | get failed | first | into int)
        let total = ($summary | parse "total is {total}" | get total | first | into int)

        $total_mutations = $total_mutations + $total
        $killed_mutations = $killed_mutations + $passed

        let emoji = if $score >= 0.90 {
            "✅"
        } else if $score >= 0.80 {
            "⚠️ "
        } else {
            "❌"
        }

        print $"  ($emoji) Score: ($score * 100 | math round -p 1)% \(($passed)/($total) mutations killed\)"

        $results = ($results | append {
            package: $pkg
            score: $score
            passed: $passed
            failed: $failed
            total: $total
        })

        print ""
    }

    print "=================================="
    print "📊 Overall Mutation Test Results"
    print "=================================="
    print ""

    let overall_score = if $total_mutations > 0 {
        $killed_mutations / $total_mutations
    } else {
        0.0
    }

    print $"Total Mutations: ($total_mutations)"
    print $"Killed Mutations: ($killed_mutations)"
    print $"Overall Score: ($overall_score * 100 | math round -p 1)%"
    print ""

    # Show per-package breakdown
    print "Per-Package Breakdown:"
    for result in $results {
        let bar_length = ($result.score * 50 | math round)
        let bar = ("█" | str repeat $bar_length)
        print $"  ($result.package | fill -a right -w 30) ($bar) ($result.score * 100 | math round -p 1)%"
    }

    print ""

    if $overall_score >= 0.85 {
        print "✅ EXCELLENT: Test suite catches 85%+ of mutations"
        exit 0
    } else if $overall_score >= 0.75 {
        print "⚠️  GOOD: Test suite catches 75%+ of mutations (aim for 85%)"
        exit 0
    } else {
        print "❌ NEEDS IMPROVEMENT: Test suite catches < 75% of mutations"
        exit 1
    }
}
