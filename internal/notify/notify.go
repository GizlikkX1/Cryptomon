// Package notify sends native desktop notifications on Linux (D-Bus /
// notify-send), macOS (Notification Center) and Windows (toast), and
// renders alert texts in the user's language.
package notify

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"

	"github.com/gen2brain/beeep"
)

type Notifier struct {
	icon string
}

// New prepares the notification icon inside dataDir.
func New(dataDir string) *Notifier {
	beeep.AppName = "Cryptomon"
	n := &Notifier{}
	path := filepath.Join(dataDir, "icon.png")
	if err := writeIcon(path); err == nil {
		n.icon = path
	}
	return n
}

func (n *Notifier) Send(title, body string) error {
	return beeep.Notify(title, body, n.icon)
}

// writeIcon draws the app icon (a rising chart in a circle) as a PNG.
func writeIcon(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	const size = 128
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	c := float64(size) / 2
	top := color.NRGBA{0x7c, 0x5c, 0xff, 0xff}
	bottom := color.NRGBA{0x00, 0xb8, 0x94, 0xff}
	line := []struct{ x, y float64 }{{24, 88}, {48, 64}, {66, 76}, {100, 38}}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx, dy := float64(x)+0.5-c, float64(y)+0.5-c
			d := math.Hypot(dx, dy)
			if d > c {
				continue
			}
			t := float64(y) / size
			px := color.NRGBA{
				R: uint8(float64(top.R)*(1-t) + float64(bottom.R)*t),
				G: uint8(float64(top.G)*(1-t) + float64(bottom.G)*t),
				B: uint8(float64(top.B)*(1-t) + float64(bottom.B)*t),
				A: uint8(255 * math.Min(1, c-d)),
			}
			for i := 0; i+1 < len(line); i++ {
				if segDist(float64(x), float64(y), line[i].x, line[i].y, line[i+1].x, line[i+1].y) < 5 {
					px.R, px.G, px.B = 255, 255, 255
				}
			}
			img.SetNRGBA(x, y, px)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func segDist(px, py, ax, ay, bx, by float64) float64 {
	vx, vy := bx-ax, by-ay
	t := ((px-ax)*vx + (py-ay)*vy) / (vx*vx + vy*vy)
	t = math.Max(0, math.Min(1, t))
	return math.Hypot(px-(ax+t*vx), py-(ay+t*vy))
}
