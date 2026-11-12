package rag

import "testing"

func TestNormalizeDashes(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"spaced em dash", "He led the migration — enabling SSO.", "He led the migration, enabling SSO."},
		{"unspaced em dash", "Go APIs—and Temporal workflows.", "Go APIs, and Temporal workflows."},
		{"spaced en dash", "Remote – fully distributed.", "Remote, fully distributed."},
		{"numeric range", "Notabene, 2021–2025, and 2017 — 2020.", "Notabene, 2021-2025, and 2017-2020."},
		{"leading bullet", "Highlights:\n— Go\n  – Kafka", "Highlights:\n- Go\n  - Kafka"},
		{"after comma", "Kafka, — and Temporal.", "Kafka, and Temporal."},
		{"before period", "He built it —.", "He built it."},
		{"trailing", "Shipped it —\nNext line.", "Shipped it\nNext line."},
		{"no dashes", "Plain text - with a hyphen.", "Plain text - with a hyphen."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeDashes(tc.in); got != tc.want {
				t.Fatalf("normalizeDashes(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
