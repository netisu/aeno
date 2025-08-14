
package aeno

import "math"

// ToonShader implements cel shading.
type ToonShader struct {
	Matrix         Matrix
	LightDirection Vector
	// ColorSteps defines the brightness thresholds and corresponding colors.
	// 0.8 might map to a bright color, 0.4 to a mid-tone, etc.
	ColorSteps map[float64]Color 
}

func NewToonShader(matrix Matrix, lightDir Vector) *ToonShader {
	return &ToonShader{
		Matrix:         matrix,
		LightDirection: lightDir.Normalize(),
		ColorSteps: map[float64]Color{
			// The key is the brightness threshold (dot product)
			0.8: HexColor("ffffaa"), // Highlight
			0.5: HexColor("ff8844"), // Mid-tone
			0.2: HexColor("a12c00"), // Shadow
			0.0: HexColor("4d1100"), // Deep Shadow
		},
	}
}

func (s *ToonShader) Vertex(v Vertex) Vertex {
	v.Output = s.Matrix.MulPositionW(v.Position)
	normalMatrix := s.Matrix.Inverse().Transpose()
	v.Normal = s.Matrix.MulDirection(v.Normal).Normalize()
	return v
}

func (s *ToonShader) Fragment(v Vertex, fromObject *Object) Color {
  intensity := math.Max(0, v.Normal.Dot(s.LightDirection))
	// Determine the final color by snapping to the nearest step
	var finalColor Color
	if intensity > 0.8 {
		finalColor = s.ColorSteps[0.8]
	} else if intensity > 0.5 {
		finalColor = s.ColorSteps[0.5]
	} else if intensity > 0.2 {
		finalColor = s.ColorSteps[0.2]
	} else {
		finalColor = s.ColorSteps[0.0]
	}
	
	// You can still use the object's base color or texture
	if fromObject.Texture != nil {
		texColor := fromObject.Texture.Sample(v.Texture.X, v.Texture.Y)
		return texColor.Mul(finalColor) // Blend texture with toon lighting
	}

	return fromObject.Color.Mul(finalColor)
}
