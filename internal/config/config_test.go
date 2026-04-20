package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FileNotExist(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg != nil {
		t.Fatalf("expected nil config, got %+v", cfg)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()

	want := Config{ProjectNumber: 42}
	if err := Save(dir, want); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected config, got nil")
	}
	if got.ProjectNumber != want.ProjectNumber {
		t.Errorf("ProjectNumber = %d, want %d", got.ProjectNumber, want.ProjectNumber)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, FileName)
	if err := os.WriteFile(path, []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	cfg, err := Load(dir)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if cfg != nil {
		t.Fatalf("expected nil config on error, got %+v", cfg)
	}
}

func TestSave_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	cfg := Config{ProjectNumber: 7}

	if err := Save(dir, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	path := filepath.Join(dir, FileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected config file to be created")
	}
}

func TestSave_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()

	if err := Save(dir, Config{ProjectNumber: 1}); err != nil {
		t.Fatalf("initial Save failed: %v", err)
	}
	if err := Save(dir, Config{ProjectNumber: 99}); err != nil {
		t.Fatalf("overwrite Save failed: %v", err)
	}

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if got.ProjectNumber != 99 {
		t.Errorf("ProjectNumber = %d, want 99", got.ProjectNumber)
	}
}
