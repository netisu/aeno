
package aeno

import "math"

// ToonShader implements cel shading with optional outlining.
type ToonShader struct {
	Matrix         Matrix
	LightDirection Vector
	CameraPosition Vector
	ColorSteps     map[float64]Color
	EnableOutline  bool    
	OutlineColor   Color   
	OutlineFactor  float64 
}

func NewToonShader(matrix Matrix, lightDir, cameraPosition Vector) *ToonShader {
	return &ToonShader{
		Matrix:         matrix,
		LightDirection: lightDir.Normalize(),
		CameraPosition: cameraPosition,
		ColorSteps: map[float64]Color{
			// The key is the brightness threshold (dot product)
			0.95: HexColor("ffffff"), // Highlight
			0.7: HexColor("cccccc"), // Mid-tone
			0.4: HexColor("888888"), // Shadow
			0.0: HexColor("444444"), // Deep Shadow
		},
		EnableOutline: false, // Off by default
		OutlineColor:  HexColor("000000"),
		OutlineFactor: 0.05,
	}
}

func (s *ToonShader) Vertex(v Vertex) Vertex {
	v.Output = s.Matrix.MulPositionW(v.Position)
	normalMatrix := s.Matrix.Inverse().Transpose()
	v.Normal = normalMatrix.MulDirection(v.Normal).Normalize()
	return v
}

func (s *ToonShader) Fragment(v Vertex, fromObject *Object) Color {
	if s.EnableOutline {
		viewDirection := s.CameraPosition.Sub(v.Position).Normalize()
		dot := viewDirection.Dot(v.Normal)
		if math.Abs(dot) < s.OutlineFactor {
			return s.OutlineColor
		}
	}
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
