package aeno

// SolidColorShader is a simple shader that renders everything in one color.
type SolidColorShader struct {
	Matrix Matrix
	Color  Color
	Thickness float64
}

func NewSolidColorShader(matrix Matrix, color Color) *SolidColorShader {
	return &SolidColorShader{matrix, color, thickness}
}

func (s *SolidColorShader) Vertex(v Vertex) Vertex {
	extrudedPosition := v.Position.Add(v.Normal.MulScalar(s.Thickness))
	v.Output = s.Matrix.MulPositionW(extrudedPosition)
	return v
}

func (s *SolidColorShader) Fragment(v Vertex, fromObject *Object) Color {
	return s.Color
}
