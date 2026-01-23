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
				fillVertex := func(v *Vector, idx uint32) {
					// Position
					v.Position = Vector{
						float64(positions[idx][0]),
						float64(positions[idx][1]),
						float64(positions[idx][2]),
					}
					// Normal
					if int(idx) < len(normals) {
						v.Normal = Vector{
							float64(normals[idx][0]),
							float64(normals[idx][1]),
							float64(normals[idx][2]),
						}
					}
					// Texture
					if int(idx) < len(texCoords) {
						tex.Texture = Vector{
							float64(texCoords[idx][0]),
							float64(texCoords[idx][1]),
							0,
						}
					}
				}

				fillVertex(&t.V1, indices[i], &t.V1)
				fillVertex(&t.V2, indices[i+1], &t.V2)
				fillVertex(&t.V3, indices[i+2], &t.V3)
				
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
