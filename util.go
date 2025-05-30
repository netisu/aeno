package aeno

import (
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Radians f
func Radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// Degrees f
func Degrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

// LatLngToXYZ f
func LatLngToXYZ(lat, lng float64) Vector {
	lat, lng = Radians(lat), Radians(lng)
	x := math.Cos(lat) * math.Cos(lng)
	y := math.Cos(lat) * math.Sin(lng)
	z := math.Sin(lat)
	return Vector{x, y, z}
}

// LoadMesh f
func LoadMesh(path string) (*Mesh, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".obj":
		return LoadOBJ(path)
	}
	return nil, fmt.Errorf("unrecognized mesh extension: %s", ext)
}

// LoadImage f
func LoadImage(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	im, _, err := image.Decode(file)
	return im, err
}

// SavePNG f
func SavePNG(path string, im image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, im)
}

// ParseFloats f
func ParseFloats(items []string) []float64 {
	result := make([]float64, len(items))
	for i, item := range items {
		f, _ := strconv.ParseFloat(item, 64)
		result[i] = f
	}
	return result
}

// Clamp f
func Clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

// ClampInt f
func ClampInt(x, lo, hi int) int {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

// AbsInt f
func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Round f
func Round(a float64) int {
	if a < 0 {
		return int(math.Ceil(a - 0.5))
	}

	return int(math.Floor(a + 0.5))
}

// RoundPlaces f
func RoundPlaces(a float64, places int) float64 {
	shift := powersOfTen[places]
	return float64(Round(a*shift)) / shift
}

var powersOfTen = []float64{1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16}
