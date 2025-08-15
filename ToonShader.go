package aeno

import "math"

// ToonShader implements cel shading with optional outlining.
type ToonShader struct {
	Matrix         Matrix
	LightDirection Vector
	CameraPosition Vector
	AmbientColor   Color
	DiffuseColor   Color
	
	// Cel Shading
	LightCutoff float64 // The point at which the light transitions to full shadow (0-1)
	ShadowBands float64 // The number of distinct shadow bands (e.g., 1, 2, 3)

	// Specular
	SpecularColor Color
	Glossiness    float64 // Smoothness of the specular highlight (0-1)

	// Rim Lighting
	RimColor Color
	RimSize  float64 // How much of the edge the rim light should cover (0-1)
}

func NewToonShader(matrix Matrix, lightDir, cameraPosition Vector, ambient, diffuse Color) *ToonShader {
	return &ToonShader{
		Matrix:         matrix,
		LightDirection: lightDir.Normalize(),
		CameraPosition: cameraPosition,
		AmbientColor:   ambient,
		DiffuseColor:   diffuse,
		LightCutoff: 0.5,
		ShadowBands: 2,

		// Specular Defaults
		SpecularColor: HexColor("ffffff"), // White
		Glossiness:    0.8,

		// Rim Lighting Defaults
		RimColor: HexColor("ffffff"), // White
		RimSize:  0.0, // Disabled by default
	}
}

func (s *ToonShader) Vertex(v Vertex) Vertex {
	v.Output = s.Matrix.MulPositionW(v.Position)
	normalMatrix := s.Matrix.Inverse().Transpose()
	v.Normal = normalMatrix.MulDirection(v.Normal).Normalize()
	return v
}

func (s *ToonShader) Fragment(v Vertex, fromObject *Object) Color {
	light := s.AmbientColor
	color := fromObject.Color
	if fromObject.Texture != nil {
		sample := fromObject.Texture.Sample(v.Texture.X, v.Texture.Y)
		if sample.A > 0 {
			color = color.Lerp(sample.DivScalar(sample.A), sample.A)
		}
	}

	nDotL := math.Max(0, v.Normal.Dot(s.LightDirection))
	shadow := math.Round(nDotL/s.LightCutoff*s.ShadowBands) / s.ShadowBands
	
	// Add the diffuse light, but use our stepped "shadow" value.
	light = light.Add(s.DiffuseColor.MulScalar(shadow))
	

	// The final color is the object's color multiplied by the calculated light.
	return color.Mul(light).Min(White) // Using .Min(White) to prevent color blowout
}
