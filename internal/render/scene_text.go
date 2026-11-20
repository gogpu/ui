package render

import (
	"math"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// SetTextRasterizer sets the optional platform renderer for this recording.
func (c *SceneCanvas) SetTextRasterizer(r widget.TextRasterizer) {
	c.textRasterizer = r
}

func (c *SceneCanvas) measureRasterText(s string, style widget.TextStyle) (widget.RasterTextMetrics, bool) {
	if c.textRasterizer == nil || s == "" {
		return widget.RasterTextMetrics{}, false
	}
	return c.textRasterizer.Measure(s, style, c.DeviceScale())
}

func (c *SceneCanvas) drawRasterText(s string, bounds geometry.Rect, style widget.TextStyle) bool {
	if c.textRasterizer == nil {
		return false
	}
	if style.Color.A == 0 {
		return true
	}
	scale := c.DeviceScale()
	bitmap, ok := c.textRasterizer.Rasterize(s, style, scale)
	if !ok || bitmap.Image == nil {
		return false
	}

	img := bitmap.Image
	x := bounds.Min.X*scale + max(0, bounds.Width()*scale-float32(bitmap.Advance))*float32(style.Align.Float64())
	y := bounds.Center().Y*scale - float32(img.Rect.Dy())/2
	c.sc.DrawImage(&scene.Image{Width: img.Rect.Dx(), Height: img.Rect.Dy(), Data: img.Pix}, scene.Affine{
		A: 1 / scale, E: 1 / scale,
		C: (float32(math.Round(float64(x))) - float32(bitmap.InkPadding)) / scale,
		F: float32(math.Round(float64(y))) / scale,
	})
	return true
}
