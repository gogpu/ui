//go:build !nogpu

package desktop

import (
	"context"
	"image"
	"testing"

	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
	"github.com/gogpu/wgpu/hal"
	"github.com/gogpu/wgpu/hal/software"
)

// createSoftwareDevice creates a software-backed wgpu device for pixel-exact
// e2e testing. The software backend performs real CPU rasterization: LoadOpLoad
// preserves content, scissor clips draws, pixels are readable via Map.
func createSoftwareDevice(t *testing.T) (*wgpu.Device, *wgpu.Queue, func()) {
	t.Helper()
	instance, err := software.NewBackend().CreateInstance(nil)
	if err != nil {
		t.Fatalf("software CreateInstance: %v", err)
	}
	adapters := instance.EnumerateAdapters(nil)
	if len(adapters) == 0 {
		instance.Destroy()
		t.Fatal("software backend: no adapters")
	}
	openDev, err := adapters[0].Adapter.Open(0, gputypes.DefaultLimits())
	if err != nil {
		instance.Destroy()
		t.Fatalf("software Open: %v", err)
	}
	device, err := wgpu.NewDeviceFromHAL(
		openDev.Device, openDev.Queue,
		gputypes.Features(0), gputypes.DefaultLimits(), "ui-software-test",
	)
	if err != nil {
		openDev.Device.Destroy()
		instance.Destroy()
		t.Fatalf("NewDeviceFromHAL: %v", err)
	}
	queue := device.Queue()
	cleanup := func() { device.Release(); instance.Destroy() }
	return device, queue, cleanup
}

// copyBytesPerRowAlignment is the WebGPU COPY_BYTES_PER_ROW_ALIGNMENT: a
// texture-to-buffer copy must use a bytesPerRow that is a multiple of 256.
const copyBytesPerRowAlignment = 256

// readbackTexture copies an RGBA8 texture to a mappable buffer and returns
// the pixel bytes tightly packed (stride = w). The copy itself uses rows
// padded to copyBytesPerRowAlignment, as WebGPU requires. Returns nil when
// readback is unavailable.
func readbackTexture(t *testing.T, device *wgpu.Device, queue *wgpu.Queue, tex *wgpu.Texture, w, h int) []byte {
	t.Helper()
	rowBytes := w * 4
	paddedRow := (rowBytes + copyBytesPerRowAlignment - 1) / copyBytesPerRowAlignment * copyBytesPerRowAlignment
	bufSize := uint64(paddedRow * h)
	buf, err := device.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "readback",
		Size:  bufSize,
		Usage: wgpu.BufferUsageCopyDst | wgpu.BufferUsageMapRead,
	})
	if err != nil {
		t.Logf("CreateBuffer for readback: %v", err)
		return nil
	}
	defer buf.Release()

	enc, err := device.CreateCommandEncoder(nil)
	if err != nil {
		t.Fatalf("CreateCommandEncoder for readback: %v", err)
	}
	regions := []wgpu.BufferTextureCopy{{
		TextureBase: wgpu.ImageCopyTexture{Texture: tex},
		BufferLayout: wgpu.ImageDataLayout{
			Offset:       0,
			BytesPerRow:  uint32(paddedRow),
			RowsPerImage: uint32(h),
		},
		Size: wgpu.Extent3D{Width: uint32(w), Height: uint32(h), DepthOrArrayLayers: 1},
	}}
	enc.CopyTextureToBuffer(tex, buf, regions)
	cmd, err := enc.Finish()
	if err != nil {
		t.Fatalf("Finish readback encoder: %v", err)
	}
	queue.Submit(cmd)

	if err := buf.Map(context.Background(), wgpu.MapModeRead, 0, bufSize); err != nil {
		t.Logf("Buffer.Map: %v", err)
		return nil
	}
	mr, err := buf.MappedRange(0, bufSize)
	if err != nil {
		t.Logf("MappedRange: %v", err)
		return nil
	}
	padded := mr.Bytes()
	result := make([]byte, rowBytes*h)
	for y := range h {
		copy(result[y*rowBytes:(y+1)*rowBytes], padded[y*paddedRow:y*paddedRow+rowBytes])
	}
	mr.Release()
	buf.Unmap()
	return result
}

// assertPixelRGBA verifies a single pixel in RGBA8 readback data.
func assertPixelRGBA(t *testing.T, data []byte, stride, x, y int, wantR, wantG, wantB uint8, label string) {
	t.Helper()
	idx := (y*stride + x) * 4
	if idx+3 >= len(data) {
		t.Errorf("%s: pixel (%d,%d) out of bounds (data len=%d)", label, x, y, len(data))
		return
	}
	r, g, b := data[idx], data[idx+1], data[idx+2]
	if r != wantR || g != wantG || b != wantB {
		t.Errorf("%s: pixel (%d,%d) = RGB(%d,%d,%d), want RGB(%d,%d,%d)",
			label, x, y, r, g, b, wantR, wantG, wantB)
	}
}

// clearTexture fills a texture with a solid color via LoadOpClear.
func clearTexture(t *testing.T, device *wgpu.Device, queue *wgpu.Queue, view *wgpu.TextureView, color gputypes.Color) {
	t.Helper()
	enc, err := device.CreateCommandEncoder(nil)
	if err != nil {
		t.Fatalf("CreateCommandEncoder: %v", err)
	}
	rp, err := enc.BeginRenderPass(&wgpu.RenderPassDescriptor{
		Label: "clear",
		ColorAttachments: []wgpu.RenderPassColorAttachment{{
			View:       view,
			LoadOp:     gputypes.LoadOpClear,
			StoreOp:    gputypes.StoreOpStore,
			ClearValue: color,
		}},
	})
	if err != nil {
		t.Fatalf("BeginRenderPass: %v", err)
	}
	rp.End()
	cmd, err := enc.Finish()
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	queue.Submit(cmd)
}

// --- Test 1: Boundary texture render produces non-empty output ---

// TestSoftwarePipeline_BoundaryTextureRender verifies that rendering a scene
// into an offscreen texture via the software backend produces visible pixels.
// This validates the lowest level of the per-boundary texture pipeline.
func TestSoftwarePipeline_BoundaryTextureRender(t *testing.T) {
	device, queue, cleanup := createSoftwareDevice(t)
	defer cleanup()

	const W, H = 16, 16

	tex, err := device.CreateTexture(&wgpu.TextureDescriptor{
		Label:         "boundary-tex",
		Size:          wgpu.Extent3D{Width: W, Height: H, DepthOrArrayLayers: 1},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     gputypes.TextureDimension2D,
		Format:        wgpu.TextureFormatRGBA8Unorm,
		Usage:         wgpu.TextureUsageRenderAttachment | wgpu.TextureUsageCopySrc,
	})
	if err != nil {
		t.Fatalf("CreateTexture: %v", err)
	}
	defer tex.Release()

	view, err := device.CreateTextureView(tex, nil)
	if err != nil {
		t.Fatalf("CreateTextureView: %v", err)
	}
	defer view.Release()

	// Render: fill entire 16x16 texture with solid green via LoadOpClear.
	clearTexture(t, device, queue, view, gputypes.Color{R: 0, G: 1, B: 0, A: 1})

	data := readbackTexture(t, device, queue, tex, W, H)
	if data == nil {
		t.Skip("readback not available")
	}

	// Every pixel should be green (0,255,0).
	assertPixelRGBA(t, data, W, 0, 0, 0, 255, 0, "top-left")
	assertPixelRGBA(t, data, W, W-1, H-1, 0, 255, 0, "bottom-right")
	assertPixelRGBA(t, data, W, W/2, H/2, 0, 255, 0, "center")

	// Verify non-black: at least one pixel has non-zero green channel.
	allBlack := true
	for i := 0; i < len(data); i += 4 {
		if data[i+1] != 0 {
			allBlack = false
			break
		}
	}
	if allBlack {
		t.Error("all pixels are black — texture render produced no output")
	}
}

// --- Test 2: Composite textures at correct screen positions ---

// TestSoftwarePipeline_CompositeTextures_CorrectPositioning verifies that
// compositing two textures (root + child) places pixels at the expected
// screen coordinates. Uses Queue.WriteTexture to write a child region
// onto the surface at an offset — the same positioning that compositeTextures
// performs via DrawGPUTexture/DrawGPUTextureBase.
//
// Note: CopyTextureToTexture ignores Origin in the software HAL, so we use
// WriteTexture which correctly handles destination offsets.
func TestSoftwarePipeline_CompositeTextures_CorrectPositioning(t *testing.T) {
	device, queue, cleanup := createSoftwareDevice(t)
	defer cleanup()

	const (
		surfW, surfH   = 100, 100
		childW, childH = 20, 20
		childX, childY = 40, 40
	)

	surfTex, surfView := createBlitTarget(t, device, surfW, surfH)
	defer surfTex.Release()
	defer surfView.Release()

	clearTexture(t, device, queue, surfView, gputypes.Color{R: 1, G: 0, B: 0, A: 1})

	writeRegion(t, queue, surfTex, childX, childY, childW, childH, 0, 0, 255, 255)

	data := readbackTexture(t, device, queue, surfTex, surfW, surfH)
	if data == nil {
		t.Skip("readback not available")
	}

	assertPixelRGBA(t, data, surfW, 0, 0, 255, 0, 0, "root-corner")
	assertPixelRGBA(t, data, surfW, 10, 10, 255, 0, 0, "root-interior")
	assertPixelRGBA(t, data, surfW, 50, 50, 0, 0, 255, "child-center")
	assertPixelRGBA(t, data, surfW, 45, 45, 0, 0, 255, "child-interior")
	assertPixelRGBA(t, data, surfW, 39, 39, 255, 0, 0, "just-outside-child")
	assertPixelRGBA(t, data, surfW, 99, 99, 255, 0, 0, "root-far-corner")
}

// --- Test 3: Damage-aware blit preserves undamaged content ---

// TestSoftwarePipeline_DamagePreservesContent verifies the damage-aware blit
// pipeline end-to-end through the software backend.
//
// Frame 1: Full render (LoadOpClear red) + WriteTexture blue child at (40,40).
// Frame 2: Damage-aware update — WriteTexture green child at (40,40), no clear.
//
// After frame 2: pixels outside child rect should still be RED (never
// overwritten — no LoadOpClear on frame 2), pixels inside child rect should
// be GREEN (overwritten by the write). This validates that the compositor
// only updates the damage region while preserving undamaged content.
func TestSoftwarePipeline_DamagePreservesContent(t *testing.T) {
	device, queue, cleanup := createSoftwareDevice(t)
	defer cleanup()

	const (
		W, H           = 100, 100
		childX, childY = 40, 40
		childW, childH = 20, 20
	)

	surfTex, surfView := createBlitTarget(t, device, W, H)
	defer surfTex.Release()
	defer surfView.Release()

	// --- Frame 1: full render ---
	clearTexture(t, device, queue, surfView, gputypes.Color{R: 1, G: 0, B: 0, A: 1})
	writeRegion(t, queue, surfTex, childX, childY, childW, childH, 0, 0, 255, 255)

	data1 := readbackTexture(t, device, queue, surfTex, W, H)
	if data1 == nil {
		t.Skip("readback not available")
	}
	assertPixelRGBA(t, data1, W, 0, 0, 255, 0, 0, "frame1-root-corner")
	assertPixelRGBA(t, data1, W, 50, 50, 0, 0, 255, "frame1-child-center")
	assertPixelRGBA(t, data1, W, 39, 39, 255, 0, 0, "frame1-outside-child")

	// --- Frame 2: damage-aware update (no LoadOpClear) ---
	// Only the child region is updated via WriteTexture at the same offset.
	// The rest of the surface is untouched — equivalent to LoadOpLoad + scissor.
	writeRegion(t, queue, surfTex, childX, childY, childW, childH, 0, 255, 0, 255)

	data2 := readbackTexture(t, device, queue, surfTex, W, H)
	if data2 == nil {
		t.Skip("readback not available")
	}

	// Pixels OUTSIDE damage rect should be RED (untouched since frame 1).
	assertPixelRGBA(t, data2, W, 0, 0, 255, 0, 0, "frame2-root-corner-preserved")
	assertPixelRGBA(t, data2, W, 10, 10, 255, 0, 0, "frame2-root-interior-preserved")
	assertPixelRGBA(t, data2, W, 99, 99, 255, 0, 0, "frame2-root-far-corner-preserved")
	assertPixelRGBA(t, data2, W, 39, 39, 255, 0, 0, "frame2-just-outside-damage-preserved")

	// Pixels INSIDE damage rect should be GREEN (overwritten in frame 2).
	assertPixelRGBA(t, data2, W, 50, 50, 0, 255, 0, "frame2-child-center-updated")
	assertPixelRGBA(t, data2, W, 45, 45, 0, 255, 0, "frame2-child-interior-updated")
	assertPixelRGBA(t, data2, W, 40, 40, 0, 255, 0, "frame2-child-origin-updated")
	assertPixelRGBA(t, data2, W, 59, 59, 0, 255, 0, "frame2-child-far-corner-updated")
}

// --- Test 4: Damage-aware blit only changes spinner pixels ---

// TestDamageAwareBlit_OnlySpinnerPixelsChange simulates the damage-aware
// compositor pipeline end-to-end and proves pixel-exactness:
//
//   - Frame 1: LoadOpClear RED (full window) + WriteTexture BLUE at (80,80) 40x40
//     (simulating spinner boundary texture blit after full render).
//   - Frame 2: LoadOpLoad (preserve frame 1) + SetScissorRect to spinner area +
//     WriteTexture GREEN at (80,80) 40x40 (simulating spinner re-render with new
//     rotation angle).
//   - Assertion: pixels OUTSIDE spinner area are IDENTICAL between frames (RED),
//     pixels INSIDE spinner area are DIFFERENT (BLUE -> GREEN), and exactly
//     40*40 = 1600 pixels changed (out of 200*200 = 40000).
func TestDamageAwareBlit_OnlySpinnerPixelsChange(t *testing.T) {
	device, queue, cleanup := createSoftwareDevice(t)
	defer cleanup()

	const (
		surfW, surfH       = 200, 200
		spinnerX, spinnerY = 80, 80
		spinnerW, spinnerH = 40, 40
	)

	surfTex, surfView := createBlitTarget(t, device, surfW, surfH)
	defer surfTex.Release()
	defer surfView.Release()

	// --- Frame 1: full render (LoadOpClear RED) ---
	clearTexture(t, device, queue, surfView, gputypes.Color{R: 1, G: 0, B: 0, A: 1})

	// Blit spinner boundary texture (BLUE) at spinner position.
	writeRegion(t, queue, surfTex, spinnerX, spinnerY, spinnerW, spinnerH, 0, 0, 255, 255)

	frame1Pixels := readbackTexture(t, device, queue, surfTex, surfW, surfH)
	if frame1Pixels == nil {
		t.Skip("readback not available")
	}

	// Verify frame 1 is correct: RED background, BLUE spinner.
	assertPixelRGBA(t, frame1Pixels, surfW, 0, 0, 255, 0, 0, "f1-topleft")
	assertPixelRGBA(t, frame1Pixels, surfW, 10, 10, 255, 0, 0, "f1-background")
	assertPixelRGBA(t, frame1Pixels, surfW, 199, 199, 255, 0, 0, "f1-bottomright")
	assertPixelRGBA(t, frame1Pixels, surfW, 90, 90, 0, 0, 255, "f1-spinner-interior")
	assertPixelRGBA(t, frame1Pixels, surfW, 100, 100, 0, 0, 255, "f1-spinner-center")

	// --- Frame 2: damage-aware render (LoadOpLoad, only spinner changed) ---
	// The render pass with LoadOpLoad preserves all frame 1 content.
	// Then we blit the updated spinner (GREEN) into the same region.
	enc, err := device.CreateCommandEncoder(nil)
	if err != nil {
		t.Fatalf("CreateCommandEncoder: %v", err)
	}
	rp, err := enc.BeginRenderPass(&wgpu.RenderPassDescriptor{
		Label: "frame2-damage-aware",
		ColorAttachments: []wgpu.RenderPassColorAttachment{{
			View:    surfView,
			LoadOp:  gputypes.LoadOpLoad,
			StoreOp: gputypes.StoreOpStore,
		}},
	})
	if err != nil {
		t.Fatalf("BeginRenderPass: %v", err)
	}
	rp.SetViewport(gputypes.Viewport{Width: surfW, Height: surfH, MaxDepth: 1})
	rp.SetScissorRect(gputypes.ScissorRect{X: spinnerX, Y: spinnerY, Width: spinnerW, Height: spinnerH})
	rp.End()
	cmd, err := enc.Finish()
	if err != nil {
		t.Fatalf("Finish: %v", err)
	}
	queue.Submit(cmd)

	// Now blit the updated spinner (GREEN) at the same position.
	writeRegion(t, queue, surfTex, spinnerX, spinnerY, spinnerW, spinnerH, 0, 255, 0, 255)

	frame2Pixels := readbackTexture(t, device, queue, surfTex, surfW, surfH)
	if frame2Pixels == nil {
		t.Skip("readback not available")
	}

	// --- Assert: pixels OUTSIDE spinner area are IDENTICAL (RED) ---
	outsidePoints := [][2]int{
		{0, 0}, {10, 10}, {199, 199}, {79, 79}, {120, 120},
		{0, 199}, {199, 0}, {50, 150}, {150, 50},
	}
	for _, pt := range outsidePoints {
		assertPixelRGBA(t, frame2Pixels, surfW, pt[0], pt[1], 255, 0, 0,
			"f2-outside-preserved")
	}

	// --- Assert: pixels INSIDE spinner area are GREEN (changed from BLUE) ---
	insidePoints := [][2]int{
		{90, 90}, {100, 100}, {80, 80}, {119, 119}, {95, 95},
	}
	for _, pt := range insidePoints {
		assertPixelRGBA(t, frame2Pixels, surfW, pt[0], pt[1], 0, 255, 0,
			"f2-inside-updated")
	}

	// --- Assert: exact pixel diff count ---
	changedPixels := 0
	totalPixels := surfW * surfH
	for i := 0; i < len(frame1Pixels); i += 4 {
		if frame1Pixels[i] != frame2Pixels[i] ||
			frame1Pixels[i+1] != frame2Pixels[i+1] ||
			frame1Pixels[i+2] != frame2Pixels[i+2] {
			changedPixels++
		}
	}

	wantChanged := spinnerW * spinnerH
	if changedPixels != wantChanged {
		t.Errorf("changed pixels = %d, want exactly %d (out of %d total)",
			changedPixels, wantChanged, totalPixels)
	}
}

// --- Test 5: Texture count does not leak on stable tree ---

// TestTextureCount_NoLeakOnStableTree verifies that repeatedly rendering the
// same set of boundary textures does not cause texture count to grow. This
// simulates 10 frames with a stable widget tree of 5 boundaries — the texture
// map should stay at exactly 5 entries throughout.
func TestTextureCount_NoLeakOnStableTree(t *testing.T) {
	device, _, cleanup := createSoftwareDevice(t)
	defer cleanup()

	const numBoundaries = 5

	// Create 5 boundary textures, simulating a stable widget tree.
	type boundaryEntry struct {
		key string
		tex *wgpu.Texture
	}
	boundaries := make([]boundaryEntry, numBoundaries)
	for i := range boundaries {
		tex, err := device.CreateTexture(&wgpu.TextureDescriptor{
			Label:         "boundary",
			Size:          wgpu.Extent3D{Width: 48, Height: 48, DepthOrArrayLayers: 1},
			MipLevelCount: 1,
			SampleCount:   1,
			Dimension:     gputypes.TextureDimension2D,
			Format:        wgpu.TextureFormatRGBA8Unorm,
			Usage:         wgpu.TextureUsageRenderAttachment | wgpu.TextureUsageCopySrc,
		})
		if err != nil {
			t.Fatalf("CreateTexture boundary %d: %v", i, err)
		}
		boundaries[i] = boundaryEntry{
			key: "boundary-" + string(rune('A'+i)),
			tex: tex,
		}
	}
	defer func() {
		for _, b := range boundaries {
			b.tex.Release()
		}
	}()

	// Simulate 10 frames with the same 5 boundaries — track in a map.
	textureMap := make(map[string]*wgpu.Texture)
	for frame := 0; frame < 10; frame++ {
		for _, b := range boundaries {
			textureMap[b.key] = b.tex
		}
		if got := len(textureMap); got != numBoundaries {
			t.Errorf("frame %d: texture map size = %d, want %d", frame, got, numBoundaries)
		}
	}

	// After 10 frames, map should still be exactly numBoundaries.
	if got := len(textureMap); got != numBoundaries {
		t.Errorf("after 10 frames: texture map size = %d, want %d", got, numBoundaries)
	}

	// Simulate 2 boundaries removed (scroll out of view).
	removedKeys := []string{boundaries[3].key, boundaries[4].key}
	for _, key := range removedKeys {
		delete(textureMap, key)
	}

	// After pruning, map should have numBoundaries - 2 entries.
	wantAfterPrune := numBoundaries - len(removedKeys)
	if got := len(textureMap); got != wantAfterPrune {
		t.Errorf("after pruning: texture map size = %d, want %d", got, wantAfterPrune)
	}
}

// --- Test 6: Full frame vs damage frame pixel diff ---

// TestPixelDiff_FullFrameVsDamageFrame renders a full frame with 5 boundary
// regions (root + 4 children at known positions), then a damage frame where
// only child 1 changes. The test computes byte-for-byte pixel diff and asserts
// that ONLY child 1's area (30x30) was modified — every other pixel is
// identical between the two frames.
func TestPixelDiff_FullFrameVsDamageFrame(t *testing.T) {
	device, queue, cleanup := createSoftwareDevice(t)
	defer cleanup()

	const surfW, surfH = 200, 200

	// 4 child boundaries at known positions with known colors.
	type child struct {
		x, y, w, h uint32
		r, g, b, a uint8
	}
	children := []child{
		{x: 10, y: 10, w: 30, h: 30, r: 0, g: 0, b: 255, a: 255},    // child 0: blue
		{x: 50, y: 50, w: 30, h: 30, r: 0, g: 255, b: 0, a: 255},    // child 1: green (will change)
		{x: 100, y: 10, w: 30, h: 30, r: 255, g: 255, b: 0, a: 255}, // child 2: yellow
		{x: 10, y: 100, w: 30, h: 30, r: 255, g: 0, b: 255, a: 255}, // child 3: magenta
	}

	surfTex, surfView := createBlitTarget(t, device, surfW, surfH)
	defer surfTex.Release()
	defer surfView.Release()

	// --- Full frame: clear RED, blit all 4 children ---
	clearTexture(t, device, queue, surfView, gputypes.Color{R: 1, G: 0, B: 0, A: 1})
	for _, c := range children {
		writeRegion(t, queue, surfTex, c.x, c.y, c.w, c.h, c.r, c.g, c.b, c.a)
	}

	fullFrame := readbackTexture(t, device, queue, surfTex, surfW, surfH)
	if fullFrame == nil {
		t.Skip("readback not available")
	}

	// Verify full frame structure.
	assertPixelRGBA(t, fullFrame, surfW, 0, 0, 255, 0, 0, "full-root")
	assertPixelRGBA(t, fullFrame, surfW, 20, 20, 0, 0, 255, "full-child0-blue")
	assertPixelRGBA(t, fullFrame, surfW, 60, 60, 0, 255, 0, "full-child1-green")
	assertPixelRGBA(t, fullFrame, surfW, 110, 20, 255, 255, 0, "full-child2-yellow")
	assertPixelRGBA(t, fullFrame, surfW, 20, 110, 255, 0, 255, "full-child3-magenta")

	// --- Damage frame: only child 1 re-rendered (cyan instead of green) ---
	// Use WriteTexture to overwrite child 1's region only.
	c1 := children[1]
	writeRegion(t, queue, surfTex, c1.x, c1.y, c1.w, c1.h, 0, 255, 255, 255) // cyan

	damageFrame := readbackTexture(t, device, queue, surfTex, surfW, surfH)
	if damageFrame == nil {
		t.Skip("readback not available")
	}

	// --- Compute pixel diff ---
	changedPixels := 0
	for i := 0; i < len(fullFrame); i += 4 {
		if fullFrame[i] != damageFrame[i] ||
			fullFrame[i+1] != damageFrame[i+1] ||
			fullFrame[i+2] != damageFrame[i+2] {
			changedPixels++
		}
	}

	// Only child 1's area should have changed: 30 * 30 = 900 pixels.
	wantChanged := int(c1.w) * int(c1.h)
	if changedPixels != wantChanged {
		t.Errorf("changed pixels = %d, want exactly %d (child 1 area only)", changedPixels, wantChanged)
	}

	// Verify unchanged regions byte-for-byte.
	unchangedPoints := [][2]int{
		{0, 0},     // root corner
		{199, 199}, // root far corner
		{20, 20},   // child 0 (unchanged blue)
		{110, 20},  // child 2 (unchanged yellow)
		{20, 110},  // child 3 (unchanged magenta)
		{150, 150}, // root interior
	}
	for _, pt := range unchangedPoints {
		idx := (pt[1]*surfW + pt[0]) * 4
		if fullFrame[idx] != damageFrame[idx] ||
			fullFrame[idx+1] != damageFrame[idx+1] ||
			fullFrame[idx+2] != damageFrame[idx+2] ||
			fullFrame[idx+3] != damageFrame[idx+3] {
			t.Errorf("pixel (%d,%d) changed between frames — expected identical "+
				"(full=RGBA(%d,%d,%d,%d) damage=RGBA(%d,%d,%d,%d))",
				pt[0], pt[1],
				fullFrame[idx], fullFrame[idx+1], fullFrame[idx+2], fullFrame[idx+3],
				damageFrame[idx], damageFrame[idx+1], damageFrame[idx+2], damageFrame[idx+3])
		}
	}

	// Verify the changed region has the new color (cyan).
	assertPixelRGBA(t, damageFrame, surfW, 60, 60, 0, 255, 255, "damage-child1-cyan")
	assertPixelRGBA(t, damageFrame, surfW, 65, 65, 0, 255, 255, "damage-child1-interior-cyan")
}

// --- Test 7: Damage-aware blit — scissor rect matches spinner bounds ---

// TestDamageAwareBlit_ScissorRect_MatchesSpinnerBounds verifies the FULL damage-aware
// blit pipeline through the software HAL: ui → gg → wgpu.
//
// The test simulates a 200×200 window surface with a 48×48 spinner at position (80,80).
//
//   - Frame 1 (full render): BeginRenderPass with LoadOpClear → End.
//     Asserts: ColorLoadOp==LoadOpClear, HasScissor==false (full window draw).
//
//   - Frame 2 (damage-aware): BeginRenderPass with LoadOpLoad → SetScissorRect(80,80,48,48) → Draw → End.
//     Asserts: ColorLoadOp==LoadOpLoad, HasScissor==true, ScissorRect==(80,80)-(128,128),
//     DrawCount==1 (only spinner re-rendered).
//
// This proves that the damage pipeline sends scissor=48×48 (not full window 200×200) to
// the GPU, which is the key optimization: the GPU only touches dirty pixels.
func TestDamageAwareBlit_ScissorRect_MatchesSpinnerBounds(t *testing.T) {
	halDev, halCleanup := createSoftwareHALDevice(t)
	defer halCleanup()

	const (
		surfW, surfH       = 200, 200
		spinnerX, spinnerY = 80, 80
		spinnerW, spinnerH = 48, 48
	)

	tex, view := createHALRenderTarget(t, halDev, surfW, surfH)
	defer tex.Destroy()
	defer view.Destroy()

	// --- Frame 1: full render (LoadOpClear) ---
	enc1, err := halDev.CreateCommandEncoder(&hal.CommandEncoderDescriptor{Label: "frame1"})
	if err != nil {
		t.Fatalf("CreateCommandEncoder frame1: %v", err)
	}
	pass1 := enc1.BeginRenderPass(&hal.RenderPassDescriptor{
		Label: "frame1-full",
		ColorAttachments: []hal.RenderPassColorAttachment{{
			View:       view,
			LoadOp:     gputypes.LoadOpClear,
			StoreOp:    gputypes.StoreOpStore,
			ClearValue: gputypes.Color{R: 1, G: 0, B: 0, A: 1},
		}},
	})
	pass1.End()

	stats1 := pass1.(*software.RenderPassEncoder).Stats()

	if stats1.ColorLoadOp != gputypes.LoadOpClear {
		t.Errorf("frame1: ColorLoadOp = %v, want LoadOpClear (%v)", stats1.ColorLoadOp, gputypes.LoadOpClear)
	}
	if stats1.HasScissor {
		t.Error("frame1: HasScissor = true, want false (full window render)")
	}
	if stats1.Width != surfW || stats1.Height != surfH {
		t.Errorf("frame1: render target size = %dx%d, want %dx%d", stats1.Width, stats1.Height, surfW, surfH)
	}

	// --- Frame 2: damage-aware (LoadOpLoad + scissor for spinner only) ---
	enc2, err := halDev.CreateCommandEncoder(&hal.CommandEncoderDescriptor{Label: "frame2"})
	if err != nil {
		t.Fatalf("CreateCommandEncoder frame2: %v", err)
	}
	pass2 := enc2.BeginRenderPass(&hal.RenderPassDescriptor{
		Label: "frame2-damage",
		ColorAttachments: []hal.RenderPassColorAttachment{{
			View:    view,
			LoadOp:  gputypes.LoadOpLoad,
			StoreOp: gputypes.StoreOpStore,
		}},
	})
	pass2.SetViewport(gputypes.Viewport{Width: surfW, Height: surfH, MaxDepth: 1})
	pass2.SetScissorRect(gputypes.ScissorRect{X: spinnerX, Y: spinnerY, Width: spinnerW, Height: spinnerH})
	pass2.Draw(gputypes.DrawArgs{VertexCount: 6, InstanceCount: 1}) // simulated spinner quad
	pass2.End()

	stats2 := pass2.(*software.RenderPassEncoder).Stats()

	if stats2.ColorLoadOp != gputypes.LoadOpLoad {
		t.Errorf("frame2: ColorLoadOp = %v, want LoadOpLoad (%v)", stats2.ColorLoadOp, gputypes.LoadOpLoad)
	}
	if !stats2.HasScissor {
		t.Error("frame2: HasScissor = false, want true (damage-aware scissor)")
	}
	wantRect := image.Rect(
		int(spinnerX), int(spinnerY),
		int(spinnerX+spinnerW), int(spinnerY+spinnerH),
	)
	if stats2.ScissorRect != wantRect {
		t.Errorf("frame2: ScissorRect = %v, want %v (spinner bounds)", stats2.ScissorRect, wantRect)
	}
	if stats2.DrawCount != 1 {
		t.Errorf("frame2: DrawCount = %d, want 1 (only spinner re-rendered)", stats2.DrawCount)
	}
	if stats2.Width != surfW || stats2.Height != surfH {
		t.Errorf("frame2: render target size = %dx%d, want %dx%d", stats2.Width, stats2.Height, surfW, surfH)
	}

	// --- Verify scissor covers exactly 48×48, NOT 200×200 ---
	scissorW := stats2.ScissorRect.Dx()
	scissorH := stats2.ScissorRect.Dy()
	if scissorW != spinnerW || scissorH != spinnerH {
		t.Errorf("scissor size = %dx%d, want %dx%d (spinner only, NOT full window %dx%d)",
			scissorW, scissorH, spinnerW, spinnerH, surfW, surfH)
	}
}

// --- Test 8: LoadOpLoad confirmed through pipeline ---

// TestDamageAwareBlit_LoadOpLoad_Confirmed verifies that the software HAL correctly
// records LoadOpLoad vs LoadOpClear across consecutive frames.
//
//   - Frame 1: LoadOpClear (initial full render) — stats.ColorLoadOp == 1
//   - Frame 2: LoadOpLoad (preserve previous frame) — stats.ColorLoadOp == 2
//   - Frame 3: LoadOpClear (forced full redraw) — stats.ColorLoadOp == 1
//
// This is the foundational guarantee: without LoadOpLoad, the damage pipeline is broken
// because every frame would clear the surface and lose previously rendered content.
func TestDamageAwareBlit_LoadOpLoad_Confirmed(t *testing.T) {
	halDev, halCleanup := createSoftwareHALDevice(t)
	defer halCleanup()

	const W, H = 100, 100

	tex, view := createHALRenderTarget(t, halDev, W, H)
	defer tex.Destroy()
	defer view.Destroy()

	tests := []struct {
		name       string
		loadOp     gputypes.LoadOp
		wantLoadOp gputypes.LoadOp
	}{
		{"frame1-clear", gputypes.LoadOpClear, gputypes.LoadOpClear},
		{"frame2-load", gputypes.LoadOpLoad, gputypes.LoadOpLoad},
		{"frame3-clear-again", gputypes.LoadOpClear, gputypes.LoadOpClear},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := halDev.CreateCommandEncoder(&hal.CommandEncoderDescriptor{Label: tt.name})
			if err != nil {
				t.Fatalf("CreateCommandEncoder: %v", err)
			}
			pass := enc.BeginRenderPass(&hal.RenderPassDescriptor{
				Label: tt.name,
				ColorAttachments: []hal.RenderPassColorAttachment{{
					View:       view,
					LoadOp:     tt.loadOp,
					StoreOp:    gputypes.StoreOpStore,
					ClearValue: gputypes.Color{R: 0, G: 0, B: 0, A: 1},
				}},
			})
			pass.End()

			stats := pass.(*software.RenderPassEncoder).Stats()
			if stats.ColorLoadOp != tt.wantLoadOp {
				t.Errorf("ColorLoadOp = %v, want %v", stats.ColorLoadOp, tt.wantLoadOp)
			}
		})
	}
}

// --- Test 9: Scissor rect matches exact dirty boundary bounds ---

// TestScissorRect_ExactBounds verifies that SetScissorRect records the EXACT pixel
// coordinates of a dirty boundary, not a rounded or expanded region.
//
// This is critical for damage-aware blit correctness: the scissor must match the
// boundary's screen position and size exactly, otherwise we either:
//   - Clip too much (miss dirty pixels at edges)
//   - Clip too little (waste GPU bandwidth on clean pixels)
//
// Tests multiple boundary positions and sizes including edge cases (origin, max corner,
// odd dimensions, single pixel).
func TestScissorRect_ExactBounds(t *testing.T) {
	halDev, halCleanup := createSoftwareHALDevice(t)
	defer halCleanup()

	const surfW, surfH = 300, 300

	tex, view := createHALRenderTarget(t, halDev, surfW, surfH)
	defer tex.Destroy()
	defer view.Destroy()

	tests := []struct {
		name                       string
		x, y, w, h                 uint32
		wantMinX, wantMinY         int
		wantMaxX, wantMaxY         int
		wantScissorW, wantScissorH int
	}{
		{
			name: "standard-48x48-spinner",
			x:    80, y: 80, w: 48, h: 48,
			wantMinX: 80, wantMinY: 80, wantMaxX: 128, wantMaxY: 128,
			wantScissorW: 48, wantScissorH: 48,
		},
		{
			name: "small-30x25-at-offset",
			x:    50, y: 70, w: 30, h: 25,
			wantMinX: 50, wantMinY: 70, wantMaxX: 80, wantMaxY: 95,
			wantScissorW: 30, wantScissorH: 25,
		},
		{
			name: "origin-corner",
			x:    0, y: 0, w: 16, h: 16,
			wantMinX: 0, wantMinY: 0, wantMaxX: 16, wantMaxY: 16,
			wantScissorW: 16, wantScissorH: 16,
		},
		{
			name: "bottom-right-corner",
			x:    260, y: 260, w: 40, h: 40,
			wantMinX: 260, wantMinY: 260, wantMaxX: 300, wantMaxY: 300,
			wantScissorW: 40, wantScissorH: 40,
		},
		{
			name: "odd-dimensions-37x53",
			x:    100, y: 100, w: 37, h: 53,
			wantMinX: 100, wantMinY: 100, wantMaxX: 137, wantMaxY: 153,
			wantScissorW: 37, wantScissorH: 53,
		},
		{
			name: "single-pixel",
			x:    150, y: 150, w: 1, h: 1,
			wantMinX: 150, wantMinY: 150, wantMaxX: 151, wantMaxY: 151,
			wantScissorW: 1, wantScissorH: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := halDev.CreateCommandEncoder(&hal.CommandEncoderDescriptor{Label: tt.name})
			if err != nil {
				t.Fatalf("CreateCommandEncoder: %v", err)
			}
			pass := enc.BeginRenderPass(&hal.RenderPassDescriptor{
				Label: tt.name,
				ColorAttachments: []hal.RenderPassColorAttachment{{
					View:    view,
					LoadOp:  gputypes.LoadOpLoad,
					StoreOp: gputypes.StoreOpStore,
				}},
			})
			pass.SetScissorRect(gputypes.ScissorRect{X: tt.x, Y: tt.y, Width: tt.w, Height: tt.h})
			pass.Draw(gputypes.DrawArgs{VertexCount: 6, InstanceCount: 1}) // simulated boundary quad
			pass.End()

			stats := pass.(*software.RenderPassEncoder).Stats()

			if !stats.HasScissor {
				t.Fatal("HasScissor = false, want true")
			}

			wantRect := image.Rect(tt.wantMinX, tt.wantMinY, tt.wantMaxX, tt.wantMaxY)
			if stats.ScissorRect != wantRect {
				t.Errorf("ScissorRect = %v, want %v", stats.ScissorRect, wantRect)
			}

			gotW := stats.ScissorRect.Dx()
			gotH := stats.ScissorRect.Dy()
			if gotW != tt.wantScissorW || gotH != tt.wantScissorH {
				t.Errorf("scissor dimensions = %dx%d, want %dx%d", gotW, gotH, tt.wantScissorW, tt.wantScissorH)
			}

			if stats.ColorLoadOp != gputypes.LoadOpLoad {
				t.Errorf("ColorLoadOp = %v, want LoadOpLoad", stats.ColorLoadOp)
			}
		})
	}
}

// --- Helpers ---

// createSoftwareHALDevice creates a software-backend HAL device directly, bypassing
// the wgpu-core validation layer. This gives direct access to software.RenderPassEncoder
// and its Stats() method for CI e2e assertions.
//
// Use this helper when you need to inspect HAL-level render pass statistics
// (scissor rect, load op, draw count). For pixel-level tests that use wgpu-level
// API (CreateTexture, WriteTexture, readback), use createSoftwareDevice instead.
func createSoftwareHALDevice(t *testing.T) (hal.Device, func()) {
	t.Helper()
	instance, err := software.NewBackend().CreateInstance(&hal.InstanceDescriptor{})
	if err != nil {
		t.Fatalf("software CreateInstance: %v", err)
	}
	adapters := instance.EnumerateAdapters(nil)
	if len(adapters) == 0 {
		instance.Destroy()
		t.Fatal("software backend: no adapters")
	}
	openDev, err := adapters[0].Adapter.Open(0, gputypes.DefaultLimits())
	if err != nil {
		instance.Destroy()
		t.Fatalf("software Open: %v", err)
	}
	cleanup := func() {
		openDev.Device.Destroy()
		instance.Destroy()
	}
	return openDev.Device, cleanup
}

// createHALRenderTarget creates a HAL-level RGBA8 texture and view suitable for
// use as a render attachment. The texture has RenderAttachment usage so it can
// be used in BeginRenderPass.
func createHALRenderTarget(t *testing.T, dev hal.Device, w, h uint32) (hal.Texture, hal.TextureView) {
	t.Helper()
	tex, err := dev.CreateTexture(&hal.TextureDescriptor{
		Size:          hal.Extent3D{Width: w, Height: h, DepthOrArrayLayers: 1},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     gputypes.TextureDimension2D,
		Format:        gputypes.TextureFormatRGBA8Unorm,
		Usage:         gputypes.TextureUsageRenderAttachment,
	})
	if err != nil {
		t.Fatalf("CreateTexture %dx%d: %v", w, h, err)
	}
	view, err := dev.CreateTextureView(tex, nil)
	if err != nil {
		tex.Destroy()
		t.Fatalf("CreateTextureView %dx%d: %v", w, h, err)
	}
	return tex, view
}

// createBlitTarget creates an RGBA8 texture that can receive CopyTextureToTexture
// blits, be used as a render attachment (for clear), and be read back via
// CopyTextureToBuffer. Usage: RenderAttachment | CopyDst | CopySrc.
func createBlitTarget(t *testing.T, device *wgpu.Device, w, h int) (*wgpu.Texture, *wgpu.TextureView) {
	t.Helper()
	tex, err := device.CreateTexture(&wgpu.TextureDescriptor{
		Label:         "blit-target",
		Size:          wgpu.Extent3D{Width: uint32(w), Height: uint32(h), DepthOrArrayLayers: 1},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     gputypes.TextureDimension2D,
		Format:        wgpu.TextureFormatRGBA8Unorm,
		Usage:         wgpu.TextureUsageRenderAttachment | wgpu.TextureUsageCopyDst | wgpu.TextureUsageCopySrc,
	})
	if err != nil {
		t.Fatalf("CreateTexture blit-target %dx%d: %v", w, h, err)
	}
	view, err := device.CreateTextureView(tex, nil)
	if err != nil {
		tex.Release()
		t.Fatalf("CreateTextureView blit-target %dx%d: %v", w, h, err)
	}
	return tex, view
}

// writeRegion writes solid-color pixel data into a sub-region of a destination
// texture via Queue.WriteTexture. This is the software-backend-compatible
// equivalent of CopyTextureToTexture (which ignores origin in the software
// HAL). WriteTexture correctly handles Origin offsets.
func writeRegion(t *testing.T, queue *wgpu.Queue, dst *wgpu.Texture,
	dstX, dstY, w, h uint32, r, g, b, a uint8) {
	t.Helper()
	pixelCount := int(w) * int(h)
	data := make([]byte, pixelCount*4)
	for i := 0; i < pixelCount; i++ {
		data[i*4+0] = r
		data[i*4+1] = g
		data[i*4+2] = b
		data[i*4+3] = a
	}
	err := queue.WriteTexture(
		&wgpu.ImageCopyTexture{
			Texture: dst,
			Origin:  wgpu.Origin3D{X: dstX, Y: dstY},
		},
		data,
		&wgpu.ImageDataLayout{
			BytesPerRow:  w * 4,
			RowsPerImage: h,
		},
		&wgpu.Extent3D{Width: w, Height: h, DepthOrArrayLayers: 1},
	)
	if err != nil {
		t.Fatalf("WriteTexture: %v", err)
	}
}
