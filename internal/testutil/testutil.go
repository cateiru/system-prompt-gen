package testutil

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TempDir creates a temporary directory for testing
func TempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "system-prompt-gen-test-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// CreateTestFile creates a test file with given content
func CreateTestFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file %s: %v", path, err)
	}
}

// ReadTestFile reads content from a test file
func ReadTestFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read test file %s: %v", path, err)
	}
	return string(content)
}

// AssertFileExists checks if a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("expected file %s to exist, but it doesn't", path)
	}
}

// AssertFileNotExists checks if a file doesn't exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected file %s to not exist, but it does", path)
	}
}

// CopyDir recursively copies a directory
func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// NormalizeLineEndings normalizes line endings for cross-platform testing
func NormalizeLineEndings(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\r", "\n")
}

// SetupTestEnv sets up environment variables for testing
func SetupTestEnv(t *testing.T, vars map[string]string) {
	t.Helper()
	original := make(map[string]string)

	for key, value := range vars {
		if orig, exists := os.LookupEnv(key); exists {
			original[key] = orig
		}
		os.Setenv(key, value)
	}

	t.Cleanup(func() {
		for key := range vars {
			if orig, exists := original[key]; exists {
				os.Setenv(key, orig)
			} else {
				os.Unsetenv(key)
			}
		}
	})
}

// StderrCapture captures stderr output for testing
func StderrCapture(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	origStderr := os.Stderr
	os.Stderr = w

	done := make(chan string)
	go func() {
		defer r.Close()
		output, _ := io.ReadAll(r)
		done <- string(output)
	}()

	fn()
	w.Close()
	os.Stderr = origStderr

	return <-done
}

// GetTestDataPath returns the path to test data files
func GetTestDataPath(filename string) string {
	return filepath.Join("../../testdata", filename)
}

// CompareFiles compares two files and returns error if different
func CompareFiles(t *testing.T, expected, actual string) {
	t.Helper()
	expectedContent := ReadTestFile(t, expected)
	actualContent := ReadTestFile(t, actual)

	expectedNorm := NormalizeLineEndings(strings.TrimSpace(expectedContent))
	actualNorm := NormalizeLineEndings(strings.TrimSpace(actualContent))

	if expectedNorm != actualNorm {
		t.Errorf("files differ:\nExpected:\n%s\n\nActual:\n%s", expectedNorm, actualNorm)
	}
}
