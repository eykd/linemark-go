package acceptance_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runLmk executes the lmk binary and returns stdout, stderr, and exit code.
func runLmk(t *testing.T, dir string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(lmkBinary, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run lmk: %v", err)
		}
	}
	return stdout.String(), stderr.String(), exitCode
}

// runLmkSuccess runs lmk expecting exit code 0 and returns stdout.
func runLmkSuccess(t *testing.T, dir string, args ...string) string {
	t.Helper()
	stdout, stderr, exitCode := runLmk(t, dir, args...)
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d\nargs: %v\nstdout: %s\nstderr: %s", exitCode, args, stdout, stderr)
	}
	return stdout
}

// initProject creates a temp dir and initializes a linemark project.
func initProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runLmkSuccess(t, dir, "init")
	return dir
}

// addNode runs lmk add and returns stdout.
func addNode(t *testing.T, dir, title string, extraArgs ...string) string {
	t.Helper()
	args := append([]string{"add", title}, extraArgs...)
	return runLmkSuccess(t, dir, args...)
}

// addNodeJSON runs lmk add --json and parses the result.
func addNodeJSON(t *testing.T, dir, title string, extraArgs ...string) map[string]interface{} {
	t.Helper()
	args := append([]string{"add", "--json", title}, extraArgs...)
	stdout := runLmkSuccess(t, dir, args...)
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse add JSON: %v\noutput: %s", err, stdout)
	}
	return result
}

// getNodeMP extracts the MP string from an add --json result.
func getNodeMP(t *testing.T, result map[string]interface{}) string {
	t.Helper()
	node, ok := result["node"].(map[string]interface{})
	if !ok {
		t.Fatal("missing node in result")
	}
	mp, ok := node["mp"].(string)
	if !ok {
		t.Fatal("missing mp in node")
	}
	return mp
}

// getNodeSID extracts the SID string from an add --json result.
func getNodeSID(t *testing.T, result map[string]interface{}) string {
	t.Helper()
	node, ok := result["node"].(map[string]interface{})
	if !ok {
		t.Fatal("missing node in result")
	}
	sid, ok := node["sid"].(string)
	if !ok {
		t.Fatal("missing sid in node")
	}
	return sid
}

// listJSON runs lmk list --json and parses the result.
func listJSON(t *testing.T, dir string, extraArgs ...string) map[string]interface{} {
	t.Helper()
	args := append([]string{"list", "--json"}, extraArgs...)
	stdout := runLmkSuccess(t, dir, args...)
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("failed to parse list JSON: %v\noutput: %s", err, stdout)
	}
	return result
}

// flattenNodes recursively collects all nodes from a list --json result into a flat slice.
func flattenNodes(nodes []interface{}) []map[string]interface{} {
	var flat []map[string]interface{}
	for _, n := range nodes {
		node := n.(map[string]interface{})
		flat = append(flat, node)
		if children, ok := node["children"].([]interface{}); ok {
			flat = append(flat, flattenNodes(children)...)
		}
	}
	return flat
}

// getNodes extracts the top-level nodes array from a list --json result.
func getNodes(t *testing.T, result map[string]interface{}) []interface{} {
	t.Helper()
	nodes, ok := result["nodes"].([]interface{})
	if !ok {
		t.Fatal("missing nodes in result")
	}
	return nodes
}

// listMDFiles returns all .md files in the directory.
func listMDFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read dir: %v", err)
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			files = append(files, e.Name())
		}
	}
	return files
}

// writeFile creates a file with the given content.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create parent dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}

// readFile reads a file's content.
func readFile(t *testing.T, dir, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	return string(content)
}

// removeFile deletes a file.
func removeFile(t *testing.T, dir, name string) {
	t.Helper()
	if err := os.Remove(filepath.Join(dir, name)); err != nil {
		t.Fatalf("failed to remove file: %v", err)
	}
}

// fileExists checks if a file exists.
func fileExists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

// snapshotFiles returns a map of filename → content for all .md files in dir.
func snapshotFiles(t *testing.T, dir string) map[string]string {
	t.Helper()
	files := listMDFiles(t, dir)
	snap := make(map[string]string, len(files))
	for _, f := range files {
		snap[f] = readFile(t, dir, f)
	}
	return snap
}

// assertSnapshotUnchanged compares current .md files against a snapshot.
func assertSnapshotUnchanged(t *testing.T, dir string, snap map[string]string) {
	t.Helper()
	current := snapshotFiles(t, dir)
	if len(current) != len(snap) {
		t.Fatalf("file count changed: had %d, now %d", len(snap), len(current))
	}
	for name, oldContent := range snap {
		newContent, ok := current[name]
		if !ok {
			t.Fatalf("file %s disappeared", name)
		}
		if newContent != oldContent {
			t.Fatalf("file %s content changed", name)
		}
	}
}
