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
	ivory = &Material{
		Albedo:           &vec2f{0.6, 0.3},
		DiffuseColor:     &vec3f{0.4, 0.4, 0.3},
		SpecularExponent: 50.0,
	}
	redRubber = &Material{
		Albedo:           &vec2f{0.9, 0.1},
		DiffuseColor:     &vec3f{0.3, 0.1, 0.1},
		SpecularExponent: 10.0,
	}
)

type vec3f struct {
	X float64
	Y float64
	Z float64
}

type vec2f struct {
	X float64
	Y float64
}

// Sphere is represented by a vec3f center and a float64 radius
type Sphere struct {
	Center   *vec3f
	Radius   float64
	Material *Material
}

// Material describes a surface of a body
type Material struct {
	Albedo           *vec2f
	DiffuseColor     *vec3f
	SpecularExponent float64
}

// Light source
type Light struct {
	Position  *vec3f
	Intensity float64
}

func main() {
	spheres := []*Sphere{
		{Center: &vec3f{-3, 0, -16}, Radius: 2, Material: ivory},
		{Center: &vec3f{-1, -1.5, -12}, Radius: 2, Material: redRubber},
		{Center: &vec3f{1.5, -0.5, -18}, Radius: 3, Material: redRubber},
		{Center: &vec3f{7, 5, -18}, Radius: 4, Material: ivory},
	}
	lights := []*Light{
		{Position: &vec3f{-20, 20, 20}, Intensity: 1.5},
		{Position: &vec3f{30, 50, -25}, Intensity: 1.8},
		{Position: &vec3f{30, 20, 30}, Intensity: 1.7},
	}
	rect := image.Rect(0, 0, totalWidth, totalHeight)
	img := image.NewRGBA(rect)

	fov := 1.0

	for j := 0; j < totalHeight; j++ {
		y := -(2*(float64(j)+0.5)/float64(totalHeight) - 1) * math.Tan(fov/2.0)
		for i := 0; i < totalWidth; i++ {
			x := (2*(float64(i)+0.5)/float64(totalWidth) - 1) * math.Tan(fov/2.0) * totalWidth / float64(totalHeight)
			dir := (&vec3f{x, y, -1}).Normalize()
			colorVec := castRay(&vec3f{0, 0, 0}, dir, spheres, lights)
			rgba := color.RGBA{
				R: uint8(math.Min(math.Max(0, colorVec.X), 1) * 255),
				G: uint8(math.Min(math.Max(0, colorVec.Y), 1) * 255),
				B: uint8(math.Min(math.Max(0, colorVec.Z), 1) * 255),
				A: 255,
			}
			img.Set(i, j, rgba)
		}
	}

	mustWriteToDisk(img, "out.png")
}

func sceneIntersect(orig *vec3f, dir *vec3f, spheres []*Sphere) (bool, *vec3f, *vec3f, *Material) {
	curDist := math.MaxFloat64
	var curMaterial *Material
	var hit *vec3f
	var n *vec3f
	for _, sphere := range spheres {
		intersect, dist := sphere.RayIntersect(orig, dir)
		if intersect && dist < curDist {
			curMaterial = sphere.Material
			hit = orig.Add(dir.MultiplyF(dist))
			n = hit.Subtract(sphere.Center).Normalize()
			curDist = dist
		}
	}
	return curDist < 1000, hit, n, curMaterial
}

func castRay(orig *vec3f, dir *vec3f, spheres []*Sphere, lights []*Light) *vec3f {
	intersect, point, n, intersectMaterial := sceneIntersect(orig, dir, spheres)
	if !intersect {
		return &vec3f{55 / 255.0, 176 / 255.0, 202 / 255.0}
	}
	diffuseLightIntensity := 0.0
	specularLightIntensity := 0.0
	for _, light := range lights {
		lightDir := (light.Position.Subtract(point)).Normalize()
		diffuseLightIntensity += light.Intensity * math.Max(0, lightDir.Multiply(n))
		specularLightIntensity += math.Pow(math.Max(0, reflect(lightDir.MultiplyF(-1), n).MultiplyF(-1).Multiply(dir)), intersectMaterial.SpecularExponent) * light.Intensity
	}

	unitVec := &vec3f{1, 1, 1}
	return intersectMaterial.DiffuseColor.MultiplyF(diffuseLightIntensity).MultiplyF(intersectMaterial.Albedo.X).Add(unitVec.MultiplyF(specularLightIntensity).MultiplyF(intersectMaterial.Albedo.Y))
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

func reflect(i *vec3f, n *vec3f) *vec3f {
	return i.Subtract(n.MultiplyF(2.0).MultiplyF(i.Multiply(n)))
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

func (v *vec3f) Add(rhs *vec3f) *vec3f {
	ret := &vec3f{}
	ret.X = v.X + rhs.X
	ret.Y = v.Y + rhs.Y
	ret.Z = v.Z + rhs.Z
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
