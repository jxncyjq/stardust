package nats

import (
	"reflect"
	"testing"
)

func TestMergeStreamSubjectsAddsMissingSubjects(t *testing.T) {
	current := []string{"orders.created"}
	required := []string{"orders.updated", "payments.created"}

	got, changed := mergeStreamSubjects(current, required)
	want := []string{"orders.created", "orders.updated", "payments.created"}

	if !changed {
		t.Errorf("mergeStreamSubjects(%v, %v) changed = false, want true", current, required)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mergeStreamSubjects(%v, %v) = %v, want %v", current, required, got, want)
	}
}

func TestMergeStreamSubjectsSkipsSubjectsCoveredByWildcard(t *testing.T) {
	current := []string{"orders.>"}
	required := []string{"orders.created", "orders.updated"}

	got, changed := mergeStreamSubjects(current, required)
	want := []string{"orders.>"}

	if changed {
		t.Errorf("mergeStreamSubjects(%v, %v) changed = true, want false", current, required)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mergeStreamSubjects(%v, %v) = %v, want %v", current, required, got, want)
	}
}

func TestSubjectPatternCovers(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		subject string
		want    bool
	}{
		{
			name:    "exact match",
			pattern: "orders.created",
			subject: "orders.created",
			want:    true,
		},
		{
			name:    "single token wildcard",
			pattern: "orders.*",
			subject: "orders.created",
			want:    true,
		},
		{
			name:    "single token wildcard does not cross token boundary",
			pattern: "orders.*",
			subject: "orders.created.us",
			want:    false,
		},
		{
			name:    "tail wildcard covers remaining tokens",
			pattern: "orders.>",
			subject: "orders.created.us",
			want:    true,
		},
		{
			name:    "tail wildcard requires at least one remaining token",
			pattern: "orders.>",
			subject: "orders",
			want:    false,
		},
		{
			name:    "tail wildcard requires prefix match",
			pattern: "orders.>",
			subject: "payments.created",
			want:    false,
		},
		{
			name:    "global tail wildcard",
			pattern: ">",
			subject: "payments.created",
			want:    true,
		},
		{
			name:    "tail wildcard only at end",
			pattern: "orders.>.created",
			subject: "orders.us.created",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := subjectPatternCovers(tt.pattern, tt.subject)
			if got != tt.want {
				t.Errorf("subjectPatternCovers(%q, %q) = %t, want %t", tt.pattern, tt.subject, got, tt.want)
			}
		})
	}
}
