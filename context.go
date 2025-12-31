package aeno

import (
	"image"
	"math"
	"runtime"
	"sync"
)

type Face int

const (
	_ Face = iota
	FaceCW
	FaceCCW
)

type Cull int

const (
	_ Cull = iota
	CullNone
	CullFront
	CullBack
)

type Context struct {
	Width        int
	Height       int
	Shader       Shader
	ColorBuffer  *image.NRGBA
	DepthBuffer  []float64
	ClearColor   Color
	ReadDepth    bool
	WriteDepth   bool
	WriteColor   bool
	AlphaBlend   bool
	Wireframe    bool
	FrontFace    Face
	Cull         Cull
	LineWidth    float64
	DepthBias    float64
	screenMatrix Matrix
	locks        []sync.Mutex
}

func NewContext(width, height int, shader Shader) *Context {
	dc := &Context{}
	dc.Width = width
	dc.Height = height
	dc.Shader = shader
	dc.ColorBuffer = image.NewNRGBA(image.Rect(0, 0, width, height))
	dc.DepthBuffer = make([]float64, width*height)
	dc.ClearColor = Transparent
	dc.ReadDepth = true
	dc.WriteDepth = true
	dc.WriteColor = true
	dc.AlphaBlend = true
	dc.Wireframe = false
	dc.FrontFace = FaceCCW
	dc.Cull = CullBack
	dc.LineWidth = 2
	dc.DepthBias = 0
	dc.screenMatrix = Screen(width, height)
	dc.locks = make([]sync.Mutex, 256)
	dc.ClearDepthBuffer()
	return dc
}

func (dc *Context) Image() image.Image {
	return dc.ColorBuffer
}

// ClearColorBufferWith uses fast memory copy to clear the buffer
func (dc *Context) ClearColorBufferWith(c Color) {
	nrgba := c.NRGBA()
	// Create a single row with the color
	row := make([]uint8, dc.Width*4)
	for x := 0; x < dc.Width; x++ {
		i := x * 4
		row[i+0] = nrgba.R
		row[i+1] = nrgba.G
		row[i+2] = nrgba.B
		row[i+3] = nrgba.A
	}
	// Copy row to all rows
	pix := dc.ColorBuffer.Pix
	stride := dc.ColorBuffer.Stride
	for y := 0; y < dc.Height; y++ {
		copy(pix[y*stride:], row)
	}
}

func (dc *Context) ClearColorBuffer() {
	dc.ClearColorBufferWith(dc.ClearColor)
}

func (dc *Context) ClearDepthBuffer() {
	for i := range dc.DepthBuffer {
		dc.DepthBuffer[i] = math.MaxFloat64
	}
}

func edge(a, b, c Vector) float64 {
	return (b.X-c.X)*(a.Y-c.Y) - (b.Y-c.Y)*(a.X-c.X)
}

func (dc *Context) rasterize(v0, v1, v2 Vertex, s0, s1, s2 Vector, fromObject *Object) {
	min := s0.Min(s1.Min(s2)).Floor()
	max := s0.Max(s1.Max(s2)).Ceil()

	x0 := int(min.X)
	x1 := int(max.X)
	y0 := int(min.Y)
	y1 := int(max.Y)

	// Clip to screen bounds
	x0 = ClampInt(x0, 0, dc.Width-1)
	x1 = ClampInt(x1, 0, dc.Width-1)
	y0 = ClampInt(y0, 0, dc.Height-1)
	y1 = ClampInt(y1, 0, dc.Height-1)

	p := Vector{float64(x0) + 0.5, float64(y0) + 0.5, 0}
	w00 := edge(s1, s2, p)
	w01 := edge(s2, s0, p)
	w02 := edge(s0, s1, p)
	a01 := s1.Y - s0.Y
	b01 := s0.X - s1.X
	a12 := s2.Y - s1.Y
	b12 := s1.X - s2.X
	a20 := s0.Y - s2.Y
	b20 := s2.X - s0.X

	ra := 1 / edge(s0, s1, s2)
	r0 := 1 / v0.Output.W
	r1 := 1 / v1.Output.W
	r2 := 1 / v2.Output.W

	// Constant loop variables
	stride := dc.Width
	pix := dc.ColorBuffer.Pix

	for y := y0; y <= y1; y++ {
		w0 := w00
		w1 := w01
		w2 := w02
		for x := x0; x <= x1; x++ {
			b0 := w0 * ra
			b1 := w1 * ra
			b2 := w2 * ra

			// Check if inside triangle
			if b0 >= 0 && b1 >= 0 && b2 >= 0 {
				i := y*stride + x
				z := b0*s0.Z + b1*s1.Z + b2*s2.Z
				bz := z + dc.DepthBias

				// Early depth test
				if !dc.ReadDepth || bz <= dc.DepthBuffer[i] {
					
					// Interpolate
					b := VectorW{b0 * r0, b1 * r1, b2 * r2, 0}
					b.W = 1 / (b.X + b.Y + b.Z)
					v := InterpolateVertexes(v0, v1, v2, b)

					colorVal := dc.Shader.Fragment(v, fromObject)

					if colorVal.A > 0 {
						// Critical Section
						lock := &dc.locks[(x+y)&255]
						lock.Lock()

						if !dc.ReadDepth || bz <= dc.DepthBuffer[i] {
							if dc.WriteDepth {
								dc.DepthBuffer[i] = z
							}
							if dc.WriteColor {
								dc.setPixel(x, y, colorVal, pix, i*4)
							}
						}
						lock.Unlock()
					}
				}
			}
			w0 += a12
			w1 += a20
			w2 += a01
		}
		w00 += b12
		w01 += b20
		w02 += b01
	}
}

// Inlined pixel setting for speed
func (dc *Context) setPixel(x, y int, c Color, pix []uint8, i int) {
	if dc.AlphaBlend && c.A < 1 {
		sr, sg, sb, sa := c.NRGBA().RGBA()
		a := (0xffff - sa) * 0x101
		
		dr := uint32(pix[i+0])
		dg := uint32(pix[i+1])
		db := uint32(pix[i+2])
		da := uint32(pix[i+3])

		pix[i+0] = uint8((dr*a/0xffff + sr) >> 8)
		pix[i+1] = uint8((dg*a/0xffff + sg) >> 8)
		pix[i+2] = uint8((db*a/0xffff + sb) >> 8)
		pix[i+3] = uint8((da*a/0xffff + sa) >> 8)
	} else {
		// Fast path opaque
		nrgba := c.NRGBA()
		pix[i+0] = nrgba.R
		pix[i+1] = nrgba.G
		pix[i+2] = nrgba.B
		pix[i+3] = nrgba.A
	}
}

func (dc *Context) line(v0, v1 Vertex, s0, s1 Vector, fromObject *Object) {
	n := s1.Sub(s0).Perpendicular().MulScalar(dc.LineWidth / 2)
	s0 = s0.Add(s0.Sub(s1).Normalize().MulScalar(dc.LineWidth / 2))
	s1 = s1.Add(s1.Sub(s0).Normalize().MulScalar(dc.LineWidth / 2))
	s00 := s0.Add(n)
	s01 := s0.Sub(n)
	s10 := s1.Add(n)
	s11 := s1.Sub(n)
	dc.rasterize(v1, v0, v0, s11, s01, s00, fromObject)
	dc.rasterize(v1, v1, v0, s10, s11, s00, fromObject)
}

func (dc *Context) drawClippedLine(v0, v1 Vertex, fromObject *Object) {
	ndc0 := v0.Output.DivScalar(v0.Output.W).Vector()
	ndc1 := v1.Output.DivScalar(v1.Output.W).Vector()
	s0 := dc.screenMatrix.MulPosition(ndc0)
	s1 := dc.screenMatrix.MulPosition(ndc1)
	dc.line(v0, v1, s0, s1, fromObject)
}

func (dc *Context) drawClippedTriangle(v0, v1, v2 Vertex, fromObject *Object) {
	ndc0 := v0.Output.DivScalar(v0.Output.W).Vector()
	ndc1 := v1.Output.DivScalar(v1.Output.W).Vector()
	ndc2 := v2.Output.DivScalar(v2.Output.W).Vector()

	if dc.Cull != CullNone {
		area := (ndc1.X-ndc0.X)*(ndc2.Y-ndc0.Y) - (ndc2.X-ndc0.X)*(ndc1.Y-ndc0.Y)
		if dc.FrontFace == FaceCW {
			area = -area
		}
		if dc.Cull == CullBack && area <= 0 {
			return
		}
		if dc.Cull == CullFront && area >= 0 {
			return
		}
	}

	s0 := dc.screenMatrix.MulPosition(ndc0)
	s1 := dc.screenMatrix.MulPosition(ndc1)
	s2 := dc.screenMatrix.MulPosition(ndc2)

	if dc.Wireframe {
		dc.wireframe(v0, v1, v2, s0, s1, s2, fromObject)
		return 
	}
	dc.rasterize(v0, v1, v2, s0, s1, s2, fromObject)
}

func (dc *Context) wireframe(v0, v1, v2 Vertex, s0, s1, s2 Vector, fromObject *Object) {
	dc.line(v0, v1, s0, s1, fromObject)
	dc.line(v1, v2, s1, s2, fromObject)
	dc.line(v2, v0, s2, s0, fromObject)
}

func (dc *Context) DrawTriangle(t *Triangle, fromObject *Object) {
	v1 := dc.Shader.Vertex(t.V1)
	v2 := dc.Shader.Vertex(t.V2)
	v3 := dc.Shader.Vertex(t.V3)

	if v1.Outside() || v2.Outside() || v3.Outside() {
		triangles := ClipTriangle(NewTriangle(v1, v2, v3))
		for _, t := range triangles {
			dc.drawClippedTriangle(t.V1, t.V2, t.V3, fromObject)
		}
	} else {
		dc.drawClippedTriangle(v1, v2, v3, fromObject)
	}
}

func (dc *Context) DrawLine(l *Line, fromObject *Object) {
	v1 := dc.Shader.Vertex(l.V1)
	v2 := dc.Shader.Vertex(l.V2)

	if v1.Outside() || v2.Outside() {
		line := ClipLine(NewLine(v1, v2))
		if line != nil {
			dc.drawClippedLine(line.V1, line.V2, fromObject)
		}
	} else {
		dc.drawClippedLine(v1, v2, fromObject)
	}
}

func (dc *Context) DrawMesh(mesh *Mesh, fromObject *Object) {
	var wg sync.WaitGroup
	// Use logical CPUs
	wn := runtime.NumCPU()
	wg.Add(wn)
	
	// Batch processing for less goroutine overhead
	for wi := 0; wi < wn; wi++ {
		go func(wi int) {
			for i := wi; i < len(mesh.Triangles); i += wn {
				dc.DrawTriangle(mesh.Triangles[i], fromObject)
			}
			for i := wi; i < len(mesh.Lines); i += wn {
				dc.DrawLine(mesh.Lines[i], fromObject)
			}
			wg.Done()
		}(wi)
	}
	wg.Wait()
}

func (dc *Context) DrawObject(o *Object) {
	if s, ok := dc.Shader.(*PhongShader); ok {
		prev := s.Matrix
		s.Matrix = s.Matrix.Mul(o.Matrix)
		dc.DrawMesh(o.Mesh, o)
		s.Matrix = prev
	} else if s, ok := dc.Shader.(*ToonShader); ok {
		prev := s.Matrix
		s.Matrix = s.Matrix.Mul(o.Matrix)
		dc.DrawMesh(o.Mesh, o)
		s.Matrix = prev
	} else {
		dc.DrawMesh(o.Mesh, o)
	}
}