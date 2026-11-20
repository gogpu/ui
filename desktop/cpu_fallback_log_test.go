package desktop

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/gogpu/gg"
)

// captureLog sends the standard logger to a buffer for the rest of the test.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	out, flags := log.Writer(), log.Flags()
	log.SetOutput(&buf)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(out)
		log.SetFlags(flags)
	})
	return &buf
}

// Without a GPU accelerator every boundary flush returns gg.ErrFallbackToCPU.
// It must be reported once, not once per boundary per frame.
func TestFlushBoundaryToTexture_CPUFallbackLoggedOnce(t *testing.T) {
	cc := gg.NewContext(100, 100)
	t.Cleanup(func() {
		if err := cc.Close(); err != nil {
			t.Errorf("close context: %v", err)
		}
	})
	root, children := buildTestLayerTree(3, 48, 48)
	rl := newRenderLoopWithTextures(root)
	probe := rl.boundaryTextures[children[0].BoundaryCacheKey()].texture
	if err := cc.FlushGPUWithView(probe, 48, 48); !errors.Is(err, gg.ErrFallbackToCPU) {
		t.Skipf("FlushGPUWithView = %v; this test needs a context without a GPU accelerator", err)
	}
	buf := captureLog(t)

	const frames = 5
	for range frames {
		for _, pic := range children {
			rl.flushBoundaryToTexture(pic, rl.boundaryTextures[pic.BoundaryCacheKey()], dummyScene(), cc, 48, 48)
		}
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("%d flushes fell back to CPU and logged %d lines, want 1:\n%s",
			frames*len(children), len(lines), buf.String())
	}
	if !strings.Contains(lines[0], "github.com/gogpu/gg/gpu") {
		t.Errorf("log line %q does not say how to enable the GPU accelerator", lines[0])
	}
}

// Other flush errors are still logged every time they happen.
func TestLogBoundaryFlushError_OtherErrorsAlwaysLogged(t *testing.T) {
	rl := &renderLoop{}
	buf := captureLog(t)
	lost := errors.New("device lost")

	rl.logBoundaryFlushError(7, gg.ErrFallbackToCPU)
	rl.logBoundaryFlushError(7, lost)
	rl.logBoundaryFlushError(8, lost)
	rl.logBoundaryFlushError(9, gg.ErrFallbackToCPU)

	got := buf.String()
	if n := strings.Count(got, lost.Error()); n != 2 {
		t.Errorf("%q logged %d times, want 2:\n%s", lost, n, got)
	}
	if n := strings.Count(got, gg.ErrFallbackToCPU.Error()); n != 1 {
		t.Errorf("%q logged %d times, want 1:\n%s", gg.ErrFallbackToCPU, n, got)
	}
}
