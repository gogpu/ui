// Package viewport3d provides a GPU-rendered viewport widget for 3D content,
// video playback, or custom shader output.
//
// Viewport3D uses the Flutter Texture widget pattern: the widget owns an
// offscreen GPU texture, an external renderer (e.g., g3d.Renderer, a video
// decoder, a compute shader) draws into it, and the Layer Tree compositor
// blits it to the surface via [compositor.ExternalTextureLayer].
//
// # Usage
//
// The widget does NOT import any specific 3D engine or renderer. Instead,
// the [OnRender] functional option takes a callback that receives the
// texture view, device, and queue as opaque handles. The consumer type-asserts
// to concrete types (e.g., *wgpu.Device, *wgpu.TextureView) as needed.
//
//	vp := viewport3d.New(
//	    viewport3d.Size(640, 480),
//	    viewport3d.OnRender(func(tv gpucontext.TextureView) {
//	        // Render 3D scene, video frame, or shader output into tv.
//	    }),
//	)
//
// # Continuous vs On-Demand Rendering
//
// By default, the viewport renders on-demand (only when explicitly
// invalidated via [Widget.Invalidate]). Use [Continuous](true) for
// content that changes every frame (animations, real-time 3D).
//
// # Enterprise References
//
//   - Flutter: TextureLayer (rendering/layer.dart), Texture widget (widgets/texture.dart)
//   - Android: TextureView (android.view.TextureView), SurfaceTexture
//   - Chrome: TextureLayer (cc/layers/texture_layer.h)
package viewport3d
