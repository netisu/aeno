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
	box := s.Objects[0].Mesh.BoundingBox().Transform(s.Objects[0].Matrix)
	for i := 1; i < len(s.Objects); i++ {
		tBox := s.Objects[i].Mesh.BoundingBox().Transform(s.Objects[i].Matrix)
		box = box.Extend(tBox)
	}

	// Center the camera target on the center of the scene
	s.Center = box.Center()

	// If Eye and Center are the same, default to looking from the front (+Z)
	dir := s.Eye.Sub(s.Center).Normalize()
	if dir.Length() < 0.0001 {
		dir = Vector{0, 0, 1} 
	}
	
	size := box.Size()
	maxDim := math.Max(size.X, math.Max(size.Y, size.Z))
	
	// We use the smaller of vertical/horizontal FOV to ensure no clipping
	theta := Radians(fovy) / 2.0
	if aspect < 1.0 {
		// If portrait, the horizontal FOV is smaller
		theta = math.Atan(math.Tan(theta) * aspect)
	}
	
	distance := (maxDim / 2.0) / math.Tan(theta)
	
	// 10% padding
	distance *= 1.1

	// Update Eye position
	s.Eye = s.Center.Add(dir.MulScalar(distance))
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
