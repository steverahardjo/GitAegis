package tests

import(
	"testing"
	"path/filepath"
	//frontend  "github.com/steverahardjo/GitAegis/frontend"
	"os"
)

func setupTestDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "gitaegis-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	os.WriteFile(filepath.Join(tmpDir, "secret.go"), []byte(`password = "1234"`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "safe.txt"), []byte(`nothing sensitive`), 0644)

	return tmpDir
}