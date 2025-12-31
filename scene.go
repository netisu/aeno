package aeno

import (
	"image/png"
	"io"
	"math"
	"os"
)

type Scene struct {
	Context *Context
	Objects []*Object
	Eye     Vector
	Center  Vector
	Up      Vector
}

func NewScene(width, height int, shader Shader) *Scene {
	return &Scene{
		Context: NewContext(width, height, shader),
		Objects: []*Object{},
		Up:      Vector{0, 0, 1}, // Default Up
	}
}

func (s *Scene) Add(o *Object) {
	s.Objects = append(s.Objects, o)
}

// FitObjectsToScene automatically positions the eye to see all objects.
func (s *Scene) FitObjectsToScene(fovy, aspect, near, far float64) {
	if len(s.Objects) == 0 {
		return
	}

	box := s.Objects[0].Mesh.BoundingBox().Transform(s.Objects[0].Matrix)
	for i := 1; i < len(s.Objects); i++ {
		tBox := s.Objects[i].Mesh.BoundingBox().Transform(s.Objects[i].Matrix)
		box = box.Extend(tBox)
	}
	s.Center = box.Center()

	matrix := LookAt(s.Eye, s.Center, s.Up).Perspective(fovy, aspect, near, far)
	shader := NewPhongShader(matrix, Vector{}, s.Eye, HexColor("000000"), HexColor("000000"))

	var addedFOV float64
	allInside := false

	// We limit to 170 degrees to prevent mathematical inversion
	for !allInside && (fovy+addedFOV) < 170 {
		allInside = true // Assume true until proven otherwise
		currentMatrix := LookAt(s.Eye, s.Center, s.Up).Perspective(fovy+addedFOV, aspect, near, far)
		shader.Matrix = currentMatrix

		for _, o := range s.Objects {
			if o.Mesh == nil {
				continue
			}
			
			// Check every triangle in the object
			for _, t := range o.Mesh.Triangles {
				// Project vertices using the shader's matrix logic
				v1 := shader.Vertex(t.V1)
				v2 := shader.Vertex(t.V2)
				v3 := shader.Vertex(t.V3)

				if v1.Outside() || v2.Outside() || v3.Outside() {
					addedFOV += 1.0 // Increase FOV slightly
					allInside = false
					break // Re-check with new FOV
				}
			}
			if !allInside { break }
		}
	}

	// Update the Scene and Shader with the final working parameters
	finalFOV := fovy + addedFOV
	s.Shader.Matrix = LookAt(s.Eye, s.Center, s.Up).Perspective(finalFOV, aspect, near, far)
	s.Shader.CameraPosition = s.Eye
}

func (s *Scene) GetSafetyClipping() (near, far float64) {
	if len(s.Objects) == 0 {
		return 0.1, 1000.0
	}
	box := s.Objects[0].Mesh.BoundingBox().Transform(s.Objects[0].Matrix)
	for i := 1; i < len(s.Objects); i++ {
		box = box.Extend(s.Objects[i].Mesh.BoundingBox().Transform(s.Objects[i].Matrix))
	}

	distToCenter := s.Eye.Sub(box.Center()).Length()
	
	// The radius of the object (from center to furthest corner)
	radius := box.Size().Length() / 2.0

	// Near cannot be <= 0, or the projection matrix math explodes
	near = distToCenter - radius
	if near < 0.1 {
		near = 0.1 
	} else {
		near *= 0.9 
	}

	far = (distToCenter + radius) * 1.1

	return near, far
}

func (s *Scene) Render() {
	for _, o := range s.Objects {
		s.Context.DrawObject(o)
	}
}

func (s *Scene) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, s.Context.Image())
}

// GenerateScene is a high-level helper
func GenerateScene(path string, objects []*Object, width, height int, fitCamera bool) {
	// Default light and material
	light := Vector{0.5, 0.5, 1}
	color := HexColor("#FFF")
	
	// Initial camera (will be moved if fitCamera is true)
	eye := Vector{2, 2, 2}
	center := Vector{0, 0, 0}
	up := Vector{0, 0, 1}

	aspect := float64(width) / float64(height)
	fovy := 50.0

	// Setup Matrices
	view := LookAt(eye, center, up)
	proj := Perspective(fovy, aspect, 0.1, 1000)
	matrix := view.Mul(proj)

	shader := NewPhongShader(matrix, light, eye, color.MulScalar(0.2), color)
	
	scene := NewScene(width, height, shader)
	scene.Objects = objects
	scene.Eye = eye
	scene.Center = center
	scene.Up = up

	if fitCamera {
		scene.FitCamera(fovy, aspect)
		// Recompute matrix with new camera pos
		view = LookAt(scene.Eye, scene.Center, scene.Up)
		shader.Matrix = view.Mul(proj)
		shader.CameraPosition = scene.Eye
	}

	scene.Render()
	scene.Save(path)
}

func GenerateSceneToWriter(w io.Writer, objects []*Object, width, height int, fit bool) error {
	// Simplified pipeline for web/streams
	scene := NewScene(width, height, nil)
	// User must configure shader manually if using this low-level func, 
	// or we provide a default:
	scene.Objects = objects
	
	// Default setup
	scene.Eye = V(3, 3, 3)
	scene.Center = V(0, 0, 0)
	scene.Up = V(0, 0, 1)
	
	if fit {
		scene.FitCamera(45, float64(width)/float64(height))
	}
	
	m := LookAt(scene.Eye, scene.Center, scene.Up).Perspective(45, float64(width)/float64(height), 0.1, 100)
	scene.Context.Shader = NewPhongShader(m, V(1,1,1), scene.Eye, HexColor("#333"), HexColor("#FFF"))

	scene.Render()
	return png.Encode(w, scene.Context.Image())

}

