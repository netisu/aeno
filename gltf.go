package aeno

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

// LoadGLTF loads a .gltf or .glb file and converts it to an aeno.Mesh
func LoadGLTF(path string) (*Mesh, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadGLTFFromReader(file)
}
func LoadGLTFFromBytes(b []byte) (*Mesh, error) {
	return LoadGLTFFromReader(bytes.NewReader(b))
}

func LoadGLTFFromReader(r io.Reader) (*Mesh, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	doc := new(gltf.Document)
	if err := gltf.NewDecoder(bytes.NewReader(data)).Decode(doc); err != nil {
		return nil, err
	}

	var allTriangles []*Triangle

	// GLTF is hierarchical. We must traverse nodes to get the correct positions.
	if len(doc.Scenes) > 0 {
		sceneIdx := uint32(0)
		if doc.Scene != nil {
			sceneIdx = *doc.Scene
		}
		for _, nodeIdx := range doc.Scenes[doc.Scene].Nodes {
			allTriangles = append(allTriangles, processGLTFNode(doc, doc.Nodes[nodeIdx], Identity())...)
		}
	}

	if len(allTriangles) == 0 {
		return nil, fmt.Errorf("no triangles found in gltf")
	}

	return NewTriangleMesh(allTriangles), nil
}
func processGLTFNode(doc *gltf.Document, node *gltf.Node, parentTransform Matrix) []*Triangle {
	var triangles []*Triangle

	local := Identity()
	
	if node.Matrix != [16]float32{} {
		m := node.Matrix
		local = Matrix{
			float64(m[0]), float64(m[4]), float64(m[8]), float64(m[12]),
			float64(m[1]), float64(m[5]), float64(m[9]), float64(m[13]),
			float64(m[2]), float64(m[6]), float64(m[10]), float64(m[14]),
			float64(m[3]), float64(m[7]), float64(m[11]), float64(m[15]),
		}
	} else {
		if node.Translation != [3]float32{0, 0, 0} {
			t := node.Translation
			local = local.Mul(Translate(V(float64(t[0]), float64(t[1]), float64(t[2]))))
		}
		if node.Rotation != [4]float32{0, 0, 0, 1} {
			r := node.Rotation
			local = local.Mul(quaternionToMatrix(float64(r[0]), float64(r[1]), float64(r[2]), float64(r[3])))
		}
		if node.Scale != [3]float32{1, 1, 1} {
			s := node.Scale
			local = local.Mul(Scale(V(float64(s[0]), float64(s[1]), float64(s[2]))))
		}
	}

	worldMatrix := parentTransform.Mul(local)

	if node.Mesh != nil {
		mesh := doc.Meshes[*node.Mesh]
		for _, primitive := range mesh.Primitives {
			triangles = append(triangles, extractGLTFPrimitive(doc, primitive, worldMatrix)...)
		}
	}

	for _, childIdx := range node.Children {
		triangles = append(triangles, processGLTFNode(doc, doc.Nodes[childIdx], worldMatrix)...)
	}

	return triangles
}

func extractGLTFPrimitive(doc *gltf.Document, primitive *gltf.Primitive, transform Matrix) []*Triangle {
	var triangles []*Triangle

	posIdx, ok := primitive.Attributes[gltf.POSITION]
	if !ok { return nil }
	
	positions, _ := modeler.ReadPosition(doc, doc.Accessors[posIdx], nil)
	
	var normals [][3]float32
	if normIdx, ok := primitive.Attributes[gltf.NORMAL]; ok {
		normals, _ = modeler.ReadNormal(doc, doc.Accessors[normIdx], nil)
	}

	var texCoords [][2]float32
	if texIdx, ok := primitive.Attributes[gltf.TEXCOORD_0]; ok {
		texCoords, _ = modeler.ReadTextureCoord(doc, doc.Accessors[texIdx], nil)
	}

	var indices []uint32
	if primitive.Indices != nil {
		indices, _ = modeler.ReadIndices(doc, doc.Accessors[*primitive.Indices], nil)
	} else {
		indices = make([]uint32, len(positions))
		for k := range indices { indices[k] = uint32(k) }
	}

	for i := 0; i < len(indices); i += 3 {
		t := &Triangle{}
		idxs := []uint32{indices[i], indices[i+1], indices[i+2]}
		verts := []*Vertex{&t.V1, &t.V2, &t.V3}

		for j, idx := range idxs {
			localPos := V(float64(positions[idx][0]), float64(positions[idx][1]), float64(positions[idx][2]))
			verts[j].Position = transform.MulPosition(localPos)
			if len(normals) > int(idx) {
				localNorm := V(float64(normals[idx][0]), float64(normals[idx][1]), float64(normals[idx][2]))
				verts[j].Normal = transform.MulDirection(localNorm)			
			}
			if len(texCoords) > int(idx) {
				verts[j].Texture = V(float64(texCoords[idx][0]), float64(texCoords[idx][1]), 0)
			}
		}

		t.FixNormals()
		triangles = append(triangles, t)
	}
	return triangles
}

func quaternionToMatrix(x, y, z, w float64) Matrix {
	m := Identity()
	m.X00 = 1 - 2*y*y - 2*z*z
	m.X01 = 2*x*y - 2*z*w
	m.X02 = 2*x*z + 2*y*w

	m.X10 = 2*x*y + 2*z*w
	m.X11 = 1 - 2*x*x - 2*z*z
	m.X12 = 2*y*z - 2*x*w

	m.X20 = 2*x*z - 2*y*w
	m.X21 = 2*y*z + 2*x*w
	m.X22 = 1 - 2*x*x - 2*y*y
	return m
}
