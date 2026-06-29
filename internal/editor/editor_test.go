package editor

import (
	"context"
	"reflect"
	"testing"

	"github.com/manningc33/journal/internal/config"
	"github.com/manningc33/journal/internal/run"
)

func TestOpenDefaultArgs(t *testing.T) {
	fake := &run.Fake{}
	cfg := config.EditorConfig{Cmd: "nvim", Args: []string{"+{line}"}, Line: 3}
	if err := Open(context.Background(), fake, cfg, "/x/entry.md"); err != nil {
		t.Fatal(err)
	}
	if len(fake.Calls) != 1 {
		t.Fatalf("calls = %d", len(fake.Calls))
	}
	c := fake.Calls[0]
	if c.Kind != "interactive" || c.Name != "nvim" {
		t.Fatalf("call = %+v", c)
	}
	if want := []string{"+3", "/x/entry.md"}; !reflect.DeepEqual(c.Args, want) {
		t.Errorf("args = %v, want %v", c.Args, want)
	}
}

func TestOpenExplicitFilePlaceholder(t *testing.T) {
	fake := &run.Fake{}
	cfg := config.EditorConfig{Cmd: "code", Args: []string{"-g", "{file}:{line}"}, Line: 3}
	if err := Open(context.Background(), fake, cfg, "/x/entry.md"); err != nil {
		t.Fatal(err)
	}
	if want := []string{"-g", "/x/entry.md:3"}; !reflect.DeepEqual(fake.Calls[0].Args, want) {
		t.Errorf("args = %v, want %v", fake.Calls[0].Args, want)
	}
}
