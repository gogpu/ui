package desktop

import (
	"image"
	"math"
	"testing"
)

// TestPhysicalDamageRect_FloorCeilAtFractionalScale pins that the GPU-scissor
// damage rect rounds identically to gg's own Context.trackDamage -- Floor on
// the min corner, Ceil on the max corner -- rather than truncating the min
// corner and round-half-up-ing the SIZE, which is a different function and can
// under-cover by one physical pixel at fractional device scales.
//
// Under LoadOpLoad that missing pixel keeps the previous frame's content, a
// real stale seam -- not merely a debug-visualization glitch, since this feeds
// the GPU scissor, not the debug overlay.
//
// Chosen numbers discriminate the two formulas: logical origin (11,5), size
// 20x10, scale 1.5.
//
//	Floor/Ceil (gg, correct):               [Floor(16.5),Floor(7.5)]..[Ceil(46.5),Ceil(22.5)] = (16,7)-(47,23)
//	truncate+round-half-up-size (pre-fix):  (16,7)-(16+int(30.5),7+int(15.5)) = (16,7)-(46,22)
//
// Integer scales are unaffected -- both formulas agree when scale is whole,
// which is why the Retina (2x) hardware session that found the debug-overlay
// findings never surfaced this.
//
// (Finding 2 in issue #195.)
func TestPhysicalDamageRect_FloorCeilAtFractionalScale(t *testing.T) {
	tests := []struct {
		name  string
		rx    int     // logical origin X (math.Round of float32 ScreenOrigin.X)
		ry    int     // logical origin Y
		bw    int     // logical boundary width
		bh    int     // logical boundary height
		scale float64 // device scale factor
		want  image.Rectangle
	}{
		// 1x: non-discriminating (both formulas agree). Guards against
		// over-correcting.
		{
			name: "1x identity",
			rx:   11, ry: 5, bw: 20, bh: 10,
			scale: 1.0,
			want:  image.Rect(11, 5, 31, 15),
		},
		// 2x: non-discriminating (integer scale, both agree).
		{
			name: "2x retina",
			rx:   11, ry: 5, bw: 20, bh: 10,
			scale: 2.0,
			want:  image.Rect(22, 10, 62, 30),
		},
		// 1.5x: DISCRIMINATING. Pre-fix gives (16,7)-(46,22). Post-fix
		// gives (16,7)-(47,23). The Max corner differs by 1px on each axis.
		{
			name: "1.5x fractional - discriminating",
			rx:   11, ry: 5, bw: 20, bh: 10,
			scale: 1.5,
			want:  image.Rect(16, 7, 47, 23),
		},
		// 1.25x: another fractional scale common on Windows.
		{
			name: "1.25x Windows",
			rx:   10, ry: 10, bw: 100, bh: 50,
			scale: 1.25,
			want: image.Rect(
				int(math.Floor(10*1.25)),
				int(math.Floor(10*1.25)),
				int(math.Ceil(110*1.25)),
				int(math.Ceil(60*1.25)),
			),
		},
		// 1.75x: discriminating fractional scale.
		{
			name: "1.75x fractional",
			rx:   7, ry: 3, bw: 30, bh: 20,
			scale: 1.75,
			want: image.Rect(
				int(math.Floor(7*1.75)),
				int(math.Floor(3*1.75)),
				int(math.Ceil(37*1.75)),
				int(math.Ceil(23*1.75)),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := physicalDamageRect(tt.rx, tt.ry, tt.bw, tt.bh, tt.scale)
			if got != tt.want {
				t.Errorf("physicalDamageRect(%d, %d, %d, %d, %g) = %v, want %v"+
					"\n  (a Max.X/Max.Y one pixel short means the truncate+round-half-up-size formula is back)",
					tt.rx, tt.ry, tt.bw, tt.bh, tt.scale, got, tt.want)
			}
		})
	}
}
