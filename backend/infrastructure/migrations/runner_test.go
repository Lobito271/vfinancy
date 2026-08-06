package migrations

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFilename(t *testing.T) {
	cases := []struct {
		name      string
		version   int
		mname     string
		direction string
		err       bool
	}{
		{"0001_create_users.up.sql", 1, "create_users", "up", false},
		{"0042_add_index.down.sql", 42, "add_index", "down", false},
		{"1_simple_name.up.sql", 1, "simple_name", "up", false},
		{"bad.sql", 0, "", "", true},
		{"abc.up.sql", 0, "", "", true},
		{"0001_test.weird.sql", 0, "", "", true},
	}
	for _, c := range cases {
		v, n, d, err := parseFilename(c.name)
		if (err != nil) != c.err {
			t.Errorf("%s: error=%v want err=%v", c.name, err, c.err)
		}
		if err != nil {
			continue
		}
		if v != c.version || n != c.mname || d != c.direction {
			t.Errorf("%s: got (%d,%s,%s) want (%d,%s,%s)", c.name, v, n, d, c.version, c.mname, c.direction)
		}
	}
}

func TestLoad_MergesUpAndDown(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"0001_init.up.sql":   "CREATE TABLE x (id INT);",
		"0001_init.down.sql": "DROP TABLE x;",
		"0002_seed.up.sql":   "INSERT INTO x VALUES (1);",
	}
	for name, body := range files {
		if err := writeFile(dir+"/"+name, body); err != nil {
			t.Fatal(err)
		}
	}

	r := &Runner{dir: dir}
	migs, err := r.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(migs) != 2 {
		t.Fatalf("expected 2 migrations, got %d", len(migs))
	}
	if migs[0].Version != 1 || migs[0].UpSQL == "" || migs[0].DownSQL == "" {
		t.Errorf("version 1 missing up/down: %+v", migs[0])
	}
	if migs[1].Version != 2 || migs[1].UpSQL == "" || migs[1].DownSQL != "" {
		t.Errorf("version 2 should have only up: %+v", migs[1])
	}
}

func TestLoad_InvalidFilesAreSkipped(t *testing.T) {
	dir := t.TempDir()
	_ = writeFile(dir+"/README.md", "ignored")
	_ = writeFile(dir+"/0001_good.up.sql", "SELECT 1;")

	r := &Runner{dir: dir}
	migs, err := r.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(migs) != 1 {
		t.Errorf("expected 1 migration, got %d", len(migs))
	}
}

func writeFile(path, body string) error {
	return os.WriteFile(filepath.Clean(path), []byte(body), 0o644)
}
