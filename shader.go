package aeno

import (
	"math"
)

// Shader shader interface
type Shader interface {
	Vertex(Vertex) Vertex
	Fragment(Vertex, *Object) Color
}

// PhongShader implements Phong shading with an optional texture.
type PhongShader struct {
	Matrix         Matrix
	LightDirection Vector
	CameraPosition Vector
	AmbientColor   Color
	DiffuseColor   Color
	SpecularColor  Color
	SpecularPower  float64
	EnableOutline  bool    // A switch to turn the effect on/off
	OutlineColor   Color   // The color of the outline
	OutlineFactor  float64 // Controls line thickness (lower is thicker)
}

// NewPhongShader f
func NewPhongShader(matrix Matrix, lightDirection, cameraPosition Vector, ambient Color, diffuse Color) *PhongShader {
	specular := Color{1, 1, 1, 1}
	return &PhongShader{
		Matrix:         matrix,
		LightDirection: lightDirection,
		CameraPosition: cameraPosition,
		AmbientColor:   ambient,
		DiffuseColor:   diffuse,
		SpecularColor:  specular,
		SpecularPower:  0,
		EnableOutline:  true,                
		OutlineColor:   HexColor("000000"),  
		OutlineFactor:  0.05,                
	}
}

// Vertex f
func (shader *PhongShader) Vertex(v Vertex) Vertex {
	v.Output = shader.Matrix.MulPositionW(v.Position)
	normalMatrix := shader.Matrix.Inverse().Transpose()
	v.Normal = normalMatrix.MulDirection(v.Normal).Normalize()
	return v
}

// Fragment f
func (shader *PhongShader) Fragment(v Vertex, fromObject *Object) Color {
	if shader.EnableOutline {
		viewDirection := shader.CameraPosition.Sub(v.Position).Normalize()
		dot := viewDirection.Dot(v.Normal)

		// If the surface normal is nearly perpendicular to the view direction, it's an edge.
		if math.Abs(dot) < shader.OutlineFactor {
			return shader.OutlineColor
		}
	}
	// If the object is flagged to use vertex colors, we return the
	// interpolated vertex color and skip all lighting and texturing.
	if fromObject.UseVertexColor {
		return v.Color
	}
	
	light := shader.AmbientColor
	color := fromObject.Color
	if fromObject.Texture != nil {
		sample := fromObject.Texture.Sample(v.Texture.X, v.Texture.Y)
		if sample.A > 0 {
			color = color.Lerp(sample.DivScalar(sample.A), sample.A)
		}
	}
	diffuse := math.Max(v.Normal.Dot(shader.LightDirection), 0)
	light = light.Add(shader.DiffuseColor.MulScalar(diffuse))
	if diffuse > 0 && shader.SpecularPower > 0 {
		camera := shader.CameraPosition.Sub(v.Position).Normalize()
		reflected := shader.LightDirection.Negate().Reflect(v.Normal)
		specular := math.Max(camera.Dot(reflected), 0)
		if specular > 0 {
			specular = math.Pow(specular, shader.SpecularPower)
			light = light.Add(shader.SpecularColor.MulScalar(specular))
		}
	}
	if color.A < 1 {
		return color.Mul(light).Min(White).DivScalar(color.A).Alpha(color.A)
	}

	return color.Mul(light).Min(White).Alpha(color.A)
}



