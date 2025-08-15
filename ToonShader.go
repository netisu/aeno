
package aeno

import "math"

// ToonShader implements cel shading with optional outlining.
type ToonShader struct {
	Matrix         Matrix
	LightDirection Vector
	CameraPosition Vector
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

func NewToonShader(matrix Matrix, lightDir, cameraPosition Vector) *ToonShader {
	return &ToonShader{
		Matrix:         matrix,
		LightDirection: lightDir.Normalize(),
		CameraPosition: cameraPosition,
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
	// Get Base Color (Albedo)
	albedo := fromObject.Color
	if fromObject.Texture != nil {
		texColor := fromObject.Texture.Sample(v.Texture.X, v.Texture.Y)
		albedo = albedo.Lerp(texColor.DivScalar(texColor.A), texColor.A)
	}

	nDotL := math.Max(0, v.Normal.Dot(s.LightDirection))
	shadow := math.Round(nDotL/s.LightCutoff*s.ShadowBands) / s.ShadowBands

	// Specular Highlight
	reflectedLight := s.LightDirection.Negate().Reflect(v.Normal)
	vDotReflected := math.Max(0, s.CameraPosition.Sub(v.Position).Normalize().Dot(reflectedLight))
	specular := Color{} // Black by default
	if vDotReflected > (1.0 - s.Glossiness) {
		specular = s.SpecularColor
	}

	// Rim Light
	viewDir := s.CameraPosition.Sub(v.Position).Normalize()
	rimFactor := 1.0 - math.Max(0, viewDir.Dot(v.Normal))
	rim := Color{} // Black by default
	if rimFactor > (1.0 - s.RimSize) {
		rim = s.RimColor
	}

	basePlusSpecular := albedo.Add(specular)
	shadedColor := basePlusSpecular.MulScalar(shadow)
	finalColor := shadedColor.Add(rim)
	return finalColor
}
