package core

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		manager := NewManager()

		assert.NotNil(t, manager)
		assert.Equal(t, 0, manager.ModuleCount())
		assert.Equal(t, 0, manager.TypeCount())
		assert.False(t, manager.IsEnabled())
		assert.Empty(t, manager.ListModules())
		assert.Empty(t, manager.ListTypes())
	})

	t.Run("register_module", func(t *testing.T) {
		manager := NewManager()
		module := NewTestModule("test", "test-module")

		err := manager.Register(module)
		require.NoError(t, err)

		assert.Equal(t, 1, manager.ModuleCount())
		assert.Equal(t, 1, manager.TypeCount())
		assert.Contains(t, manager.ListModules(), "test-module")
		assert.Contains(t, manager.ListTypes(), "test")

		// Test retrieval
		retrieved, exists := manager.GetModule("test-module")
		assert.True(t, exists)
		assert.Equal(t, module, retrieved)

		// Test by type
		typeModules := manager.GetModulesByType("test")
		assert.Len(t, typeModules, 1)
		assert.Equal(t, module, typeModules[0])
	})

	t.Run("register_nil_module", func(t *testing.T) {
		manager := NewManager()

		err := manager.Register(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot register nil module")
	})

	t.Run("register_empty_name", func(t *testing.T) {
		manager := NewManager()
		module := NewTestModule("test", "")

		err := manager.Register(module)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "module name cannot be empty")
	})

	t.Run("register_empty_type", func(t *testing.T) {
		manager := NewManager()
		module := NewTestModule("", "test-module")

		err := manager.Register(module)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "module type cannot be empty")
	})

	t.Run("register_duplicate_name", func(t *testing.T) {
		manager := NewManager()
		module1 := NewTestModule("test", "duplicate-name")
		module2 := NewTestModule("other", "duplicate-name")

		err := manager.Register(module1)
		require.NoError(t, err)

		err = manager.Register(module2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})

	t.Run("unregister_module", func(t *testing.T) {
		manager := NewManager()
		module := NewTestModule("test", "test-module")

		// Register first
		err := manager.Register(module)
		require.NoError(t, err)

		// Unregister
		err = manager.Unregister("test-module")
		require.NoError(t, err)

		assert.Equal(t, 0, manager.ModuleCount())
		assert.Equal(t, 0, manager.TypeCount())

		// Should not exist anymore
		_, exists := manager.GetModule("test-module")
		assert.False(t, exists)

		typeModules := manager.GetModulesByType("test")
		assert.Empty(t, typeModules)
	})

	t.Run("unregister_active_module", func(t *testing.T) {
		manager := NewManager()
		module := NewTestModule("test", "test-module")

		// Register and enable
		err := manager.Register(module)
		require.NoError(t, err)

		ctx := context.Background()
		err = module.Enable(ctx)
		require.NoError(t, err)
		assert.True(t, module.IsActive())

		// Unregister should disable first
		err = manager.Unregister("test-module")
		require.NoError(t, err)

		assert.True(t, module.disableCalled)
		assert.False(t, module.IsActive())
	})

	t.Run("unregister_nonexistent", func(t *testing.T) {
		manager := NewManager()

		err := manager.Unregister("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("get_default_module", func(t *testing.T) {
		manager := NewManager()

		// Register modules with different priorities
		module1 := NewTestModule("test", "low-priority")
		module1.SetPriority(10)

		module2 := NewTestModule("test", "high-priority")
		module2.SetPriority(100)

		module3 := NewTestModule("test", "medium-priority")
		module3.SetPriority(50)

		err := manager.Register(module1)
		require.NoError(t, err)

		err = manager.Register(module2)
		require.NoError(t, err)

		err = manager.Register(module3)
		require.NoError(t, err)

		// Should return highest priority module
		defaultModule, exists := manager.GetDefaultModule("test")
		assert.True(t, exists)
		assert.Equal(t, module2, defaultModule)

		// Non-existent type
		_, exists = manager.GetDefaultModule("nonexistent")
		assert.False(t, exists)
	})
}

func TestManagerDependencyResolution(t *testing.T) {
	t.Run("simple_dependency", func(t *testing.T) {
		manager := NewManager()

		// Create modules with dependencies
		moduleA := NewTestModule("typeA", "moduleA")
		moduleB := NewTestModule("typeB", "moduleB")
		moduleB.SetRequires("typeA")

		err := manager.Register(moduleA)
		require.NoError(t, err)

		err = manager.Register(moduleB)
		require.NoError(t, err)

		// Resolve load order
		loadOrder, err := manager.ResolveLoadOrder()
		require.NoError(t, err)

		assert.Len(t, loadOrder, 2)
		assert.Equal(t, "moduleA", loadOrder[0]) // A should come before B
		assert.Equal(t, "moduleB", loadOrder[1])
	})

	t.Run("chain_dependency", func(t *testing.T) {
		manager := NewManager()

		moduleA := NewTestModule("typeA", "moduleA")
		moduleB := NewTestModule("typeB", "moduleB")
		moduleC := NewTestModule("typeC", "moduleC")

		moduleB.SetRequires("typeA")
		moduleC.SetRequires("typeB")

		err := manager.Register(moduleA)
		require.NoError(t, err)

		err = manager.Register(moduleB)
		require.NoError(t, err)

		err = manager.Register(moduleC)
		require.NoError(t, err)

		loadOrder, err := manager.ResolveLoadOrder()
		require.NoError(t, err)

		assert.Len(t, loadOrder, 3)
		assert.Equal(t, "moduleA", loadOrder[0])
		assert.Equal(t, "moduleB", loadOrder[1])
		assert.Equal(t, "moduleC", loadOrder[2])
	})

	t.Run("priority_ordering", func(t *testing.T) {
		manager := NewManager()

		moduleA := NewTestModule("typeA", "moduleA")
		moduleA.SetPriority(10)

		moduleB := NewTestModule("typeB", "moduleB")
		moduleB.SetPriority(100)

		// No dependencies, so should be ordered by priority
		err := manager.Register(moduleA)
		require.NoError(t, err)

		err = manager.Register(moduleB)
		require.NoError(t, err)

		loadOrder, err := manager.ResolveLoadOrder()
		require.NoError(t, err)

		assert.Len(t, loadOrder, 2)
		assert.Equal(t, "moduleB", loadOrder[0]) // Higher priority first
		assert.Equal(t, "moduleA", loadOrder[1])
	})

	t.Run("missing_dependency", func(t *testing.T) {
		manager := NewManager()

		moduleB := NewTestModule("typeB", "moduleB")
		moduleB.SetRequires("typeA") // typeA doesn't exist

		err := manager.Register(moduleB)
		require.NoError(t, err)

		_, err = manager.ResolveLoadOrder()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no modules of that type are registered")
	})

	t.Run("circular_dependency", func(t *testing.T) {
		manager := NewManager()

		// Create circular dependency
		moduleA := NewTestModule("typeA", "moduleA")
		moduleB := NewTestModule("typeB", "moduleB")

		moduleA.SetRequires("typeB")
		moduleB.SetRequires("typeA")

		err := manager.Register(moduleA)
		require.NoError(t, err)

		err = manager.Register(moduleB)
		require.NoError(t, err)

		_, err = manager.ResolveLoadOrder()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "circular dependency detected")
	})
}

func TestManagerLifecycle(t *testing.T) {
	t.Run("enable_all_success", func(t *testing.T) {
		manager := NewManager()
		ctx := context.Background()

		moduleA := NewTestModule("typeA", "moduleA")
		moduleB := NewTestModule("typeB", "moduleB")
		moduleB.SetRequires("typeA")

		err := manager.Register(moduleA)
		require.NoError(t, err)

		err = manager.Register(moduleB)
		require.NoError(t, err)

		// Enable all
		err = manager.EnableAll(ctx)
		require.NoError(t, err)

		assert.True(t, manager.IsEnabled())

		// Check that modules were enabled in correct order
		assert.True(t, moduleA.enableCalled)
		assert.True(t, moduleB.enableCalled)
		assert.True(t, moduleA.IsActive())
		assert.True(t, moduleB.IsActive())
	})

	t.Run("enable_all_with_error", func(t *testing.T) {
		manager := NewManager()
		ctx := context.Background()

		moduleA := NewTestModule("typeA", "moduleA")
		moduleB := NewTestModule("typeB", "moduleB")
		moduleB.SetRequires("typeA")
		moduleB.enableError = assert.AnError

		err := manager.Register(moduleA)
		require.NoError(t, err)

		err = manager.Register(moduleB)
		require.NoError(t, err)

		// Enable all should fail
		err = manager.EnableAll(ctx)
		assert.Error(t, err)

		// Should be a ModuleError
		var moduleErr *ModuleError
		assert.ErrorAs(t, err, &moduleErr)
		assert.Equal(t, "typeB", moduleErr.ModuleType)
		assert.Equal(t, "moduleB", moduleErr.ModuleName)
		assert.Equal(t, "enable", moduleErr.Operation)

		assert.False(t, manager.IsEnabled())
	})

	t.Run("disable_all_success", func(t *testing.T) {
		manager := NewManager()
		ctx := context.Background()

		moduleA := NewTestModule("typeA", "moduleA")
		moduleB := NewTestModule("typeB", "moduleB")
		moduleB.SetRequires("typeA")

		err := manager.Register(moduleA)
		require.NoError(t, err)

		err = manager.Register(moduleB)
		require.NoError(t, err)

		// Enable all first
		err = manager.EnableAll(ctx)
		require.NoError(t, err)

		// Disable all
		err = manager.DisableAll(ctx)
		require.NoError(t, err)

		assert.False(t, manager.IsEnabled())

		// Check that modules were disabled in reverse order
		assert.True(t, moduleA.disableCalled)
		assert.True(t, moduleB.disableCalled)
		assert.False(t, moduleA.IsActive())
		assert.False(t, moduleB.IsActive())
	})

	t.Run("disable_all_when_not_enabled", func(t *testing.T) {
		manager := NewManager()
		ctx := context.Background()

		moduleA := NewTestModule("typeA", "moduleA")
		err := manager.Register(moduleA)
		require.NoError(t, err)

		// Disable without enabling should not error
		err = manager.DisableAll(ctx)
		require.NoError(t, err)

		assert.False(t, moduleA.disableCalled)
	})
}

func TestManagerResourcePaths(t *testing.T) {
	t.Run("set_and_get_resource_path", func(t *testing.T) {
		manager := NewManager()

		manager.SetResourcePath("test-module", "/path/to/resources")

		path, exists := manager.GetResourcePath("test-module")
		assert.True(t, exists)
		assert.Equal(t, "/path/to/resources", path)
	})

	t.Run("get_nonexistent_resource_path", func(t *testing.T) {
		manager := NewManager()

		path, exists := manager.GetResourcePath("nonexistent")
		assert.False(t, exists)
		assert.Empty(t, path)
	})
}

func TestManagerConcurrency(t *testing.T) {
	t.Run("concurrent_register", func(t *testing.T) {
		manager := NewManager()

		// Register modules concurrently
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(i int) {
				module := NewTestModule("test", fmt.Sprintf("module-%d", i))
				err := manager.Register(module)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		assert.Equal(t, 10, manager.ModuleCount())
	})

	t.Run("concurrent_read_write", func(t *testing.T) {
		manager := NewManager()

		// Add some initial modules
		for i := 0; i < 5; i++ {
			module := NewTestModule("test", fmt.Sprintf("module-%d", i))
			err := manager.Register(module)
			require.NoError(t, err)
		}

		done := make(chan bool, 20)

		// Concurrent readers
		for i := 0; i < 10; i++ {
			go func() {
				_ = manager.ListModules()
				_ = manager.ListTypes()
				_, _ = manager.GetDefaultModule("test")
				done <- true
			}()
		}

		// Concurrent resource path operations
		for i := 0; i < 10; i++ {
			go func(i int) {
				manager.SetResourcePath(fmt.Sprintf("module-%d", i%5), fmt.Sprintf("/path/%d", i))
				_, _ = manager.GetResourcePath(fmt.Sprintf("module-%d", i%5))
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 20; i++ {
			<-done
		}

		// Should still have all modules
		assert.Equal(t, 5, manager.ModuleCount())
	})
}
