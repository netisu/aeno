package aeno

import "github.com/go-gl/mathgl/mgl64"

type SolidColorShader struct {
	Matrix Matrix // View-Projection matrix
	Color  Color
}

func (s *SolidColorShader) GetMatrix() Matrix {
	return s.Matrix
}

func (s *SolidColorShader) Vertex(v Vertex, modelMatrix mgl64.Mat4) Vertex {
	mvp := s.Matrix.Mul(modelMatrix)
	v.Output = mvp.MulPositionW(v.Position)
	return v
}

func (s *SolidColorShader) Fragment(v Vertex, fromObject *Object) Color {
	return s.Color
}
