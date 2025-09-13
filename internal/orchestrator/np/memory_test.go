package np

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/svaan1/tcc-go/internal/ffmpeg"
	"github.com/svaan1/tcc-go/internal/metrics"
)

func TestNewInMemoryNodePool(t *testing.T) {
	pool := NewInMemoryNodePool()
	if pool == nil {
		t.Fatal("NewInMemoryNodePool() returned nil")
	}
	if pool.nodes == nil {
		t.Fatal("nodes map not initialized")
	}
}

func TestRegisterNode(t *testing.T) {
	pool := NewInMemoryNodePool()
	ctx := context.Background()

	t.Run("valid registration", func(t *testing.T) {
		req := &NodeRegistration{
			Name: "test-node",
			Profiles: []ffmpeg.EncodingProfile{
				{Codec: "h264"},
				{Codec: "h265"},
			},
		}

		node, err := pool.RegisterNode(ctx, req)
		if err != nil {
			t.Fatalf("RegisterNode() failed: %v", err)
		}

		if node.ID == uuid.Nil {
			t.Error("node ID should not be nil")
		}
		if node.Name != req.Name {
			t.Errorf("expected name %s, got %s", req.Name, node.Name)
		}
		if len(node.Profiles) != len(req.Profiles) {
			t.Errorf("expected %d codecs, got %d", len(req.Profiles), len(node.Profiles))
		}
	})

	t.Run("nil registration", func(t *testing.T) {
		_, err := pool.RegisterNode(ctx, nil)
		if err == nil {
			t.Error("expected error for nil registration")
		}
	})

	t.Run("empty name", func(t *testing.T) {
		req := &NodeRegistration{Name: ""}
		_, err := pool.RegisterNode(ctx, req)
		if err == nil {
			t.Error("expected error for empty name")
		}
	})
}

func TestUnregisterNode(t *testing.T) {
	pool := NewInMemoryNodePool()
	ctx := context.Background()

	// Register a node first
	req := &NodeRegistration{Name: "test-node"}
	node, _ := pool.RegisterNode(ctx, req)

	t.Run("valid unregistration", func(t *testing.T) {
		err := pool.UnregisterNode(ctx, node.ID)
		if err != nil {
			t.Fatalf("UnregisterNode() failed: %v", err)
		}

		// Verify node is removed
		_, err = pool.GetNode(ctx, node.ID)
		if err == nil {
			t.Error("expected error when getting removed node")
		}
	})

	t.Run("non-existent node", func(t *testing.T) {
		err := pool.UnregisterNode(ctx, uuid.New())
		if err == nil {
			t.Error("expected error for non-existent node")
		}
	})
}

func TestUpdateNodeMetrics(t *testing.T) {
	pool := NewInMemoryNodePool()
	ctx := context.Background()

	req := &NodeRegistration{Name: "test-node"}
	node, _ := pool.RegisterNode(ctx, req)

	t.Run("valid update", func(t *testing.T) {
		usage := &metrics.ResourceUsage{
			CPUUsagePercent:    50.0,
			MemoryUsagePercent: 30.0,
			DiskUsagePercent:   20.0,
		}

		err := pool.UpdateNodeMetrics(ctx, node.ID, usage)
		if err != nil {
			t.Fatalf("UpdateNodeMetrics() failed: %v", err)
		}

		// Verify update
		updatedNode, _ := pool.GetNode(ctx, node.ID)
		if updatedNode.ResourceUsage == nil {
			t.Error("resource usage should not be nil")
		}
		if updatedNode.ResourceUsage.CPUUsagePercent != 50.0 {
			t.Errorf("expected CPU usage 50.0, got %f", updatedNode.ResourceUsage.CPUUsagePercent)
		}
	})

	t.Run("non-existent node", func(t *testing.T) {
		usage := &metrics.ResourceUsage{}
		err := pool.UpdateNodeMetrics(ctx, uuid.New(), usage)
		if err == nil {
			t.Error("expected error for non-existent node")
		}
	})
}

func TestGetNode(t *testing.T) {
	pool := NewInMemoryNodePool()
	ctx := context.Background()

	req := &NodeRegistration{Name: "test-node"}
	registeredNode, _ := pool.RegisterNode(ctx, req)

	t.Run("existing node", func(t *testing.T) {
		node, err := pool.GetNode(ctx, registeredNode.ID)
		if err != nil {
			t.Fatalf("GetNode() failed: %v", err)
		}
		if node.ID != registeredNode.ID {
			t.Error("returned node ID doesn't match")
		}
	})

	t.Run("non-existent node", func(t *testing.T) {
		_, err := pool.GetNode(ctx, uuid.New())
		if err == nil {
			t.Error("expected error for non-existent node")
		}
	})
}

func TestListNodes(t *testing.T) {
	pool := NewInMemoryNodePool()
	ctx := context.Background()

	// Register multiple nodes
	for i := 0; i < 5; i++ {
		req := &NodeRegistration{Name: "test-node-" + string(rune('0'+i))}
		pool.RegisterNode(ctx, req)
	}

	t.Run("list all nodes", func(t *testing.T) {
		nodes, err := pool.ListNodes(ctx, 0, 10)
		if err != nil {
			t.Fatalf("ListNodes() failed: %v", err)
		}
		if len(nodes) != 5 {
			t.Errorf("expected 5 nodes, got %d", len(nodes))
		}
	})

	t.Run("pagination", func(t *testing.T) {
		nodes, err := pool.ListNodes(ctx, 2, 2)
		if err != nil {
			t.Fatalf("ListNodes() failed: %v", err)
		}
		if len(nodes) != 2 {
			t.Errorf("expected 2 nodes, got %d", len(nodes))
		}
	})

	t.Run("offset beyond range", func(t *testing.T) {
		nodes, err := pool.ListNodes(ctx, 10, 5)
		if err != nil {
			t.Fatalf("ListNodes() failed: %v", err)
		}
		if len(nodes) != 0 {
			t.Errorf("expected 0 nodes, got %d", len(nodes))
		}
	})
}

func TestGetAvailableNodes(t *testing.T) {
	pool := NewInMemoryNodePool()
	ctx := context.Background()

	// Register nodes with different codecs
	node1Req := &NodeRegistration{
		Name: "node1",
		Profiles: []ffmpeg.EncodingProfile{
			{Codec: "h264"},
			{Codec: "h265"},
		},
	}
	node2Req := &NodeRegistration{
		Name: "node2",
		Profiles: []ffmpeg.EncodingProfile{
			{Codec: "h264"},
			{Codec: "vp9"},
		},
	}
	node3Req := &NodeRegistration{
		Name: "node3",
		Profiles: []ffmpeg.EncodingProfile{
			{Codec: "h265"},
		},
	}

	pool.RegisterNode(ctx, node1Req)
	pool.RegisterNode(ctx, node2Req)
	pool.RegisterNode(ctx, node3Req)

	t.Run("no filter", func(t *testing.T) {
		nodes, err := pool.GetAvailableNodes(ctx, nil)
		if err != nil {
			t.Fatalf("GetAvailableNodes() failed: %v", err)
		}
		if len(nodes) != 3 {
			t.Errorf("expected 3 nodes, got %d", len(nodes))
		}
	})

	t.Run("empty requirements", func(t *testing.T) {
		filter := &NodeFilter{Codec: ""}
		nodes, err := pool.GetAvailableNodes(ctx, filter)
		if err != nil {
			t.Fatalf("GetAvailableNodes() failed: %v", err)
		}
		if len(nodes) != 3 {
			t.Errorf("expected 3 nodes, got %d", len(nodes))
		}
	})

	t.Run("filter by h264", func(t *testing.T) {
		filter := &NodeFilter{Codec: "h264"}
		nodes, err := pool.GetAvailableNodes(ctx, filter)
		if err != nil {
			t.Fatalf("GetAvailableNodes() failed: %v", err)
		}
		if len(nodes) != 2 {
			t.Errorf("expected 2 nodes with h264, got %d", len(nodes))
		}
	})

	t.Run("filter by h265", func(t *testing.T) {
		filter := &NodeFilter{Codec: "h265"}
		nodes, err := pool.GetAvailableNodes(ctx, filter)
		if err != nil {
			t.Fatalf("GetAvailableNodes() failed: %v", err)
		}
		if len(nodes) != 2 {
			t.Errorf("expected 2 nodes with h265, got %d", len(nodes))
		}
	})

	t.Run("filter by non-existent codec", func(t *testing.T) {
		filter := &NodeFilter{Codec: "av1"}
		nodes, err := pool.GetAvailableNodes(ctx, filter)
		if err != nil {
			t.Fatalf("GetAvailableNodes() failed: %v", err)
		}
		if len(nodes) != 0 {
			t.Errorf("expected 0 nodes, got %d", len(nodes))
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	pool := NewInMemoryNodePool()
	ctx := context.Background()

	// Test concurrent registration and unregistration
	done := make(chan bool, 10)

	for i := 0; i < 5; i++ {
		go func(id int) {
			req := &NodeRegistration{Name: "concurrent-node-" + string(rune('0'+id))}
			node, err := pool.RegisterNode(ctx, req)
			if err != nil {
				t.Errorf("concurrent RegisterNode() failed: %v", err)
			} else {
				// Immediately unregister
				pool.UnregisterNode(ctx, node.ID)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		go func() {
			pool.ListNodes(ctx, 0, 10)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
