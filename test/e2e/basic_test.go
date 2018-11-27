package e2e

import (
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// TestGetStatsSummary creates a pod having two containers and queries the /stats/summary endpoint of the virtual-kubelet.
// It expects this endpoint to return stats for the current node, as well as for the aforementioned pod and each of its two containers.
func TestGetStatsSummary(t *testing.T) {
	// Create a pod with prefix "nginx-0-" having two containers.
	pod, err := f.CreatePod(f.CreateDummyPodObjectWithPrefix("nginx-0-", "bar", "baz"))
	if err != nil {
		t.Fatal(err)
	}
	// Delete the "nginx-0-X" pod after the test finishes.
	defer func() {
		if err := f.DeletePod(pod.Namespace, pod.Name); err != nil && !apierrors.IsNotFound(err) {
			t.Error(err)
		}
	}()

	// Wait for the "nginx-0-X" pod to be reported as running and ready.
	if err := f.WaitUntilPodReady(pod.Namespace, pod.Name); err != nil {
		t.Fatal(err)
	}

	// Grab the stats from the provider.
	stats, err := f.GetStatsSummary()
	if err != nil {
		t.Fatal(err)
	}

	// Make sure that we've got stats for the current node.
	if stats.Node.NodeName != f.NodeName {
		t.Fatalf("expected stats for node %s, got stats for node %s", f.NodeName, stats.Node.NodeName)
	}

	// Make sure that we've got stats for a single pod.
	desiredPodStatsCount := 1
	currentPodStatsCount := len(stats.Pods)
	if currentPodStatsCount != desiredPodStatsCount {
		t.Fatalf("expected stats for %d pods, got stats for %d pods", desiredPodStatsCount, currentPodStatsCount)
	}

	// Make sure that the pod for which stats were returned is the pod which we've created above.
	sp := stats.Pods[0]
	if sp.PodRef.Namespace != pod.Namespace || sp.PodRef.Name != pod.Name || string(sp.PodRef.UID) != string(pod.UID) {
		t.Fatalf("expected (%s, %s, %s), got (%s, %s, %s)", pod.Namespace, pod.Name, pod.UID, sp.PodRef.Namespace, sp.PodRef.Name, sp.PodRef.UID)
	}

	// Make sure that we've got stats for two containers.
	desiredContainerStatsCount := 2
	currentContainerStatsCount := len(sp.Containers)
	if currentPodStatsCount != desiredPodStatsCount {
		t.Fatalf("expected stats for %d containers, got stats for %d containers", desiredContainerStatsCount, currentContainerStatsCount)
	}
}

// TestPodLifecycle creates two pods and verifies that the provider has been asked to create them.
// Then, it deletes one of the pods and verifies that the provider has been asked to delete it.
// These verifications are made using the /stats/summary endpoint of the virtual-kubelet, by checking for the presence or absence of the pods.
// Hence, the provider being tested must implement the PodMetricsProvider interface.
func TestPodLifecycle(t *testing.T) {
	var (
		currentPodCount int
		desiredPodCount int
	)

	// Create a pod with prefix "nginx-0-" having a single container.
	pod0, err := f.CreatePod(f.CreateDummyPodObjectWithPrefix("nginx-0-", "foo"))
	if err != nil {
		t.Fatal(err)
	}
	// Delete the "nginx-0-X" pod after the test finishes.
	defer func() {
		if err := f.DeletePod(pod0.Namespace, pod0.Name); err != nil && !apierrors.IsNotFound(err) {
			t.Error(err)
		}
	}()

	// Create a pod with prefix "nginx-1-" having a single container.
	pod1, err := f.CreatePod(f.CreateDummyPodObjectWithPrefix("nginx-1-", "bar"))
	if err != nil {
		t.Fatal(err)
	}
	// Delete the "nginx-1-Y" pod after the test finishes.
	defer func() {
		if err := f.DeletePod(pod0.Namespace, pod0.Name); err != nil && !apierrors.IsNotFound(err) {
			t.Error(err)
		}
	}()

	// Wait for the "nginx-0-X" pod to be reported as running and ready.
	if err := f.WaitUntilPodReady(pod0.Namespace, pod0.Name); err != nil {
		t.Fatal(err)
	}
	// Wait for the "nginx-1-Y" pod to be reported as running and ready.
	if err := f.WaitUntilPodReady(pod1.Namespace, pod1.Name); err != nil {
		t.Fatal(err)
	}

	// Grab the stats from the provider.
	stats, err := f.GetStatsSummary()
	if err != nil {
		t.Fatal(err)
	}

	// Count the number of pods for which stats were returned.
	desiredPodCount = 2
	currentPodCount = len(stats.Pods)
	if currentPodCount != desiredPodCount {
		t.Fatalf("expected %d pods, provider knows about %d", desiredPodCount, currentPodCount)
	}

	// Delete the "nginx-1" pod.
	if err := f.DeletePod(pod1.Namespace, pod1.Name); err != nil {
		t.Fatal(err)
	}

	// Wait for the "nginx-1-Y" pod to be reported as having been marked for deletion.
	if err := f.WaitUntilPodDeleted(pod1.Namespace, pod1.Name); err != nil {
		t.Fatal(err)
	}

	// Grab the stats from the provider.
	stats, err = f.GetStatsSummary()
	if err != nil {
		t.Fatal(err)
	}

	// Count the number of pods for which stats were returned.
	desiredPodCount = 1
	currentPodCount = len(stats.Pods)
	if currentPodCount != desiredPodCount {
		t.Fatalf("expected %d pods, provider knows about %d", desiredPodCount, currentPodCount)
	}
}
