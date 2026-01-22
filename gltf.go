package aeno

import (
	"fmt"
	"github.com/qmuntal/gltf"
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
			positions, err := gltf.Modeler{Model: doc}.Position(primitive.Indices, primitive.Attributes)
			if err != nil {
				continue
			}
			
			// extract normals
			normals, _ := gltf.Modeler{Model: doc}.Normal(primitive.Indices, primitive.Attributes)
			
			// extract texture coords
			texCoords, _ := gltf.Modeler{Model: doc}.TextureCoord(0, primitive.Indices, primitive.Attributes)

			// GLTF returns flat arrays, we need to group them into triplets
			for i := 0; i < len(positions); i += 3 {
				t := &Triangle{}

				// Vertex 1
				t.V1.Position = Vector{float64(positions[i][0]), float64(positions[i][1]), float64(positions[i][2])}
				if len(normals) > i {
					t.V1.Normal = Vector{float64(normals[i][0]), float64(normals[i][1]), float64(normals[i][2])}
				}
				if len(texCoords) > i {
					t.V1.Texture = Vector{float64(texCoords[i][0]), float64(texCoords[i][1]), 0}
				}

				// Vertex 2
				t.V2.Position = Vector{float64(positions[i+1][0]), float64(positions[i+1][1]), float64(positions[i+1][2])}
				if len(normals) > i+1 {
					t.V2.Normal = Vector{float64(normals[i+1][0]), float64(normals[i+1][1]), float64(normals[i+1][2])}
				}
				if len(texCoords) > i+1 {
					t.V2.Texture = Vector{float64(texCoords[i+1][0]), float64(texCoords[i+1][1]), 0}
				}

				// Vertex 3
				t.V3.Position = Vector{float64(positions[i+2][0]), float64(positions[i+2][1]), float64(positions[i+2][2])}
				if len(normals) > i+2 {
					t.V3.Normal = Vector{float64(normals[i+2][0]), float64(normals[i+2][1]), float64(normals[i+2][2])}
				}
				if len(texCoords) > i+2 {
					t.V3.Texture = Vector{float64(texCoords[i+2][0]), float64(texCoords[i+2][1]), 0}
				}

				allTriangles = append(allTriangles, t)
			}
		}
	}

	if len(allTriangles) == 0 {
		return nil, fmt.Errorf("no triangles found in gltf")
	}

	return NewTriangleMesh(allTriangles), nil
}
