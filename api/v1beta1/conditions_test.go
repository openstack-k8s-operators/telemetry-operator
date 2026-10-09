package v1beta1

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestPersesResourceAvailable checks that PersesResourceAvailable reads the
// "Available" condition correctly: true when Available is True, false with the
// condition's reason/message otherwise (and a default when it is unset).
func TestPersesResourceAvailable(t *testing.T) {
	tests := []struct {
		name        string
		conditions  []metav1.Condition
		wantAvail   bool
		wantReason  string
		wantMessage string
	}{
		{
			name:        "no conditions yet (freshly created / perses-operator not reconciled)",
			conditions:  nil,
			wantAvail:   false,
			wantReason:  "PersesReconciling",
			wantMessage: "Waiting for the Perses operator to reconcile the object; no Perses instance may be available",
		},
		{
			name: "available true",
			conditions: []metav1.Condition{
				{Type: "Available", Status: metav1.ConditionTrue, Reason: "Reconciled", Message: "ok"},
				{Type: "Degraded", Status: metav1.ConditionFalse},
			},
			wantAvail: true,
		},
		{
			name: "available false propagates the perses-operator reason and message",
			conditions: []metav1.Condition{
				{Type: "Available", Status: metav1.ConditionFalse, Reason: "PersesMissing", Message: "no Perses instances found matching the label selector"},
				{Type: "Degraded", Status: metav1.ConditionTrue, Reason: "PersesMissing"},
			},
			wantAvail:   false,
			wantReason:  "PersesMissing",
			wantMessage: "no Perses instances found matching the label selector",
		},
		{
			name: "degraded true but available true is not misread as unavailable",
			conditions: []metav1.Condition{
				{Type: "Degraded", Status: metav1.ConditionTrue, Reason: "SomethingTransient"},
				{Type: "Available", Status: metav1.ConditionTrue, Reason: "Reconciled"},
			},
			wantAvail: true,
		},
		{
			name: "available false with empty reason falls back to a non-empty reason",
			conditions: []metav1.Condition{
				{Type: "Available", Status: metav1.ConditionFalse, Reason: "", Message: "backend error"},
			},
			wantAvail:   false,
			wantReason:  "PersesBackendNotReady",
			wantMessage: "backend error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			avail, reason, message := PersesResourceAvailable(tt.conditions)
			g.Expect(avail).To(Equal(tt.wantAvail))
			if !tt.wantAvail {
				g.Expect(reason).To(Equal(tt.wantReason))
				g.Expect(message).To(Equal(tt.wantMessage))
			}
		})
	}
}
