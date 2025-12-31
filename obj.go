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
	vs := []Vector{{}} 
	vts := []Vector{{}}
	vns := []Vector{{}}
	
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
			y := 0.0
			if len(fields) > 2 { y = pf(fields[2]) }
			vts = append(vts, Vector{pf(fields[1]), y, 0})
		case "vn":
			vns = append(vns, Vector{pf(fields[1]), pf(fields[2]), pf(fields[3])})
		case "f":
			fvs, fvts, fvns := parseFace(fields[1:])
			// Fan triangulation
			for i := 1; i < len(fvs)-1; i++ {
				t := &Triangle{}
				
				// Apply fixIndex using the CURRENT length of the slices
				t.V1.Position = vs[fixIndex(fvs[0], len(vs))]
				t.V2.Position = vs[fixIndex(fvs[i], len(vs))]
				t.V3.Position = vs[fixIndex(fvs[i+1], len(vs))]
				
				if len(vts) > 1 {
					t.V1.Texture = vts[fixIndex(fvts[0], len(vts))]
					t.V2.Texture = vts[fixIndex(fvts[i], len(vts))]
					t.V3.Texture = vts[fixIndex(fvts[i+1], len(vts))]
				}
				if len(vns) > 1 {
					t.V1.Normal = vns[fixIndex(fvns[0], len(vns))]
					t.V2.Normal = vns[fixIndex(fvns[i], len(vns))]
					t.V3.Normal = vns[fixIndex(fvns[i+1], len(vns))]
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
func fixIndex(i, n int) int {
	if i > 0 {
		return i
	}
	if i < 0 {
		return n + i
	}
	return 0
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