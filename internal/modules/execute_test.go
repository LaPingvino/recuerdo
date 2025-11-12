package modules

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LaPingvino/openteacher/internal/core"
)

func TestExecuteModule(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		module := NewExecuteModule()

		assert.Equal(t, "execute", module.Type())
		assert.Equal(t, "execute-module", module.Name())
		assert.Equal(t, []string{"event"}, module.Requires())
		assert.Equal(t, 1000, module.Priority())
		assert.False(t, module.IsActive())
		assert.False(t, module.IsRunning())
		assert.Equal(t, "all", module.GetProfile())
	})

	t.Run("profile_management", func(t *testing.T) {
		module := NewExecuteModule()

		// Test setting valid profile
		err := module.SetProfile("gui")
		require.NoError(t, err)
		assert.Equal(t, "gui", module.GetProfile())

		// Test setting another profile
		err = module.SetProfile("cli")
		require.NoError(t, err)
		assert.Equal(t, "cli", module.GetProfile())

		// Test empty profile should fail
		err = module.SetProfile("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "profile cannot be empty")
		// Profile should remain unchanged
		assert.Equal(t, "cli", module.GetProfile())
	})

	t.Run("lifecycle", func(t *testing.T) {
		ctx := context.Background()
		module := NewExecuteModule()

		// Enable module
		err := module.Enable(ctx)
		require.NoError(t, err)
		assert.True(t, module.IsActive())
		assert.False(t, module.IsRunning()) // Not running until StartRunning is called

		// Disable module
		err = module.Disable(ctx)
		require.NoError(t, err)
		assert.False(t, module.IsActive())
		assert.False(t, module.IsRunning())
	})

	t.Run("start_running_with_quick_cancel", func(t *testing.T) {
		module := NewExecuteModule()

		// Enable first
		ctx := context.Background()
		err := module.Enable(ctx)
		require.NoError(t, err)

		// Create context that cancels quickly
		runCtx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Start running - should return when context is cancelled
		err = module.StartRunning(runCtx)
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)

		// Module should not be running anymore
		assert.False(t, module.IsRunning())
	})

	t.Run("start_running_already_running", func(t *testing.T) {
		module := NewExecuteModule()

		// Enable first
		ctx := context.Background()
		err := module.Enable(ctx)
		require.NoError(t, err)

		// Create a context that we can cancel
		runCtx, cancel := context.WithCancel(context.Background())

		// Start in goroutine
		errChan := make(chan error, 1)
		go func() {
			errChan <- module.StartRunning(runCtx)
		}()

		// Wait a bit for it to start
		time.Sleep(50 * time.Millisecond)
		assert.True(t, module.IsRunning())

		// Try to start again - should fail
		err = module.StartRunning(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already running")

		// Cancel and wait for first run to complete
		cancel()
		err = <-errChan
		assert.Equal(t, context.Canceled, err)
		assert.False(t, module.IsRunning())
	})

	t.Run("interface_compliance", func(t *testing.T) {
		module := NewExecuteModule()

		// Should implement Module interface
		var _ core.Module = module

		// Should implement ExecuteModule interface
		var _ core.ExecuteModule = module

		// Test ExecuteModule methods
		assert.Equal(t, "all", module.GetProfile())
		err := module.SetProfile("test")
		require.NoError(t, err)
		assert.Equal(t, "test", module.GetProfile())
	})

	t.Run("event_module_integration", func(t *testing.T) {
		module := NewExecuteModule()
		eventModule := NewEventModule()

		// Set event module
		module.SetEventModule(eventModule)

		// The execute module should now have start/stop events
		// We can't directly test this without exposing internal fields,
		// but we can verify the module still works
		ctx := context.Background()
		err := module.Enable(ctx)
		require.NoError(t, err)

		// This should work without panicking
		runCtx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err = module.StartRunning(runCtx)
		assert.Error(t, err) // Expected timeout error
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestExecuteModuleConcurrency(t *testing.T) {
	t.Run("concurrent_profile_changes", func(t *testing.T) {
		module := NewExecuteModule()

		done := make(chan bool, 10)

		// Concurrent profile changes
		for i := 0; i < 10; i++ {
			go func(i int) {
				profile := []string{"gui", "cli", "all", "test", "minimal"}[i%5]
				err := module.SetProfile(profile)
				assert.NoError(t, err)
				done <- true
			}(i)
		}

		// Wait for all changes
		for i := 0; i < 10; i++ {
			<-done
		}

		// Should have some valid profile
		profile := module.GetProfile()
		assert.NotEmpty(t, profile)
		assert.Contains(t, []string{"gui", "cli", "all", "test", "minimal"}, profile)
	})

	t.Run("concurrent_enable_disable", func(t *testing.T) {
		module := NewExecuteModule()
		ctx := context.Background()

		done := make(chan bool, 20)

		// Concurrent enable/disable operations
		for i := 0; i < 10; i++ {
			go func() {
				err := module.Enable(ctx)
				assert.NoError(t, err)
				done <- true
			}()

			go func() {
				err := module.Disable(ctx)
				assert.NoError(t, err)
				done <- true
			}()
		}

		// Wait for all operations
		for i := 0; i < 20; i++ {
			<-done
		}

		// Module should be in a consistent state (either enabled or disabled)
		// We don't test specific state since operations could interleave
		assert.NotPanics(t, func() {
			_ = module.IsActive()
			_ = module.IsRunning()
		})
	})
}
