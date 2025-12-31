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

// FitCamera automatically positions the eye to see all objects.
func (s *Scene) FitCamera(fovy, aspect float64) {
	if len(s.Objects) == 0 {
		return
	}
	
	// Calculate bounding box of all objects
	box := EmptyBox
	for _, o := range s.Objects {
		tBox := o.Mesh.BoundingBox().Transform(o.Matrix)
		box = box.Extend(tBox)
	}

	// Center the camera target on the center of the scene
	s.Center = box.Center()

	dir := s.Center.Sub(s.Eye).Normalize()
	if dir.Length() == 0 {
		dir = Vector{0, -1, 0}
	}
	
	// Radius of the scene's bounding sphere
	radius := box.Size().Length() / 2.0
	

	sinFov := math.Sin(Radians(fovy) / 2.0)
	distance := radius / sinFov
	
	// Add a little padding (10%)
	distance *= 1.1

	s.Eye = s.Center.Sub(dir.MulScalar(distance))
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