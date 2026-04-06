package aeno

import (
	"bytes"
	"image"
	"net/http"
	"time"
	_ "image/jpeg" // Ensure decoders are present
	_ "image/png"
	"math"
)

type Texture interface {
	Sample(u, v float64) Color
	BilinearSample(u, v float64) Color
}

type ImageTexture struct {
	Width  int
	Height int
	Image  image.Image
}

func NewImageTexture(im image.Image) Texture {
	return &ImageTexture{
		Width:  im.Bounds().Dx(),
		Height: im.Bounds().Dy(),
		Image:  im,
	}
}

func LoadTexture(path string) (Texture, error) {
	im, err := LoadImage(path)
	if err != nil {
		return nil, err
	}
	return NewImageTexture(im), nil
}

func LoadTextureFromURL(url string) Texture {
	client := http.Client{
		Timeout: 10 * time.Second, // Prevent hanging
	}
	resp, err := client.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()
	
	im, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil
	}
	return NewImageTexture(im)
}

func TexFromBytes(data []byte) Texture {
	im, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	return NewImageTexture(im)
}

func (t *ImageTexture) Sample(u, v float64) Color {
	// Wrap coords
	u = u - math.Floor(u)
	v = v - math.Floor(v)
	
	x := int(u * float64(t.Width))
	y := int(v * float64(t.Height))
	
	// Bounds check
	if x >= t.Width { x = t.Width - 1 }
	if y >= t.Height { y = t.Height - 1 }
	
	return MakeColor(t.Image.At(x, y))
}

func (t *ImageTexture) BilinearSample(u, v float64) Color {
	u = math.Max(0, math.Min(1.0-1e-9, u))
    v = math.Max(0, math.Min(1.0-1e-9, v))

   	uScaled := u * float64(t.Width-1)
    vScaled := v * float64(t.Height-1)

    x0, y0 := int(uScaled), int(vScaled)
    x1, y1 := x0+1, y0+1

	if x1 >= t.Width { x1 = x0 }
    if y1 >= t.Height { y1 = y0 }

    uFrac := uScaled - float64(x0)
    vFrac := vScaled - float64(y0)

    c00 := MakeColor(t.Image.At(x0, y0))
    c10 := MakeColor(t.Image.At(x1, y0))
    c01 := MakeColor(t.Image.At(x0, y1))
    c11 := MakeColor(t.Image.At(x1, y1))

    top := c00.Lerp(c10, uFrac)
    bottom := c01.Lerp(c11, uFrac)

    return top.Lerp(bottom, vFrac)
}
