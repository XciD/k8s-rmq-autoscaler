package main

import (
	"testing"
)

var (
	app = &App{
		key:          "key",
		minWorkers:   1,
		maxWorkers:   10,
		readyWorkers: 0,
		replicas:     1,
		steps:        1,
		offset:       0,
	}
)

func TestUnstable(t *testing.T) {
	incReplicas := app.scale(0, 1)

	if incReplicas != 0 {
		t.Error("Expected 0, got ", incReplicas)
	}

	app.readyWorkers = 1
	incReplicas = app.scale(0, 1)

	if incReplicas != 0 {
		t.Error("Expected 0, got ", incReplicas)
	}
}

func TestScaleUp(t *testing.T) {
	app.readyWorkers = 1
	app.replicas = 1

	incReplicas := app.scale(1, 2)

	// Should increase the workers
	if incReplicas != 1 {
		t.Error("Expected 1, got ", incReplicas)
	}

	app.readyWorkers = 2
	app.replicas = 2
	incReplicas = app.scale(2, 4)

	// Should increase one worker
	if incReplicas != 1 {
		t.Error("Expected 1, got ", incReplicas)
	}

	app.readyWorkers = 10
	app.replicas = 10
	incReplicas = app.scale(10, 11)

	// Should do nothing, too much worker
	if incReplicas != 0 {
		t.Error("Expected 0, got ", incReplicas)
	}
}

func TestScaleDown(t *testing.T) {
	app.readyWorkers = 2
	app.replicas = 2

	incReplicas := app.scale(2, 1)

	// Should decrease the workers
	if incReplicas != -1 {
		t.Error("Expected -1, got ", incReplicas)
	}

	app.readyWorkers = 1
	app.replicas = 1
	incReplicas = app.scale(1, 0)

	// Should do nothing, need at least one worker
	if incReplicas != 0 {
		t.Error("Expected 0, got ", incReplicas)
	}
}

func TestMinMax(t *testing.T) {
	app.readyWorkers = 0
	app.replicas = 0
	incReplicas := app.scale(0, 0)

	// Should return to min workers
	if incReplicas != 1 {
		t.Error("Expected 1, got ", incReplicas)
	}

	app.readyWorkers = 11
	app.replicas = 11
	incReplicas = app.scale(11, 12)

	// Should return to max workers
	if incReplicas != -1 {
		t.Error("Expected -1, got ", incReplicas)
	}

	app.overrideLimits = true

	app.readyWorkers = 11
	app.replicas = 11
	incReplicas = app.scale(11, 12)

	// Should do nothing
	if incReplicas != 0 {
		t.Error("Expected à, got ", incReplicas)
	}

	app.readyWorkers = 0
	app.replicas = 0
	incReplicas = app.scale(0, 0)

	// Should do nothing
	if incReplicas != 0 {
		t.Error("Expected 0, got ", incReplicas)
	}
}

func TestOffset(t *testing.T) {
	app.readyWorkers = 2
	app.replicas = 2
	app.offset = 2

	incReplicas := app.scale(2, 3)

	// We want at least 2 worker more than message in queue
	if incReplicas != 1 {
		t.Error("Expected 1, got ", incReplicas)
	}
}

func TestStepUp(t *testing.T) {
	app.offset = 0
	app.readyWorkers = 2
	app.replicas = 2
	app.steps = 2

	incReplicas := app.scale(2, 4)

	// We want at least 3 worker more than message in queue, but step limit to 2
	if incReplicas != 2 {
		t.Error("Expected 2, got ", incReplicas)
	}

	app.steps = 3
	incReplicas = app.scale(2, 4)

	// We want at least 3 worker more than message in queue, but step limit to 2
	if incReplicas != 2 {
		t.Error("Expected 2, got ", incReplicas)
	}
}

func TestStepDown(t *testing.T) {
	app.offset = 0
	app.readyWorkers = 4
	app.replicas = 4
	app.steps = 2

	incReplicas := app.scale(4, 2)

	// We want at least 3 worker more than message in queue, but step limit to 2
	if incReplicas != -2 {
		t.Error("Expected -2, got ", incReplicas)
	}

	app.steps = 3
	incReplicas = app.scale(4, 2)

	// We want at least 3 worker more than message in queue, but step limit to 2
	if incReplicas != -2 {
		t.Error("Expected -2, got ", incReplicas)
	}
}
