package aeno

import (
	"fmt"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
)

// LoadGLTF loads a .gltf or .glb file and converts it to an aeno.Mesh
func LoadGLTF(path string) (*Mesh, error) {
	doc, err := gltf.Open(path)
	if err != nil {
		return nil, err
	}

	var allTriangles []*Triangle

	for _, mesh := range doc.Meshes {
		for _, primitive := range mesh.Primitives {
			// We only support Triangles (mode 4)
			if primitive.Mode != gltf.PrimitiveTriangles {
				continue
			}

			// extract positions
			posIdx, ok := primitive.Attributes[gltf.POSITION]
			if !ok {
				continue
			}
			positions, err := modeler.ReadPosition(doc, doc.Accessors[posIdx], nil)
			if err != nil {
				return nil, err
			}
			
			// extract normals
			var normals [][3]float32
			if normIdx, ok := primitive.Attributes[gltf.NORMAL]; ok {
				normals, _ = modeler.ReadNormal(doc, doc.Accessors[normIdx], nil)
			}

			// Read Texture Coordinates
			var texCoords [][2]float32
			if texIdx, ok := primitive.Attributes[gltf.TEXCOORD_0]; ok {
				texCoords, _ = modeler.ReadTextureCoord(doc, doc.Accessors[texIdx], nil)
			}

			var indices []uint32
			if primitive.Indices != nil {
				// ReadIndices automatically converts uint8/uint16/uint32 to []uint32
				indices, err = modeler.ReadIndices(doc, doc.Accessors[*primitive.Indices], nil)
				if err != nil {
					return nil, err
				}
			} else {
				// If no indices are provided, generate linear indices (0, 1, 2, ...)
				indices = make([]uint32, len(positions))
				for k := range indices {
					indices[k] = uint32(k)
				}
			}

			for i := 0; i < len(indices); i += 3 {
				t := &Triangle{}
				i1, i2, i3 := indices[i], indices[i+1], indices[i+2]

				t.V1.Position = Vector{float64(positions[i1][0]), float64(positions[i1][1]), float64(positions[i1][2])}
				if len(normals) > int(i1) {
					t.V1.Normal = Vector{float64(normals[i1][0]), float64(normals[i1][1]), float64(normals[i1][2])}
				}
				if len(texCoords) > int(i1) {
					t.V1.Texture = Vector{float64(texCoords[i1][0]), float64(texCoords[i1][1]), 0}
				}

				t.V2.Position = Vector{float64(positions[i2][0]), float64(positions[i2][1]), float64(positions[i2][2])}
				if len(normals) > int(i2) {
					t.V2.Normal = Vector{float64(normals[i2][0]), float64(normals[i2][1]), float64(normals[i2][2])}
				}
				if len(texCoords) > int(i2) {
					t.V2.Texture = Vector{float64(texCoords[i2][0]), float64(texCoords[i2][1]), 0}
				}

				t.V3.Position = Vector{float64(positions[i3][0]), float64(positions[i3][1]), float64(positions[i3][2])}
				if len(normals) > int(i3) {
					t.V3.Normal = Vector{float64(normals[i3][0]), float64(normals[i3][1]), float64(normals[i3][2])}
				}
				if len(texCoords) > int(i3) {
					t.V3.Texture = Vector{float64(texCoords[i3][0]), float64(texCoords[i3][1]), 0}
				}
				
				t.FixNormals()
				allTriangles = append(allTriangles, t)
			}
		}
	}

	if len(allTriangles) == 0 {
		return nil, fmt.Errorf("no triangles found in gltf")
	}

	return NewTriangleMesh(allTriangles), nil
}
