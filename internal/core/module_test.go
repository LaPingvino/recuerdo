package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModule is a simple test implementation of the Module interface
type TestModule struct {
	*BaseModule
	enableCalled  bool
	disableCalled bool
	enableError   error
	disableError  error
}

func NewTestModule(moduleType, moduleName string) *TestModule {
	return &TestModule{
		BaseModule: NewBaseModule(moduleType, moduleName),
	}
}

func (t *TestModule) Enable(ctx context.Context) error {
	t.enableCalled = true
	if t.enableError != nil {
		return t.enableError
	}
	return t.BaseModule.Enable(ctx)
}

func (t *TestModule) Disable(ctx context.Context) error {
	t.disableCalled = true
	if t.disableError != nil {
		return t.disableError
	}
	return t.BaseModule.Disable(ctx)
}

func TestBaseModule(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		module := NewBaseModule("test", "test-module")

		assert.Equal(t, "test", module.Type())
		assert.Equal(t, "test-module", module.Name())
		assert.Empty(t, module.Requires())
		assert.Empty(t, module.Uses())
		assert.Equal(t, 0, module.Priority())
		assert.False(t, module.IsActive())
	})

	t.Run("dependencies", func(t *testing.T) {
		module := NewBaseModule("test", "test-module")

		module.SetRequires("dependency1", "dependency2")
		module.SetUses("optional1", "optional2")

		requires := module.Requires()
		uses := module.Uses()

		assert.Equal(t, []string{"dependency1", "dependency2"}, requires)
		assert.Equal(t, []string{"optional1", "optional2"}, uses)

		// Test that returned slices are copies (defensive copying)
		requires[0] = "modified"
		assert.Equal(t, []string{"dependency1", "dependency2"}, module.Requires())
	})

	t.Run("priority", func(t *testing.T) {
		module := NewBaseModule("test", "test-module")

		module.SetPriority(100)
		assert.Equal(t, 100, module.Priority())
	})

	t.Run("lifecycle", func(t *testing.T) {
		ctx := context.Background()
		module := NewBaseModule("test", "test-module")

		// Initially inactive
		assert.False(t, module.IsActive())

		// Enable
		err := module.Enable(ctx)
		require.NoError(t, err)
		assert.True(t, module.IsActive())

		// Disable
		err = module.Disable(ctx)
		require.NoError(t, err)
		assert.False(t, module.IsActive())
	})
}

func TestTestModule(t *testing.T) {
	t.Run("lifecycle_calls", func(t *testing.T) {
		ctx := context.Background()
		module := NewTestModule("test", "test-module")

		// Enable
		err := module.Enable(ctx)
		require.NoError(t, err)
		assert.True(t, module.enableCalled)
		assert.True(t, module.IsActive())

		// Disable
		err = module.Disable(ctx)
		require.NoError(t, err)
		assert.True(t, module.disableCalled)
		assert.False(t, module.IsActive())
	})

	t.Run("enable_error", func(t *testing.T) {
		ctx := context.Background()
		module := NewTestModule("test", "test-module")

		expectedError := assert.AnError
		module.enableError = expectedError

		err := module.Enable(ctx)
		assert.Equal(t, expectedError, err)
		assert.True(t, module.enableCalled)
		assert.False(t, module.IsActive()) // Should remain inactive on error
	})

	t.Run("disable_error", func(t *testing.T) {
		ctx := context.Background()
		module := NewTestModule("test", "test-module")

		// First enable successfully
		err := module.Enable(ctx)
		require.NoError(t, err)

		// Then set up disable error
		expectedError := assert.AnError
		module.disableError = expectedError

		err = module.Disable(ctx)
		assert.Equal(t, expectedError, err)
		assert.True(t, module.disableCalled)
		assert.True(t, module.IsActive()) // Should remain active on disable error
	})
}

func TestModuleError(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		originalErr := assert.AnError
		moduleErr := NewModuleError("test", "test-module", "enable", originalErr)

		assert.Equal(t, "test", moduleErr.ModuleType)
		assert.Equal(t, "test-module", moduleErr.ModuleName)
		assert.Equal(t, "enable", moduleErr.Operation)
		assert.Equal(t, originalErr, moduleErr.Err)
	})

	t.Run("error_message", func(t *testing.T) {
		originalErr := assert.AnError
		moduleErr := NewModuleError("test", "test-module", "enable", originalErr)

		expected := "module test-module (test) failed during enable: assert.AnError general error for testing"
		assert.Equal(t, expected, moduleErr.Error())
	})

	t.Run("unwrap", func(t *testing.T) {
		originalErr := assert.AnError
		moduleErr := NewModuleError("test", "test-module", "enable", originalErr)

		assert.Equal(t, originalErr, moduleErr.Unwrap())
	})
}

// TestEventHandler is a helper for testing event handlers
type TestEventHandler struct {
	called    bool
	lastData  interface{}
	returnErr error
	callCount int
}

func (h *TestEventHandler) Handle(data interface{}) error {
	h.called = true
	h.lastData = data
	h.callCount++
	return h.returnErr
}

func (h *TestEventHandler) HandlerFunc() EventHandler {
	return func(data interface{}) error {
		return h.Handle(data)
	}
}

func TestEventHandlerExecution(t *testing.T) {
	t.Run("handler_creation", func(t *testing.T) {
		handler := &TestEventHandler{}
		handlerFunc := handler.HandlerFunc()

		assert.NotNil(t, handlerFunc)
		assert.False(t, handler.called)
		assert.Equal(t, 0, handler.callCount)
	})

	t.Run("handler_execution", func(t *testing.T) {
		handler := &TestEventHandler{}
		handlerFunc := handler.HandlerFunc()

		testData := "test data"
		err := handlerFunc(testData)

		require.NoError(t, err)
		assert.True(t, handler.called)
		assert.Equal(t, testData, handler.lastData)
		assert.Equal(t, 1, handler.callCount)
	})

	t.Run("handler_error", func(t *testing.T) {
		handler := &TestEventHandler{
			returnErr: assert.AnError,
		}
		handlerFunc := handler.HandlerFunc()

		err := handlerFunc("test data")

		assert.Equal(t, assert.AnError, err)
		assert.True(t, handler.called)
		assert.Equal(t, 1, handler.callCount)
	})
}

// Integration test to ensure our interfaces work together
func TestModuleInterfaceCompatibility(t *testing.T) {
	t.Run("module_interface_compliance", func(t *testing.T) {
		// Create a module and ensure it implements Module interface
		var module Module = NewTestModule("test", "test-module")

		assert.Equal(t, "test", module.Type())
		assert.Equal(t, "test-module", module.Name())
		assert.Empty(t, module.Requires())
		assert.Empty(t, module.Uses())
		assert.Equal(t, 0, module.Priority())
		assert.False(t, module.IsActive())

		ctx := context.Background()
		err := module.Enable(ctx)
		require.NoError(t, err)
		assert.True(t, module.IsActive())

		err = module.Disable(ctx)
		require.NoError(t, err)
		assert.False(t, module.IsActive())
	})
}
