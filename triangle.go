package aeno

// Triangle f
type Triangle struct {
	V1, V2, V3 Vertex
}

// NewTriangle f
func NewTriangle(v1, v2, v3 Vertex) *Triangle {
	t := Triangle{v1, v2, v3}
	t.FixNormals()
	return &t
}

// NewTriangleForPoints f
func NewTriangleForPoints(p1, p2, p3 Vector) *Triangle {
	v1 := Vertex{Position: p1}
	v2 := Vertex{Position: p2}
	v3 := Vertex{Position: p3}
	return NewTriangle(v1, v2, v3)
}

// IsDegenerate f
func (t *Triangle) IsDegenerate() bool {
	p1 := t.V1.Position
	p2 := t.V2.Position
	p3 := t.V3.Position
	if p1 == p2 || p1 == p3 || p2 == p3 {
		return true
	}
	if p1.IsDegenerate() || p2.IsDegenerate() || p3.IsDegenerate() {
		return true
	}
	return false
}

// Normal f
func (t *Triangle) Normal() Vector {
	e1 := t.V2.Position.Sub(t.V1.Position)
	e2 := t.V3.Position.Sub(t.V1.Position)
	return e1.Cross(e2).Normalize()
}

// Area f
func (t *Triangle) Area() float64 {
	e1 := t.V2.Position.Sub(t.V1.Position)
	e2 := t.V3.Position.Sub(t.V1.Position)
	n := e1.Cross(e2)
	return n.Length() / 2
}

// FixNormals f
func (t *Triangle) FixNormals() {
	n := t.Normal()
	zero := Vector{}
	if t.V1.Normal == zero {
		t.V1.Normal = n
	}
	if t.V2.Normal == zero {
		t.V2.Normal = n
	}
	if t.V3.Normal == zero {
		t.V3.Normal = n
	}
}

// BoundingBox f
func (t *Triangle) BoundingBox() Box {
	min := t.V1.Position.Min(t.V2.Position).Min(t.V3.Position)
	max := t.V1.Position.Max(t.V2.Position).Max(t.V3.Position)
	return Box{min, max}
}

// Transform f
func (t *Triangle) Transform(matrix Matrix) {
	t.V1.Position = matrix.MulPosition(t.V1.Position)
	t.V2.Position = matrix.MulPosition(t.V2.Position)
	t.V3.Position = matrix.MulPosition(t.V3.Position)
	t.V1.Normal = matrix.MulDirection(t.V1.Normal)
	t.V2.Normal = matrix.MulDirection(t.V2.Normal)
	t.V3.Normal = matrix.MulDirection(t.V3.Normal)
}

// ReverseWinding f
func (t *Triangle) ReverseWinding() {
	t.V1, t.V3 = t.V3, t.V1
	t.V1.Normal = t.V1.Normal.Negate()
	t.V2.Normal = t.V2.Normal.Negate()
	t.V3.Normal = t.V3.Normal.Negate()
}

// SetColor set the color of the vertexes
func (t *Triangle) SetColor(c Color) {
	t.V1.Color = c
	t.V2.Color = c
	t.V3.Color = c
}
