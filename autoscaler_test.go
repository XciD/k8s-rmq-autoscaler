package main

import (
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	app = &App{
		key:               "key",
		minWorkers:        1,
		maxWorkers:        10,
		messagesPerWorker: 1,
		readyWorkers:      0,
		replicas:          1,
		steps:             1,
		offset:            0,
		createdDate:       time.Now(),
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

func TestMessagePerWorker(t *testing.T) {
	app.offset = 0
	app.readyWorkers = 4
	app.replicas = 4
	app.steps = 1
	app.messagesPerWorker = 2

	incReplicas := app.scale(4, 8)

	if incReplicas != 0 {
		t.Error("Expected 0, got ", incReplicas)
	}

	incReplicas = app.scale(4, 9)

	if incReplicas != 1 {
		t.Error("Expected 1, got ", incReplicas)
	}
}

func TestCoolDown(t *testing.T) {
	isCoolDown := app.isCoolDown()

	if isCoolDown != false {
		t.Error("Expected false, got ", isCoolDown)
	}

	app.coolDownDelay = time.Minute

	isCoolDown = app.isCoolDown()

	if isCoolDown != true {
		t.Error("Expected true, got ", isCoolDown)
	}

	app.createdDate = app.createdDate.Add(-time.Minute)

	isCoolDown = app.isCoolDown()

	// We want at least 3 worker more than message in queue, but step limit to 2
	if isCoolDown != false {
		t.Error("Expected false, got ", isCoolDown)
	}
}

func TestCreateApp(t *testing.T) {
	deployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				"k8s-rmq-autoscaler/enable": "true",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: 1,
		},
	}

	app, err := createApp(deployment, "test")

	if app != nil {
		t.Error("App should not be created")
	}
	if err.Error() != "deployment: test has no property `queue` not filled" {
		t.Error("Error message not right", err)
	}

	// Add the missing information
	deployment.ObjectMeta.Annotations["k8s-rmq-autoscaler/queue"] = "queue"

	app, err = createApp(deployment, "test")

	if app != nil {
		t.Error("App should not be created")
	}
	if err.Error() != "deployment: test has no property `vhost` not filled" {
		t.Error("Error message not right", err)
	}

	// Add the missing information
	deployment.ObjectMeta.Annotations["k8s-rmq-autoscaler/vhost"] = "vhost"

	app, err = createApp(deployment, "test")

	if app != nil {
		t.Error("App should not be created")
	}
	if err.Error() != "deployment: test has no property `min-workers` not filled" {
		t.Error("Error message not right", err)
	}

	// Add a non int value
	deployment.ObjectMeta.Annotations["k8s-rmq-autoscaler/min-workers"] = "nan"

	app, err = createApp(deployment, "test")

	if app != nil {
		t.Error("App should not be created")
	}
	if err.Error() != "deployment: test property `min-workers` is not an int (ex: 1)" {
		t.Error("Error message not right", err)
	}

	// Add a missing value
	deployment.ObjectMeta.Annotations["k8s-rmq-autoscaler/min-workers"] = "1"

	app, err = createApp(deployment, "test")

	if app != nil {
		t.Error("App should not be created")
	}
	if err.Error() != "deployment: test has no property `max-workers` not filled" {
		t.Error("Error message not right", err)
	}

	// Add a missing value
	deployment.ObjectMeta.Annotations["k8s-rmq-autoscaler/max-workers"] = "2"

	app, err = createApp(deployment, "test")

	if app == nil {
		t.Error("App should be created with default values")
	}

	if app.ref != deployment {
		t.Error("Missing reference to deployment")
	}
	if app.coolDownDelay != 0 {
		t.Error("coolDownDelay default value not 0")
	}
	if app.key != "test" {
		t.Error("key not right")
	}
	if app.offset != 0 {
		t.Error("offset default value not 0")
	}
	if app.steps != 1 {
		t.Error("steps default value not 1")
	}
	if app.replicas != 2 {
		t.Error("replicas not read correctly")
	}
	if app.readyWorkers != 1 {
		t.Error("readyWorkers not read correctly")
	}
	if app.overrideLimits != false {
		t.Error("overrideLimits default value not false")
	}
	if app.safeUnscale != true {
		t.Error("overrideLimits default value not true")
	}
	if app.queue != "queue" {
		t.Error("queue not set correctly")
	}
	if app.minWorkers != 1 {
		t.Error("minWorkers not set correctly")
	}
	if app.maxWorkers != 2 {
		t.Error("maxWorkers not set correctly")
	}

	// Add a optional annotations
	deployment.ObjectMeta.Annotations["k8s-rmq-autoscaler/messages-per-worker"] = "2"

	app, err = createApp(deployment, "test")

	if app == nil {
		t.Error("App should be created with default values")
	}

	if app.messagesPerWorker != 2 {
		t.Error("messagesPerWorker not set correctly")
	}

	// Add a optional annotations
	deployment.ObjectMeta.Annotations["k8s-rmq-autoscaler/steps"] = "2"

	app, err = createApp(deployment, "test")

	if app == nil {
		t.Error("App should be created with default values")
	}

	if app.steps != 2 {
		t.Error("steps not set correctly")
	}

	// Add a optional annotations
	deployment.ObjectMeta.Annotations["k8s-rmq-autoscaler/offset"] = "2"

	app, err = createApp(deployment, "test")

	if app == nil {
		t.Error("App should be created with default values")
	}

	if app.offset != 2 {
		t.Error("steps not set correctly")
	}

	// Add a optional annotations
	deployment.ObjectMeta.Annotations["k8s-rmq-autoscaler/override"] = "true"

	app, err = createApp(deployment, "test")

	if app == nil {
		t.Error("App should be created with default values")
	}

	if app.overrideLimits != true {
		t.Error("overrideLimits not set correctly")
	}

	// Add a optional annotations
	deployment.ObjectMeta.Annotations["k8s-rmq-autoscaler/safe-unscale"] = "false"

	app, err = createApp(deployment, "test")

	if app == nil {
		t.Error("App should be created with default values")
	}

	if app.safeUnscale != false {
		t.Error("safeUnscale not set correctly")
	}

	// Add a optional annotations
	deployment.ObjectMeta.Annotations["k8s-rmq-autoscaler/cooldown-delay"] = "5m0s"

	app, err = createApp(deployment, "test")

	if app == nil {
		t.Error("App should be created with default values")
	}

	if app.coolDownDelay != 5*time.Minute {
		t.Error("coolDownDelay not set correctly")
	}
}
