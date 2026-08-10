package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	got, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
	if got != Default() {
		t.Errorf("missing file: got %+v, want defaults %+v", got, Default())
	}
}

func TestLoadPartialFillsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sshblog.yaml")
	if err := os.WriteFile(path, []byte("title: custom.example\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "custom.example" {
		t.Errorf("title not read: %q", got.Title)
	}
	def := Default()
	if got.Description != def.Description || got.Help != def.Help {
		t.Errorf("missing fields should fall back to defaults: %+v", got)
	}
}

func TestLoadTrimsDescriptionTrailingNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sshblog.yaml")
	yaml := "description: |\n  line one\n  line two\n"
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if want := "line one\nline two"; got.Description != want {
		t.Errorf("description = %q, want %q", got.Description, want)
	}
}

// TestRepoConfigMatchesDefaults ensures the checked-in sshblog.yaml reproduces the
// values the blog previously hardcoded.
func TestRepoConfigMatchesDefaults(t *testing.T) {
	got, err := Load("../../sshblog.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if got != Default() {
		t.Errorf("sshblog.yaml drifted from defaults:\n got  %+v\n want %+v", got, Default())
	}
}
