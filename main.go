package main

import (
	"bufio"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const (
	totalWidth  = 1024
	totalHeight = 768
)

var (
	mint      = color.RGBA{245, 255, 250, 255}
	slateGray = color.RGBA{112, 128, 144, 255}
	black     = color.RGBA{0, 0, 0, 255}
)

type vec3f struct {
	X float64
	Y float64
	Z float64
}

func (v *vec3f) Multiply(rhs *vec3f) float64 {
	var ret float64
	ret += v.X * rhs.X
	ret += v.Y * rhs.Y
	ret += v.Z * rhs.Z
	return ret
}

func (v *vec3f) MultiplyF(rhs float64) *vec3f {
	ret := &vec3f{}
	ret.X = v.X * rhs
	ret.Y = v.Y * rhs
	ret.Z = v.Z * rhs
	return ret

}

func (v *vec3f) Subtract(rhs *vec3f) *vec3f {
	ret := &vec3f{}
	ret.X = v.X - rhs.X
	ret.Y = v.Y - rhs.Y
	ret.Z = v.Z - rhs.Z
	return ret
}

func (v *vec3f) Normalize() *vec3f {
	return v.MultiplyF(1 / v.norm())
}

func (v *vec3f) norm() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Sphere is represented by a vec3f center and a float64 radius
type Sphere struct {
	Center vec3f
	Radius float64
	Color  color.RGBA
}

// RayIntersect determines if the provided ray interescts with s.
// If an interesction occurs, the distance is also returns.
// If not intersections, a zero value for distance is returned
func (s Sphere) RayIntersect(orig *vec3f, dir *vec3f) (bool, float64) {
	l := s.Center.Subtract(orig)
	tca := l.Multiply(dir)
	d2 := l.Multiply(l) - tca*tca
	if d2 > s.Radius*s.Radius {
		return false, 0
	}
	thc := math.Sqrt(s.Radius*s.Radius - d2)
	t := tca - thc
	t1 := tca + thc
	if t < 0 {
		t = t1
	}
	if t < 0 {
		return false, 0
	}
	return true, t
}

func castRay(orig *vec3f, dir *vec3f, spheres []*Sphere) color.RGBA {
	curDist := math.MaxFloat64
	curColor := color.RGBA{55, 176, 202, 255}
	for _, sphere := range spheres {
		intersect, dist := sphere.RayIntersect(orig, dir)
		if intersect && dist < curDist {
			curColor = sphere.Color
			curDist = dist
		}
	}
	return curColor
}

func main() {
	spheres := []*Sphere{
		{Center: vec3f{-4, -1, -27}, Radius: 4, Color: black},
		{Center: vec3f{-3, 0, -16}, Radius: 2, Color: mint},
		{Center: vec3f{-3, 2, -14}, Radius: 2, Color: slateGray},
	}
	rect := image.Rect(0, 0, totalWidth, totalHeight)
	img := image.NewRGBA(rect)

	fov := 1.0

	for j := 0; j < totalHeight; j++ {
		y := -(2*(float64(j)+0.5)/float64(totalHeight) - 1) * math.Tan(fov/2.0)
		for i := 0; i < totalWidth; i++ {
			x := (2*(float64(i)+0.5)/float64(totalWidth) - 1) * math.Tan(fov/2.0) * totalWidth / float64(totalHeight)
			dir := (&vec3f{x, y, -1}).Normalize()
			img.Set(i, j, castRay(&vec3f{0, 0, 0}, dir, spheres))
		}
	}

	mustWriteToDisk(img, "out.png")
}

func mustWriteToDisk(img image.Image, filename string) {
	// Create output file
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	// Create buffer for file
	buf := bufio.NewWriter(f)
	// Encode img as PNG into buffer
	err = png.Encode(buf, img)
	if err != nil {
		_ = f.Close()
		panic(err)
	}
	// Ensure the entire file is written to disk
	err = buf.Flush()
	if err != nil {
		panic(err)
	}
	// Close the file
	err = f.Close()
	if err != nil {
		panic(err)
	}
}
