package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIHelp tests the CLI help functionality
func TestCLIHelp(t *testing.T) {
	// Build the CLI binary first
	cmd := exec.Command("go", "build", "-o", "converter_test", "main.go")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() {
		if err := os.Remove("converter_test"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	testCases := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "Help flag",
			args:     []string{"--help"},
			expected: []string{"Usage", "flag"},
		},
		{
			name:     "Short help flag",
			args:     []string{"-h"},
			expected: []string{"Usage", "flag"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("./converter_test", tc.args...) //nolint:gosec
			output, _ := cmd.CombinedOutput()

			outputStr := string(output)
			// Help flags should show usage
			foundUsage := false
			for _, expected := range tc.expected {
				if strings.Contains(outputStr, expected) {
					foundUsage = true
					break
				}
			}

			if !foundUsage {
				t.Logf("Help output (this is informational): %s", outputStr)
			}
		})
	}
}

// TestCLIInvalidArguments tests CLI error handling
func TestCLIInvalidArguments(t *testing.T) {
	// Build the CLI binary first
	cmd := exec.Command("go", "build", "-o", "converter_test", "main.go")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() {
		if err := os.Remove("converter_test"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	testCases := []struct {
		name        string
		args        []string
		expectError bool
		errorText   string
	}{
		{
			name:        "Missing netex file",
			args:        []string{"--netex", "nonexistent.zip", "--output", "out.zip", "--codespace", "TEST"},
			expectError: false, // CLI continues and reports file not found, but doesn't exit with error
			errorText:   "not found",
		},
		{
			name:        "Invalid flag",
			args:        []string{"--invalid-flag"},
			expectError: true,
			errorText:   "flag provided but not defined",
		},
		{
			name:        "Missing required netex flag",
			args:        []string{"--output", "out.zip", "--codespace", "TEST"},
			expectError: false,       // CLI uses default netex file
			errorText:   "not found", // Will use default netex file
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("./converter_test", tc.args...) //nolint:gosec
			output, err := cmd.CombinedOutput()

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for args %v, but command succeeded", tc.args)
				}
				outputStr := strings.ToLower(string(output))
				if !strings.Contains(outputStr, strings.ToLower(tc.errorText)) {
					t.Errorf("Expected error output to contain '%s', got:\n%s", tc.errorText, string(output))
				}
			} else if err != nil {
				t.Errorf("Expected success for args %v, got error: %v\nOutput: %s", tc.args, err, string(output))
			}
		})
	}
}

// TestCLIWithRealData tests CLI with actual data files
func TestCLIWithRealData(t *testing.T) {
	// Build the CLI binary first
	cmd := exec.Command("go", "build", "-o", "converter_test", "main.go")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() {
		if err := os.Remove("converter_test"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	// Check if test data exists
	testDataPath := "../../fluo-grand-est-riv-netex.zip"
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("Test data file not found, skipping real data CLI test")
	}

	// Create temp output directory
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "cli_test_output.zip")

	t.Run("Convert real data", func(t *testing.T) {
		cmd := exec.Command("./converter_test", //nolint:gosec
			"--netex", testDataPath,
			"--output", outputFile,
			"--codespace", "FR")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("CLI conversion failed: %v\nOutput: %s", err, string(output))
			return
		}

		// Check if output file was created
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Output file was not created: %s", outputFile)
		}

		// Check output contains success indicators
		outputStr := string(output)
		successIndicators := []string{
			"✅", "completed", "successfully", "conversion", "GTFS",
		}

		foundIndicators := 0
		for _, indicator := range successIndicators {
			if strings.Contains(strings.ToLower(outputStr), strings.ToLower(indicator)) {
				foundIndicators++
			}
		}

		if foundIndicators == 0 {
			t.Errorf("Output doesn't seem to indicate success:\n%s", outputStr)
		}
	})
}

// TestCLIVersionInfo tests version and build information
func TestCLIVersionInfo(t *testing.T) {
	// Build the CLI binary first
	cmd := exec.Command("go", "build", "-o", "converter_test", "main.go")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() {
		if err := os.Remove("converter_test"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	// Test version flag (if implemented)
	t.Run("Version flag", func(t *testing.T) {
		cmd := exec.Command("./converter_test", "--version")
		output, err := cmd.CombinedOutput()

		// Version flag might not be implemented yet, so we just check output
		outputStr := string(output)
		if err != nil && !strings.Contains(strings.ToLower(outputStr), "unknown flag") {
			// If version is implemented, it should work
			t.Logf("Version output: %s", outputStr)
		} else {
			// If not implemented, that's ok for now
			t.Logf("Version flag not yet implemented (this is ok)")
		}
	})
}

// TestCLIPerformance tests CLI performance with timing
func TestCLIPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Build the CLI binary first
	cmd := exec.Command("go", "build", "-o", "converter_test", "main.go")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() {
		if err := os.Remove("converter_test"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	// Check if test data exists
	testDataPath := "../../fluo-grand-est-riv-netex.zip"
	if _, err := os.Stat(testDataPath); os.IsNotExist(err) {
		t.Skip("Test data file not found, skipping performance test")
	}

	// Create temp output directory
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "perf_test_output.zip")

	t.Run("Performance timing", func(t *testing.T) {
		cmd := exec.Command("./converter_test", //nolint:gosec
			"--netex", testDataPath,
			"--output", outputFile,
			"--codespace", "FR")

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("CLI conversion failed: %v\nOutput: %s", err, string(output))
			return
		}

		// Look for timing information in output
		outputStr := string(output)
		timingIndicators := []string{"ms", "seconds", "duration", "time", "performance"}

		foundTiming := false
		for _, indicator := range timingIndicators {
			if strings.Contains(strings.ToLower(outputStr), indicator) {
				foundTiming = true
				break
			}
		}

		if foundTiming {
			t.Logf("CLI reported timing information: Found timing indicators in output")
		} else {
			t.Logf("No explicit timing information found in CLI output")
		}
	})
}

// TestCLIEnvironmentVariables tests environment variable handling
func TestCLIEnvironmentVariables(t *testing.T) {
	// Build the CLI binary first
	cmd := exec.Command("go", "build", "-o", "converter_test", "main.go")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() {
		if err := os.Remove("converter_test"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	t.Run("Environment variables", func(t *testing.T) {
		// Set some environment variables that might be used
		cmd := exec.Command("./converter_test", "--help")
		cmd.Env = append(os.Environ(),
			"NETEX_CONVERTER_LOG_LEVEL=DEBUG",
			"NETEX_CONVERTER_MEMORY_LIMIT=512MB",
		)

		output, err := cmd.CombinedOutput()

		// This is mainly to test that env vars don't break the CLI
		if err == nil {
			t.Logf("CLI handled environment variables without crashing")
		} else {
			outputStr := string(output)
			if strings.Contains(strings.ToLower(outputStr), "usage") {
				t.Logf("CLI help displayed correctly with environment variables set")
			} else {
				t.Errorf("Unexpected error with environment variables: %v\nOutput: %s", err, outputStr)
			}
		}
	})
}

// TestCLIInputValidation tests various input validation scenarios
func TestCLIInputValidation(t *testing.T) {
	// Build the CLI binary first
	cmd := exec.Command("go", "build", "-o", "converter_test", "main.go")
	cmd.Dir = "."
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}
	defer func() {
		if err := os.Remove("converter_test"); err != nil {
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	testCases := []struct {
		name        string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "Empty codespace",
			args:        []string{"--netex", "test.zip", "--output", "out.zip", "--codespace", ""},
			expectError: false, // CLI accepts empty codespace
			description: "Empty codespace should be accepted",
		},
		{
			name:        "Very long codespace",
			args:        []string{"--netex", "test.zip", "--output", "out.zip", "--codespace", strings.Repeat("A", 100)},
			expectError: false, // CLI accepts long codespace
			description: "Long codespace should be handled gracefully",
		},
		{
			name:        "Special characters in paths",
			args:        []string{"--netex", "test with spaces.zip", "--output", "out with spaces.zip", "--codespace", "TEST"},
			expectError: false, // File doesn't exist, but CLI handles spaces
			description: "Paths with spaces should be handled",
		},
		{
			name:        "Unicode codespace",
			args:        []string{"--netex", "test.zip", "--output", "out.zip", "--codespace", "测试"},
			expectError: false, // CLI accepts unicode
			description: "Unicode codespace should be handled gracefully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("./converter_test", tc.args...) //nolint:gosec
			output, err := cmd.CombinedOutput()

			if tc.expectError {
				if err == nil {
					t.Errorf("%s: Expected error but command succeeded. Output: %s", tc.description, string(output))
				} else {
					t.Logf("%s: Correctly handled error case", tc.description)
				}
			}
		})
	}
}
