package skills_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/budaev/stell/internal/skills"
)

func TestLoadSkillLazyBody(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "demo", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: demo\ndescription: a demo skill\ntriggers:\n  keywords: [\"hello\"]\n---\nSECRET_BODY\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	mgr, err := skills.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	s, err := mgr.Get("demo")
	if err != nil {
		t.Fatal(err)
	}
	if s.Body() != "" {
		t.Fatalf("body loaded too early: %q", s.Body())
	}
	if len(mgr.Index()) != 1 || mgr.Index()[0].Name != "demo" {
		t.Fatalf("index=%v", mgr.Index())
	}
	loaded, err := mgr.LoadSkill("demo")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Body() == "" || loaded.Body() != "SECRET_BODY\n" && loaded.Body() != "SECRET_BODY" {
		if loaded.Body() == "" {
			t.Fatal("expected body after LoadSkill")
		}
	}
}

func TestMatchTrigger(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, []byte("---\nname: x\ndescription: d\ntriggers:\n  keywords: [\"unique-kw\"]\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mgr, err := skills.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(mgr.Match("please unique-kw now", nil)); n != 1 {
		t.Fatalf("match=%d", n)
	}
	if n := len(mgr.Match("unrelated", nil)); n != 0 {
		t.Fatalf("unexpected match=%d", n)
	}
}
