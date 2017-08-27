package state

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestOpen(t *testing.T) {
	t.Parallel()

	t.Run("happy_path", func(t *testing.T) {
		t.Parallel()

		dir, err := ioutil.TempDir("", "rssdl_state_test_")
		if err != nil {
			t.Fatalf("Couldn't create temporary directory: %v", err)
		}
		defer os.RemoveAll(dir)
		fn := filepath.Join(dir, "state")

		s, err := Open(fn)
		if err != nil {
			t.Fatalf("Couldn't open state: %v", err)
		}

		if v := s.GetOrder("key1"); v != "" {
			t.Errorf("s.GetOrder(%q) = %q, want %q", "key1", v, "")
		}
		if v := s.GetOrder("key2"); v != "" {
			t.Errorf("s.GetOrder(%q) = %q, want %q", "key2", v, "")
		}

		if err := s.SetOrder("key1", "val1"); err != nil {
			t.Errorf("s.SetOrder(%q, %q) got unexpected error: %v", "key1", "val1", err)
		}
		if err := s.SetOrder("key2", "val2"); err != nil {
			t.Errorf("s.SetOrder(%q, %q) got unexpected error: %v", "key2", "val2", err)
		}

		if v := s.GetOrder("key1"); v != "val1" {
			t.Errorf("s.GetOrder(%q) = %q, want %q", "key1", v, "val1")
		}
		if v := s.GetOrder("key2"); v != "val2" {
			t.Errorf("s.GetOrder(%q) = %q, want %q", "key2", v, "val2")
		}

		if err := s.SetOrder("key1", "val3"); err != nil {
			t.Errorf("s.SetOrder(%q, %q) got unexpected error: %v", "key1", "val3", err)
		}

		if v := s.GetOrder("key1"); v != "val3" {
			t.Errorf("s.GetOrder(%q) = %q, want %q", "key1", v, "val3")
		}
		if v := s.GetOrder("key2"); v != "val2" {
			t.Errorf("s.GetOrder(%q) = %q, want %q", "key2", v, "val2")
		}

		s, err = Open(fn)
		if err != nil {
			t.Fatalf("Couldn't open state: %v", err)
		}
		if v := s.GetOrder("key1"); v != "val3" {
			t.Errorf("s.GetOrder(%q) = %q, want %q", "key1", v, "val3")
		}
		if v := s.GetOrder("key2"); v != "val2" {
			t.Errorf("s.GetOrder(%q) = %q, want %q", "key2", v, "val2")
		}
	})

	t.Run("changes_kept_when_write_fails", func(t *testing.T) {
		t.Parallel()

		dir, err := ioutil.TempDir("", "rssdl_state_test_")
		if err != nil {
			t.Fatalf("Couldn't create temporary directory: %v", err)
		}
		defer os.RemoveAll(dir)
		fn := filepath.Join(dir, "state")

		s, err := Open(fn)
		if err != nil {
			t.Fatalf("Couldn't open state: %v", err)
		}
		if err := os.Chmod(dir, 0500); err != nil {
			t.Fatalf("Couldn't modify directory permissions: %v", err)
		}
		if err := s.SetOrder("key1", "val1"); err == nil {
			t.Fatal("s.SetOrder(%q, %q) expected error", "key1", "val1")
		}
		if err := os.Chmod(dir, 0700); err != nil {
			t.Fatalf("Couldn't modify directory permissions: %v", err)
		}
		if err := s.SetOrder("key2", "val2"); err != nil {
			t.Errorf("s.SetOrder(%q, %q) got unexpected error: %v", "key2", "val2", err)
		}

		s, err = Open(fn)
		if err != nil {
			t.Fatalf("Couldn't open state: %v", err)
		}
		if v := s.GetOrder("key1"); v != "val1" {
			t.Errorf("s.GetOrder(%q) = %q, want %q", "key1", v, "val1")
		}
		if v := s.GetOrder("key2"); v != "val2" {
			t.Errorf("s.GetOrder(%q) = %q, want %q", "key2", v, "val2")
		}
	})

	t.Run("unparseable", func(t *testing.T) {
		t.Parallel()

		dir, err := ioutil.TempDir("", "rssdl_state_test_")
		if err != nil {
			t.Fatalf("Couldn't create temporary directory: %v", err)
		}
		defer os.RemoveAll(dir)
		fn := filepath.Join(dir, "state")

		if err := ioutil.WriteFile(fn, []byte("garbage"), 0600); err != nil {
			t.Fatalf("Couldn't create state file: %v", err)
		}

		_, err = Open(fn)
		re := regexp.MustCompile(`could not parse state`)
		if err == nil || !re.MatchString(err.Error()) {
			t.Errorf("Open got error %q, wanted error matching pattern %q", err, re)
		}
	})

	t.Run("unreadable", func(t *testing.T) {
		t.Parallel()

		dir, err := ioutil.TempDir("", "rssdl_state_test_")
		if err != nil {
			t.Fatalf("Couldn't create temporary directory: %v", err)
		}
		defer os.RemoveAll(dir)
		fn := filepath.Join(dir, "state")

		if err := ioutil.WriteFile(fn, nil, 0); err != nil {
			t.Fatalf("Couldn't create state file: %v", err)
		}

		_, err = Open(fn)
		re := regexp.MustCompile(`could not read state file`)
		if err == nil || !re.MatchString(err.Error()) {
			t.Errorf("Open got error %q, wanted error matching pattern %q", err, re)
		}
	})

	t.Run("unwritable", func(t *testing.T) {
		t.Parallel()

		dir, err := ioutil.TempDir("", "rssdl_state_test_")
		if err != nil {
			t.Fatalf("Couldn't create temporary directory: %v", err)
		}
		defer os.RemoveAll(dir)
		fn := filepath.Join(dir, "state")

		if err := ioutil.WriteFile(fn, nil, 0640); err != nil {
			t.Fatalf("Couldn't create state file: %v", err)
		}
		if err := os.Chmod(dir, 0500); err != nil {
			t.Fatalf("Couldn't modify directory permissions: %v", err)
		}

		_, err = Open(fn)
		re := regexp.MustCompile(`could not write state file`)
		if err == nil || !re.MatchString(err.Error()) {
			t.Errorf("Open got error %q, wanted error matching pattern %q", err, re)
		}
	})
}
