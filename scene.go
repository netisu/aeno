package aeno

import (
	"image/png"
	"log"
	"sync"
	"io"
	"os"
)

// Scene struct to store all data for a scene
type Scene struct {
	Context         *Context
	Objects         []*Object
	Shader          Shader
	eye, center, up Vector
	fovy, aspect    float64
}

// NewScene returns a new scene
func NewScene(eye Vector, center Vector, up Vector, fovy float64, size int, scale int, shader Shader) *Scene {
	aspect := float64(size) / float64(size)
	context := NewContext(size*scale, size*scale, 5, shader)
	return &Scene{context, nil, shader, eye, center, up, fovy, aspect}
}
// AddObject adds an object to the scene
func (s *Scene) AddObject(o *Object) {
	s.Objects = append(s.Objects, o)
}

// AddObjects is a convenience method to add multiple objects
func (s *Scene) AddObjects(objects []*Object) {
	for _, o := range objects {
		s.AddObject(o)
	}
}

// A helper function to update the matrix in a shader if the shader supports it.
func updateShaderMatrix(shader Shader, matrix Matrix) {
	if s, ok := shader.(interface{ SetMatrix(Matrix) }); ok {
		s.SetMatrix(matrix)
	} else if p, ok := shader.(*PhongShader); ok { // Fallback for specific known types
		p.Matrix = matrix
	} else if t, ok := shader.(*ToonShader); ok {
		t.Matrix = matrix
	}
}

// FitObjectsToScene fits the objects into a 0.5 unit bounding box
func (s *Scene) FitObjectsToScene(eye, center, up Vector, fovy, aspect, near, far float64) (matrix Matrix) {
	currentFovy := fovy
	for { // Loop indefinitely until we find a FOV that contains all geometry
		allInside := true
		matrix = LookAt(eye, center, up).Perspective(currentFovy, aspect, near, far)

		for _, o := range s.Objects {
			if o.Mesh == nil {
				continue
			}
			for _, t := range o.Mesh.Triangles {
				v1 := Vertex{Output: matrix.MulPositionW(t.V1.Position)}
				v2 := Vertex{Output: matrix.MulPositionW(t.V2.Position)}
				v3 := Vertex{Output: matrix.MulPositionW(t.V3.Position)}

				// The .Outside() method checks if the transformed vertex is outside the canonical view volume.
				if v1.Outside() || v2.Outside() || v3.Outside() {
					allInside = false
					break 
				}
			}
			if !allInside {
				break
			}
		}

		if allInside {
			return matrix
		}

		// If we're here, at least one vertex was outside. Increase the FOV and try again.
		currentFovy += 2.0
		if currentFovy >= 175 {
			// Add a safety break to prevent an infinite loop in case of extreme geometric configurations.
			log.Println("aeno: FitObjectsToScene FOV reached 175 degrees, returning current best-fit matrix.")
			return matrix
		}
	}
}

// Draw draws the scene
func (s *Scene) Draw(fit bool, path string, objects []*Object) {
	s.AddObjects(objects)
	if fit {
		newMatrix := s.FitObjectsToScene(s.eye, s.center, s.up, s.fovy, s.aspect, 1, 999)
		if p, ok := s.Shader.(*PhongShader); ok {
			p.Matrix = newMatrix
		} else if t, ok := s.Shader.(*ToonShader); ok {
			t.Matrix = newMatrix
		}
	}
	var wg sync.WaitGroup
	wg.Add(len(s.Objects))
	for _, o := range s.Objects {
		if o.Mesh == nil {
			wg.Done()
			log.Printf("Object attempted to render with nil mesh")
			continue
		}
		s.Context.DrawObject(o, &wg)
	}
	wg.Wait()

	file, err := os.Create(path)
	if err != nil {
		log.Printf("aeno: could not create file in Draw: %v", err)
		return
	}
	defer file.Close()

	if err := png.Encode(file, s.Context.Image()); err != nil {
		log.Printf("aeno: could not encode png in Draw: %v", err)
	}
}

func (s *Scene) DrawToWriter(fit bool, writer io.Writer, objects []*Object) error {
	s.AddObjects(objects)
	if fit {
		newMatrix := s.FitObjectsToScene(s.eye, s.center, s.up, s.fovy, s.aspect, 1, 999)
		if p, ok := s.Shader.(*PhongShader); ok {
			p.Matrix = newMatrix
		} else if t, ok := s.Shader.(*ToonShader); ok {
			t.Matrix = newMatrix
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(s.Objects))
	for _, o := range s.Objects {
		if o.Mesh == nil {
			log.Printf("Object attempted to render with nil mesh")
			wg.Done()
			continue
		}
		s.Context.DrawObject(o, &wg)
	}
	wg.Wait()
	
	// Encode the final image directly to the provided writer.
	return png.Encode(writer, s.Context.Image())
}

func GenerateScene(fit bool, path string, objects []*Object, eye Vector, center Vector, up Vector, fovy float64, size int, scale int, light Vector, ambient string, diffuse string, near, far float64) {
	file, err := os.Create(path)
	if err != nil {
		log.Printf("aeno: could not create file for GenerateScene: %v", err)
		return
	}
	defer file.Close()

	err = GenerateSceneToWriter(file, objects, eye, center, up, fovy, size, scale, light, ambient, diffuse, near, far, fit)
	if err != nil {
		log.Printf("aeno: could not generate scene to file: %v", err)
	}
}
func GenerateSceneWithShader(fit bool, shader Shader, path string, objects []*Object, eye Vector, center Vector, up Vector, fovy float64, size int, scale int) {
	// Directly pass the provided shader to the scene
	scene := NewScene(eye, center, up, fovy, size, scale, shader)
	scene.Draw(fit, path, objects)
}

func GenerateSceneToWriter(writer io.Writer, objects []*Object, eye Vector, center Vector, up Vector, fovy float64, size int, scale int, light Vector, ambient string, diffuse string, near, far float64, fit bool) error {
	aspect := float64(size) / float64(size)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	
	shader := NewPhongShader(matrix, light, eye, HexColor(ambient), HexColor(diffuse))
	scene := NewScene(eye, center, up, fovy, size, scale, shader)

	// Call the new core drawing method.
	return scene.DrawToWriter(fit, writer, objects)
}








