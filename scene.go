package aeno

import (
	"bytes"
	"image/png"
	"log"
	"sync"
	"io"
	"os"

	"github.com/nfnt/resize"
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

// FitObjectsToScene fits the objects into a 0.5 unit bounding box
func (s *Scene) FitObjectsToScene(eye, center, up Vector, fovy, aspect, near, far float64) (matrix Matrix) {
	matrix = LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	shader := NewPhongShader(matrix, Vector{}, eye, HexColor("000000"), HexColor("000000"))

	allMesh := NewEmptyMesh()
	var boxes []Box
	for _, o := range s.Objects {
		if o.Mesh == nil {
			continue
		}
		allMesh.Add(o.Mesh)
		bb := o.Mesh.BoundingBox()
		boxes = append(boxes, bb)
	}
	box := BoxForBoxes(boxes)
	b := NewCubeForBox(box)
	b.BiUnitCube()
	allMesh.FitInside(b.BoundingBox(), V(0.5, 0.5, 0.5))

	indexed := 0
	var addedFOV float64
	for _, o := range s.Objects {
		if o.Mesh == nil {
			continue
		}
		num := len(o.Mesh.Triangles)
		tris := allMesh.Triangles[indexed : num+indexed]
		allInside := false
		for !allInside && len(tris) > 0 {
			for _, t := range tris {
				v1 := shader.Vertex(t.V1)
				v2 := shader.Vertex(t.V2)
				v3 := shader.Vertex(t.V3)

				if v1.Outside() || v2.Outside() || v3.Outside() {
					addedFOV += 5
					matrix = LookAt(eye, center, up).Perspective(fovy+addedFOV, aspect, near, far)
					shader.Matrix = matrix
					allInside = false
				} else {
					allInside = true
				}
			}
		}

		o.Mesh = NewTriangleMesh(tris)
		indexed += num
	}

	return
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

	buf := new(bytes.Buffer)
	png.Encode(buf, image)
	SavePNG(path, image)
	return
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
	return png.Encode(writer, img)
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





