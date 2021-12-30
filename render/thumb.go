// Copyright 2020 Daniel Erat <dan@erat.org>.
// All rights reserved.

package render

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

// genThumb reads an image from p, scales it down to the supplied dimensions,
// and returns base64-encoded GIF data.
func genThumb(p string, width, height int) (string, error) {
	// Require thumbnails to be small enough that we know that their colors will fit in the palette.
	// We don't do anything to choose colors intelligently while quantizing.
	if width*height > 256 {
		return "", errors.New("only 256 or fewer pixels supported")
	}

	f, err := os.Open(p)
	if err != nil {
		return "", err
	}
	defer f.Close()

	si, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}

	di := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(di, di.Bounds(), si, si.Bounds(), draw.Src, nil)

	var b bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &b)
	if err := gif.Encode(enc, di, &gif.Options{
		NumColors: width * height,
		Quantizer: &quantizer{},
	}); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}
	return b.String(), nil
}

// quantizer implements the draw.Quantizer interface (poorly).
// It just appends the image's unique colors to the palette until the palette is full.
type quantizer struct{}

func (q *quantizer) Quantize(p color.Palette, m image.Image) color.Palette {
	b := m.Bounds()
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			c := m.At(x, y)

			// Check if the color is already in the palette.
			if len(p) > 0 {
				r, g, b, _ := c.RGBA()
				pc := p.Convert(c)
				pr, pg, pb, _ := pc.RGBA()
				if r == pr && g == pg && b == pb {
					continue
				}
			}

			// Add the color to the palette.
			p = append(p, c)
			if len(p) == cap(p) { // palette is full
				return p
			}
		}
	}
	return p
}
