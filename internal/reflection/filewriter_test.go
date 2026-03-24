package reflection

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/phuc/coding-agent-reflection/internal/model"
)

func testReflection() model.Reflection {
	return model.Reflection{
		ID:            uuid.New(),
		Date:          time.Date(2026, 3, 24, 0, 0, 0, 0, time.UTC),
		Summary:       "Did good work.",
		ShouldDo:      "Keep it up.",
		ShouldNotDo:   "Avoid shortcuts.",
		ConfigChanges: "none",
		CreatedAt:     time.Now(),
	}
}

func TestWriteReflectionFile_FirstFile(t *testing.T) {
	dir := t.TempDir()
	r := testReflection()

	path, err := WriteReflectionFile(dir, r.Date, r, 5)
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(path) != "20260324-001.md" {
		t.Errorf("expected 20260324-001.md, got %s", filepath.Base(path))
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "date: 2026-03-24") {
		t.Error("missing frontmatter date")
	}
	if !strings.Contains(string(content), "Did good work.") {
		t.Error("missing summary content")
	}
}

func TestWriteReflectionFile_SecondFile(t *testing.T) {
	dir := t.TempDir()
	r := testReflection()

	WriteReflectionFile(dir, r.Date, r, 5)
	path2, err := WriteReflectionFile(dir, r.Date, r, 3)
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(path2) != "20260324-002.md" {
		t.Errorf("expected 20260324-002.md, got %s", filepath.Base(path2))
	}
}

func TestWriteReflectionFile_CreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "dir")
	r := testReflection()

	path, err := WriteReflectionFile(dir, r.Date, r, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file not created")
	}
}
