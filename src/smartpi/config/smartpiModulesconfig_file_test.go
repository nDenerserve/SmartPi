package config

import (
	"reflect"
	"testing"
)

func TestSplitList(t *testing.T) {
	def := []string{"root", "smartpi"}
	cases := []struct {
		raw  string
		want []string
	}{
		{"", def},
		{"   ", def},
		{",", def},
		{"alice", []string{"alice"}},
		{"root, smartpi ,bob", []string{"root", "smartpi", "bob"}},
		{"a,,b,", []string{"a", "b"}},
	}
	for _, c := range cases {
		if got := splitList(c.raw, def); !reflect.DeepEqual(got, c.want) {
			t.Errorf("splitList(%q) = %q, want %q", c.raw, got, c.want)
		}
	}
}

func TestSplitList_DefaultsAreCopied(t *testing.T) {
	def := []string{"root"}
	got := splitList("", def)
	got[0] = "changed"
	if def[0] != "root" {
		t.Fatal("splitList returned the defaults slice itself, callers could modify it")
	}
}
