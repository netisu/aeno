package aeno

// SolidColorShader is a simple shader that renders everything in one color.
type SolidColorShader struct {
	Matrix Matrix
	Color  Color
}

func NewSolidColorShader(matrix Matrix, color Color) *SolidColorShader {
	return &SolidColorShader{matrix, color}
}

func (s *SolidColorShader) Vertex(v Vertex) Vertex {
	v.Output = s.Matrix.MulPositionW(v.Position)
	return v
}

func (s *SolidColorShader) Fragment(v Vertex, fromObject *Object) Color {
	return s.Color
}
