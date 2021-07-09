// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8sobjectsreceiver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/obsreport"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver/testutils"
)

func TestReceiver(t *testing.T) {
	client := fake.NewSimpleClientset()
	sink := new(consumertest.MetricsSink)

	r := setupReceiver(client, sink, 10*time.Second)

	// Setup k8s resources.
	numPods := 2
	numNodes := 1
	createPods(t, client, numPods)
	createNodes(t, client, numNodes)

	ctx := context.Background()
	require.NoError(t, r.Start(ctx, componenttest.NewNopHost()))

	// Expects metric data from nodes and pods where each metric data
	// struct corresponds to one resource.
	expectedNumMetrics := numPods + numNodes
	var initialMetricsCount int
	require.Eventually(t, func() bool {
		initialMetricsCount = sink.MetricsCount()
		return initialMetricsCount == expectedNumMetrics
	}, 10*time.Second, 100*time.Millisecond,
		"metrics not collected")

	numPodsToDelete := 1
	deletePods(t, client, numPodsToDelete)

	// Expects metric data from a node, since other resources were deleted.
	expectedNumMetrics = (numPods - numPodsToDelete) + numNodes
	var metricsCountDelta int
	require.Eventually(t, func() bool {
		metricsCountDelta = sink.MetricsCount() - initialMetricsCount
		return metricsCountDelta == expectedNumMetrics
	}, 10*time.Second, 100*time.Millisecond,
		"updated metrics not collected")

	r.Shutdown(ctx)
}

func TestReceiverTimesOutAfterStartup(t *testing.T) {
	client := fake.NewSimpleClientset()

	// Mock initial cache sync timing out, using a small timeout.
	r := setupReceiver(client, consumertest.NewNop(), 1*time.Millisecond)

	createPods(t, client, 1)

	ctx := context.Background()
	require.NoError(t, r.Start(ctx, componenttest.NewNopHost()))
	require.Eventually(t, func() bool {
		return r.resourceWatcher.initialSyncTimedOut.Load()
	}, 10*time.Second, 100*time.Millisecond)
	require.NoError(t, r.Shutdown(ctx))
}

func TestReceiverWithManyResources(t *testing.T) {
	client := fake.NewSimpleClientset()
	sink := new(consumertest.MetricsSink)

	r := setupReceiver(client, sink, 10*time.Second)

	numPods := 1000
	createPods(t, client, numPods)

	ctx := context.Background()
	require.NoError(t, r.Start(ctx, componenttest.NewNopHost()))

	require.Eventually(t, func() bool {
		return sink.MetricsCount() == numPods
	}, 10*time.Second, 100*time.Millisecond,
		"metrics not collected")

	r.Shutdown(ctx)
}

var numCalls *atomic.Int32
var consumeMetadataInvocation = func() {
	if numCalls != nil {
		numCalls.Inc()
	}
}

func TestReceiverWithMetadata(t *testing.T) {
	client := fake.NewSimpleClientset()
	next := &mockExporterWithK8sMetadata{MetricsSink: new(consumertest.MetricsSink)}
	numCalls = atomic.NewInt32(0)

	r := setupReceiver(client, next, 10*time.Second)
	r.config.MetadataExporters = []string{"nop/withmetadata"}

	// Setup k8s resources.
	pods := createPods(t, client, 1)

	ctx := context.Background()
	require.NoError(t, r.Start(ctx, nopHostWithExporters{}))

	// Mock an update on the Pod object. It appears that the fake clientset
	// does not pass on events for updates to resources.
	require.Len(t, pods, 1)
	updatedPod := getUpdatedPod(pods[0])
	r.resourceWatcher.onUpdate(pods[0], updatedPod)

	// Should not result in ConsumerKubernetesMetadata invocation.
	r.resourceWatcher.onUpdate(pods[0], pods[0])

	deletePods(t, client, 1)

	// Ensure ConsumeKubernetesMetadata is called twice, once for the add and
	// then for the update.
	require.Eventually(t, func() bool {
		return int(numCalls.Load()) == 2
	}, 10*time.Second, 100*time.Millisecond,
		"metadata not collected")

	r.Shutdown(ctx)
}

func getUpdatedPod(pod *corev1.Pod) interface{} {
	return &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			UID:       pod.UID,
			Labels: map[string]string{
				"key": "value",
			},
		},
	}
}

func setupReceiver(
	client *fake.Clientset,
	consumer consumer.Metrics,
	initialSyncTimeout time.Duration) *kubernetesReceiver {

	logger := zap.NewNop()
	config := &Config{
		CollectionInterval:         1 * time.Second,
		NodeConditionTypesToReport: []string{"Ready"},
	}

	rw := newResourceWatcher(logger, client, config.NodeConditionTypesToReport, initialSyncTimeout)
	rw.dataCollector.SetupMetadataStore(&corev1.Service{}, &testutils.MockStore{})

	return &kubernetesReceiver{
		resourceWatcher: rw,
		logger:          logger,
		config:          config,
		consumer:        consumer,
		obsrecv:         obsreport.NewReceiver(obsreport.ReceiverSettings{ReceiverID: config.ID(), Transport: "http"}),
	}
}
