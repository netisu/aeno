package aeno

import "net/http"

type Object struct {
	Mesh           *Mesh
	Texture        Texture
	Color          Color
	Matrix         Matrix
	UseVertexColor bool
}

func NewObject(mesh *Mesh) *Object {
	return &Object{
		Mesh:   mesh,
		Matrix: Identity(),
		Color:  White,
	}
}

func NewObjectFromFile(path string) (*Object, error) {
	mesh, err := LoadOBJ(path)
	if err != nil {
		return nil, err
	}
	return NewObject(mesh), nil
}

func NewObjectFromURL(url string) *Object {
	resp, err := http.Get(url)
	if err != nil {
		return NewObject(NewEmptyMesh())
	}
	defer resp.Body.Close()
	mesh, _ := LoadOBJFromReader(resp.Body)
	return NewObject(mesh)
}

func (o *Object) SetColor(c Color) {
	o.Color = c
}

func (o *Object) Transform(m Matrix) {
	o.Matrix = o.Matrix.Mul(m)
}