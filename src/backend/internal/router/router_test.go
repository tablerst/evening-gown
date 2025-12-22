package router

import (
	"testing"

	"go.uber.org/goleak"
)

func TestNew_DoesNotLeakGoroutines(t *testing.T) {
	defer goleak.VerifyNone(t)

	_ = New(Dependencies{})
}
