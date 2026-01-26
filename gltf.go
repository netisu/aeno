package aeno

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

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

	if len(doc.Scenes) > 0 {
		sceneIdx := 0
		if doc.Scene != nil {
			sceneIdx = int(*doc.Scene)
		}
		for _, nodeIdx := range doc.Scenes[sceneIdx].Nodes {
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
	isDefaultMatrix := true
	for _, val := range node.Matrix {
		if val != 0 {
			isDefaultMatrix = false
			break
		}
	}

	if !isDefaultMatrix {
		m := node.Matrix
		local = Matrix{
			X00: float64(m[0]), X01: float64(m[4]), X02: float64(m[8]), X03: float64(m[12]),
			X10: float64(m[1]), X11: float64(m[5]), X12: float64(m[9]), X13: float64(m[13]),
			X20: float64(m[2]), X21: float64(m[6]), X22: float64(m[10]), X23: float64(m[14]),
			X30: float64(m[3]), X31: float64(m[7]), X32: float64(m[11]), X33: float64(m[15]),
		}
	} else {
		
		s := node.Scale
		sx, sy, sz := float64(s[0]), float64(s[1]), float64(s[2])
		if sx == 0 && sy == 0 && sz == 0 {
			sx, sy, sz = 1, 1, 1
		}
		local = local.Mul(Scale(V(sx, sy, sz)))

		r := node.Rotation
		rx, ry, rz, rw := float64(r[0]), float64(r[1]), float64(r[2]), float64(r[3])
		if rx == 0 && ry == 0 && rz == 0 && rw == 0 {
			rw = 1
		}
		rotMat := quaternionToMatrix(rx, ry, rz, rw)
		local = rotMat.Mul(local)

		// Translation
		t := node.Translation
		tx, ty, tz := float64(t[0]), float64(t[1]), float64(t[2])
		if tx != 0 || ty != 0 || tz != 0 {
			fmt.Printf("DEBUG: Applying Translation to %s: [%.2f, %.2f, %.2f]\n", node.Name, tx, ty, tz)
		}
		local = Translate(V(tx, ty, tz)).Mul(local)
	}

	// Calculate World Matrix
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
	if !ok {
		return nil
	}

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
		for k := range indices {
			indices[k] = uint32(k)
		}
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
