package monitor

import (
	"os"
	"sync"
	"testing"
	"time"

	tp "github.com/accuknox/KubeArmor/KubeArmor/types"
)

func TestContainerMonitor(t *testing.T) {
	// Set up Test Data

	// containers
	Containers := map[string]tp.Container{}
	ContainersLock := &sync.Mutex{}

	// ActivePidMap
	ActivePidMap := map[string]tp.PidMap{}
	ActivePidMapLock := &sync.Mutex{}

	// Create Container Monitor

	containerMonitor := NewContainerMonitor("none", "file:/tmp/kubearmor-system.log", &Containers, &ContainersLock, &ActivePidMap, &ActivePidMapLock)
	if containerMonitor == nil {
		t.Log("[FAIL] Failed to create ContainerMonitor")
		return
	}

	t.Log("[PASS] Created ContainerMonitor")

	// Destroy Container Monitor

	if err := containerMonitor.DestroyContainerMonitor(); err != nil {
		t.Log("[FAIL] Failed to destroy ContainerMonitor")
	}

	t.Log("[PASS] Destroyed ContainerMonitor")

	// Remove system log

	if err := os.Remove("/tmp/kubearmor-system.log"); err != nil {
		t.Errorf("[FAIL] Failed to remove /tmp/kubearmor-system.log (%s)", err.Error())
		return
	}

	t.Log("[PASS] Removed /tmp/kubearmor-system.log")
}

func TestTraceSyscall(t *testing.T) {
	// Set up Test Data

	// containers
	Containers := map[string]tp.Container{}
	ContainersLock := &sync.Mutex{}

	// ActivePidMap
	ActivePidMap := map[string]tp.PidMap{}
	ActivePidMapLock := &sync.Mutex{}

	// Create Container Monitor

	containerMonitor := NewContainerMonitor("none", "file:/tmp/kubearmor-system.log", &Containers, &ContainersLock, &ActivePidMap, &ActivePidMapLock)
	if containerMonitor == nil {
		t.Log("[FAIL] Failed to create ContainerMonitor")
		return
	}

	t.Log("[PASS] Created ContainerMonitor")

	// Get the current directory

	dir := os.Getenv("PWD")

	t.Logf("[PASS] Got the current directory (%s)", dir)

	// Initialize BPF

	if err := containerMonitor.InitBPF(dir + "/.."); err != nil {
		t.Errorf("[FAIL] Failed to initialize BPF (%s)", err.Error())
		return
	}

	t.Logf("[PASS] Initialized BPF (Dir: %s/..)", dir)

	// wait for a while

	time.Sleep(time.Second * 1)

	// Start to trace syscalls

	go containerMonitor.TraceSyscall()

	t.Log("[PASS] Started to trace syscalls")

	// wait for a while

	time.Sleep(time.Second * 1)

	// Destroy Container Monitor

	if err := containerMonitor.DestroyContainerMonitor(); err != nil {
		t.Log("[FAIL] Failed to destroy ContainerMonitor")
	}

	t.Log("[PASS] Destroyed ContainerMonitor")

	// Remove system log

	if err := os.Remove("/tmp/kubearmor-system.log"); err != nil {
		t.Errorf("[FAIL] Failed to remove /tmp/kubearmor-system.log (%s)", err.Error())
		return
	}

	t.Log("[PASS] Removed /tmp/kubearmor-system.log")
}
