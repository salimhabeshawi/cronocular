package timer

import (
	"testing"
	"time"
)

func TestControllerTransitionsThroughFocusAndRest(t *testing.T) {
	restStarted := make(chan struct{}, 1)
	restEnded := make(chan struct{}, 1)
	controller := NewController(Config{
		Focus: 20 * time.Millisecond,
		Rest:  20 * time.Millisecond,
	}, Actions{
		RestStarted: func() { restStarted <- struct{}{} },
		RestEnded:   func() { restEnded <- struct{}{} },
	})

	now := time.Now()
	controller.advance(now.Add(25 * time.Millisecond))

	select {
	case <-restStarted:
	case <-time.After(150 * time.Millisecond):
		t.Fatal("rest reminder did not fire")
	}

	if got := controller.Snapshot().Phase; got != PhaseRest {
		t.Fatalf("phase = %s, want %s", got, PhaseRest)
	}

	controller.advance(now.Add(50 * time.Millisecond))

	select {
	case <-restEnded:
	case <-time.After(150 * time.Millisecond):
		t.Fatal("rest completion did not fire")
	}

	if got := controller.Snapshot().Phase; got != PhaseFocus {
		t.Fatalf("phase = %s, want %s", got, PhaseFocus)
	}
}

func TestPauseResumePreservesRemainingTime(t *testing.T) {
	controller := NewController(Config{
		Focus: time.Minute,
		Rest:  time.Second,
	}, Actions{})

	controller.Pause()
	before := controller.Snapshot().Remaining
	time.Sleep(10 * time.Millisecond)
	after := controller.Snapshot().Remaining

	if before != after {
		t.Fatalf("paused remaining changed from %s to %s", before, after)
	}

	controller.Resume()
	if controller.Snapshot().Paused {
		t.Fatal("controller is still paused after resume")
	}
}
