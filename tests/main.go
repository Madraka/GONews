package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/joho/godotenv"
)

var (
	testType = flag.String("type", "all", "Type of tests to run: unit, integration, e2e, all")
	verbose  = flag.Bool("v", false, "Verbose output")
	coverage = flag.Bool("coverage", false, "Generate coverage report")
	race     = flag.Bool("race", false, "Enable race detection")
)

func main() {
	flag.Parse()

	// Load test environment variables - try local first, then default
	err := godotenv.Load(".env.test.local")
	if err != nil {
		// If local config doesn't exist, try default test config
		if err2 := godotenv.Load(".env.test"); err2 != nil {
			fmt.Printf("Warning: Could not load test environment files: %v, %v\n", err, err2)
			fmt.Println("Make sure you have either .env.test or .env.test.local configured")
		} else {
			fmt.Println("Using Docker test configuration (.env.test)")
		}
	} else {
		fmt.Println("Using local test configuration (.env.test.local)")
	}

	// Set test environment
	if err := os.Setenv("GIN_MODE", "test"); err != nil {
		fmt.Printf("Warning: Failed to set GIN_MODE: %v\n", err)
	}
	if err := os.Setenv("TEST_ENV", "true"); err != nil {
		fmt.Printf("Warning: Failed to set TEST_ENV: %v\n", err)
	}

	// Configure test flags
	testArgs := []string{"test"}

	if *verbose {
		testArgs = append(testArgs, "-v")
	}

	if *coverage {
		testArgs = append(testArgs, "-coverprofile=coverage.out")
	}

	if *race {
		testArgs = append(testArgs, "-race")
	}

	// Determine which tests to run
	var testPaths []string
	switch *testType {
	case "unit":
		testPaths = []string{"./tests/unit/..."}
	case "integration":
		testPaths = []string{"./tests/integration/..."}
	case "e2e":
		testPaths = []string{"./tests/e2e/..."}
	case "all":
		testPaths = []string{"./tests/unit/...", "./tests/integration/...", "./tests/e2e/..."}
	default:
		fmt.Printf("Invalid test type: %s. Use: unit, integration, e2e, or all\n", *testType)
		os.Exit(1)
	}

	fmt.Printf("Running %s tests...\n", *testType)

	// Run tests for each path
	allPassed := true
	for _, path := range testPaths {
		fmt.Printf("\n=== Running tests in %s ===\n", path)

		// Prepare the go test command
		args := append(testArgs, path)
		cmd := exec.Command("go", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = ".." // Run from the project root

		// Run the tests
		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ Tests failed in %s: %v\n", path, err)
			allPassed = false
		} else {
			fmt.Printf("✅ Tests passed in %s\n", path)
		}
	}

	if allPassed {
		fmt.Println("\n✅ All tests passed!")
		os.Exit(0)
	} else {
		fmt.Println("\n❌ Some tests failed!")
		os.Exit(1)
	}
}

// Print usage information
func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nTest Runner for News API\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -type=unit                 # Run unit tests\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -type=integration -v       # Run integration tests with verbose output\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -type=all -coverage        # Run all tests with coverage\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -type=e2e -race            # Run e2e tests with race detection\n", os.Args[0])
	}
}
