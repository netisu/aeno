package aeno

import (
    "github.com/nfnt/resize"
	"image/png"
	"io"
	"os"
	"log"
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
	
	viewMatrix := LookAt(s.Eye, s.Center, s.Up)
	matrix := viewMatrix.Perspective(fovy, aspect, near, far)
	shader := NewPhongShader(matrix, Vector{}, s.Eye, HexColor("000000"), HexColor("000000"))

	allMesh := NewEmptyMesh()
	var boxes []Box
	for _, o := range s.Objects {
		if o.Mesh == nil {
			continue
		}
		
		movedMesh := o.Mesh.Copy() 
		movedMesh.Transform(o.Matrix)
		
		allMesh.Add(movedMesh)
		bb := o.Mesh.BoundingBox()
		boxes = append(boxes, bb)
	}

	totalBox := BoxForBoxes(boxes)
	unitCube := NewCubeForBox(totalBox)
	unitCube.BiUnitCube() 
	
	// Physically transform the vertices of allMesh
	allMesh.FitInside(unitCube.BoundingBox(), V(0.5, 0.5, 0.5))

	indexed := 0
	var addedFOV float64
	for _, o := range s.Objects {
		if o.Mesh == nil { continue }
		
		num := len(o.Mesh.Triangles)
		// Extract the newly scaled triangles from the combined mesh
		tris := allMesh.Triangles[indexed : num+indexed]
		
		allInside := false
		for !allInside && len(tris) > 0 {
			for _, t := range tris {
				v1 := shader.Vertex(t.V1)
				v2 := shader.Vertex(t.V2)
				v3 := shader.Vertex(t.V3)

				if v1.Outside() || v2.Outside() || v3.Outside() {
					addedFOV += 5
					matrix = viewMatrix.Perspective(fovy+addedFOV, aspect, near, far)
					shader.Matrix = matrix
					allInside = false
				} else {
					allInside = true
				}
			}
		}

		o.Mesh = NewTriangleMesh(tris)
		o.Matrix = Identity()
		indexed += num
	}
	
	if phong, ok := s.Context.Shader.(*PhongShader); ok {
		phong.Matrix = viewMatrix.Perspective(fovy+addedFOV+2.0, aspect, near, far)
	}
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
func GenerateScene(fit bool, path string, objects []*Object, eye Vector, center Vector, up Vector, fovy float64, size int, scale int, light Vector, ambient string, diffuse string, near, far float64) {
	file, err := os.Create(path)
	if err != nil {
		log.Printf("aeno: could not create file for GenerateScene: %v", err)
		return
	}
	defer file.Close()

	// Direct call to the writer version
	err = GenerateSceneToWriter(file, objects, eye, center, up, fovy, size, scale, light, ambient, diffuse, near, far, fit)
	if err != nil {
		log.Printf("aeno: could not generate scene to file: %v", err)
	}
}

func GenerateSceneToWriter(writer io.Writer, objects []*Object, eye Vector, center Vector, up Vector, fovy float64, size int, scale int, light Vector, ambient string, diffuse string, near, far float64, fit bool) error {
	renderSize := size * scale
    aspect := float64(size) / float64(size)

	// Initial matrix setup
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)
	shader := NewPhongShader(matrix, light, eye, HexColor(ambient), HexColor(diffuse))
	
	scene := NewScene(renderSize, renderSize, shader)
	scene.Objects = objects
	scene.Eye = eye
	scene.Center = center
	scene.Up = up

	scene.Context.ClearColorBufferWith(Transparent)
	scene.Context.ClearDepthBuffer()

	if fit {
		scene.FitObjectsToScene(fovy, aspect, near, far)
	}

	scene.Render()
	
    if scale > 1 {
        // Use the resize package as in your original code
        resized := resize.Resize(uint(size), uint(size), scene.Context.Image(), resize.Bilinear)
        return png.Encode(writer, resized)
    }

	return png.Encode(writer, scene.Context.Image())
}


