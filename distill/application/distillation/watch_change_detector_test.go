package distillation

import "testing"

func TestWatchChangeDetectorHasRelevantChangeWhenIdenticalCycles(t *testing.T) {
	detector := NewWatchChangeDetector()

	changed := detector.HasRelevantChange("tests: 12\\nfailures: 0", "tests: 12\\nfailures: 0")
	if changed {
		t.Fatalf("expected no relevant change for identical cycles")
	}
}

func TestWatchChangeDetectorIgnoresTimingOnlyChanges(t *testing.T) {
	detector := NewWatchChangeDetector()

	previous := "build completed in 120ms\\nchecks: 10"
	current := "build completed in 240ms\\nchecks: 10"

	changed := detector.HasRelevantChange(previous, current)
	if changed {
		t.Fatalf("expected no relevant change when only timing changed")
	}
}

func TestWatchChangeDetectorDetectsFailuresCountChange(t *testing.T) {
	detector := NewWatchChangeDetector()

	previous := "watch run\\nfailures: 0"
	current := "watch run\\nfailures: 1"

	changed := detector.HasRelevantChange(previous, current)
	if !changed {
		t.Fatalf("expected relevant change when failure count changed")
	}
}

func TestWatchChangeDetectorDetectsNewErrorLine(t *testing.T) {
	detector := NewWatchChangeDetector()

	previous := "watch run\\nall good"
	current := "watch run\\nERROR: database timeout"

	changed := detector.HasRelevantChange(previous, current)
	if !changed {
		t.Fatalf("expected relevant change when new error line appears")
	}
}
