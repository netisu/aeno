package aeno

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"
)

func LoadOBJ(path string) (*Mesh, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadOBJFromReader(file)
}

func LoadOBJFromBytes(b []byte) (*Mesh, error) {
	return LoadOBJFromReader(bytes.NewReader(b))
}

func LoadOBJFromReader(r io.Reader) (*Mesh, error) {
	vs := make([]Vector, 1, 1024)
	vts := make([]Vector, 1, 1024)
	vns := make([]Vector, 1, 1024)
	
	var triangles []*Triangle
	scanner := bufio.NewScanner(r)
	
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 || line[0] == '#' {
			continue
		}
		
		fields := strings.Fields(line)
		if len(fields) == 0 { continue }
		
		switch fields[0] {
		case "v":
			vs = append(vs, Vector{pf(fields[1]), pf(fields[2]), pf(fields[3])})
		case "vt":
			vts = append(vts, Vector{pf(fields[1]), pf(fields[2]), 0})
		case "vn":
			vns = append(vns, Vector{pf(fields[1]), pf(fields[2]), pf(fields[3])})
		case "f":
			args := fields[1:]
			fvs := make([]int, len(args))
			fvts := make([]int, len(args))
			fvns := make([]int, len(args))
			
			for i, arg := range args {
				vertex := strings.Split(arg+"//", "/")
				fvs[i] = fixIndex(vertex[0], len(vs))
				fvts[i] = fixIndex(vertex[1], len(vts))
				fvns[i] = fixIndex(vertex[2], len(vns))
			}

			for i := 1; i < len(fvs)-1; i++ {
				t := &Triangle{}
				i1, i2, i3 := 0, i, i+1
				
				t.V1.Position = vs[fvs[i1]]
				t.V2.Position = vs[fvs[i2]]
				t.V3.Position = vs[fvs[i3]]
				
				if fvns[i1] > 0 {
					t.V1.Normal = vns[fvns[i1]]
					t.V2.Normal = vns[fvns[i2]]
					t.V3.Normal = vns[fvns[i3]]
				}
				if fvts[i1] > 0 {
					t.V1.Texture = vts[fvts[i1]]
					t.V2.Texture = vts[fvts[i2]]
					t.V3.Texture = vts[fvts[i3]]
				}
				
				t.FixNormals()
				triangles = append(triangles, t)
			}
		}
	}
	return NewTriangleMesh(triangles), scanner.Err()
}

// Helper for fast float parsing
func pf(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// Helper to handle negative indices in OBJ
func fixIndex(value string, length int) int {
	if value == "" {
		return 0
	}
	parsed, _ := strconv.Atoi(value)
	if parsed < 0 {
		return parsed + length
	}
	return parsed
}

func parseFace(args []string) ([]int, []int, []int) {
	n := len(args)
	vi := make([]int, n)
	vt := make([]int, n)
	vn := make([]int, n)
	
	for i, s := range args {
		// Manual split is faster than strings.Split for simple cases
		// formats: v, v/vt, v//vn, v/vt/vn
		parts := strings.Split(s, "/")
		
		vi[i], _ = strconv.Atoi(parts[0])
		
		if len(parts) > 1 && parts[1] != "" {
			vt[i], _ = strconv.Atoi(parts[1])
		}
		if len(parts) > 2 && parts[2] != "" {
			vn[i], _ = strconv.Atoi(parts[2])
		}
	}
	return vi, vt, vn
}