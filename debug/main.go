package main

import (
	"fmt"
	"log"
	
	"github.com/netisu/aeno"
)

func main() {
	path := "arm_left.glb" 
	
	fmt.Println("--- STARTING DEBUG ---")
	mesh, err := aeno.LoadGLTF(path)
	if err != nil {
		log.Fatal(err)
	}

	box := mesh.BoundingBox()
	center := box.Center()
	
	fmt.Printf("--- MESH STATS ---\n")
	fmt.Printf("Triangles: %d\n", len(mesh.Triangles))
	fmt.Printf("Bounding Box Min: %+v\n", box.Min)
	fmt.Printf("Bounding Box Max: %+v\n", box.Max)
	fmt.Printf("Bounding Box Center: %+v\n", center)
	
	expectedY := 5.25
	if center.Y < expectedY-1.0 || center.Y > expectedY+1.0 {
		fmt.Printf("Mesh is NOT translated correctly. Expected Y ~%.2f, got %.2f\n", expectedY, center.Y)
	} else {
		fmt.Printf("Mesh translation applied!\n")
	}
}
