package vcs

import (
	"context"
	"errors"
	"testing"

	"github.com/manningc33/journal/internal/run"
)

func TestCommitStagesAllThenCommits(t *testing.T) {
	fake := &run.Fake{}
	if err := Commit(context.Background(), fake, "/repo", "2026-06-28 hi", "/repo/a.md", "/repo/b.md"); err != nil {
		t.Fatal(err)
	}
	if len(fake.Calls) != 2 {
		t.Fatalf("calls = %v", fake.Calls)
	}
	add, commit := fake.Calls[0], fake.Calls[1]
	if add.Name != "git" || add.Args[0] != "-C" || add.Args[2] != "add" {
		t.Errorf("add call = %+v", add)
	}
	if add.Args[3] != "/repo/a.md" || add.Args[4] != "/repo/b.md" {
		t.Errorf("expected both files staged, got %+v", add.Args)
	}
	if commit.Args[2] != "commit" || commit.Args[4] != "2026-06-28 hi" {
		t.Errorf("commit call = %+v", commit)
	}
}

func TestCommitRejectsNoFiles(t *testing.T) {
	fake := &run.Fake{}
	if err := Commit(context.Background(), fake, "/repo", "m"); err == nil {
		t.Error("expected error when no files given")
	}
}

func TestCommitStopsOnAddError(t *testing.T) {
	fake := &run.Fake{CaptureErr: errors.New("boom")}
	if err := Commit(context.Background(), fake, "/repo", "m", "/repo/x.md"); err == nil {
		t.Fatal("expected error")
	}
	if len(fake.Calls) != 1 {
		t.Errorf("should stop after failed add, calls = %v", fake.Calls)
	}
}
