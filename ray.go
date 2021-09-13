package main

import (
	"fmt"
	"math"
	"os"
	"strings"
)

type tuple struct {
	x float64
	y float64
	z float64
}

func (t tuple) add(t1 tuple) tuple {
	return tuple{t.x + t1.x, t.y + t1.y, t.z + t1.z}
}

func (t tuple) subtract(t1 tuple) tuple {
	return tuple{t.x - t1.x, t.y - t1.y, t.z - t1.z}
}

func (t tuple) multiply(v float64) tuple {
	return tuple{t.x * v, t.y * v, t.z * v}
}

func (t tuple) divide(v float64) tuple {
	return tuple{t.x / v, t.y / v, t.z / v}
}

func (t tuple) dot(t1 tuple) float64 {
	return (t.x * t1.x) + (t.y * t1.y) + (t.z * t1.z)
}

func (t tuple) length() float64 {
	return math.Sqrt(t.dot(t))
}

type color struct {
	r float64
	g float64
	b float64
}

func (c color) multiply(v float64) color {
	return color{c.r * v, c.g * v, c.b * v}
}

type scene struct {
	v       viewport
	spheres []sphere
	lights  []light
}

type sphere struct {
	center tuple
	radius float64
	color  color
}

type light struct {
	ltype     string
	intensity float64
	position  tuple
}

type canvas struct {
	width  int
	height int
	pixels [][]color
}

func (c canvas) init(w, h int) canvas {
	pixels := make([][]color, h+1)
	for x, pixel := range pixels {
		pixels[x] = make([]color, w)
		for y, _ := range pixel {
			pixels[x][y] = color{0, 0, 0}
		}
	}
	return canvas{w, h, pixels}
}

func toPPM(c canvas) string {
	return ppmHeader(c) + ppmPixelData(c) + ppmFooter()
}

func ppmHeader(c canvas) string {
	PPMFlavor := "P3"
	MaxColorValue := 255
	return fmt.Sprintf("%s\n%d %d\n%d\n", PPMFlavor, c.width, c.height, MaxColorValue)
}

func ppmPixelData(c canvas) string {
	var sb strings.Builder
	var psb *strings.Builder = &sb
	var rowdata string
	var row *string = &rowdata
	var data string

	var MaxCharacters = 70

	for _, pixel := range c.pixels {
		for _, color := range pixel {
			writePixelDataFor(psb, color.r, row)
			writePixelDataFor(psb, color.g, row)
			writePixelDataFor(psb, color.b, row)
		}
		l := len(rowdata)
		for l > 0 {
			if l > MaxCharacters {
				d, s := split(rowdata)
				data += d + "\n"
				rowdata = s
				l = len(rowdata)
			} else {
				data += strings.TrimSpace(rowdata) + "\n"
				rowdata = ""
				l = len(rowdata)
			}
		}
	}
	return data
}

func ppmFooter() string {
	return "\n"
}

func writePixelDataFor(psb *strings.Builder, c float64, r *string) {
	(*psb).WriteString(fmt.Sprintf("%d ", int(math.Round(c))))
	*r += (*psb).String()
	(*psb).Reset()
}

func (c color) clamp() color {
	MinColorValue := 0.0
	MaxColorValue := 255.0
	return color{clamp(c.r, MinColorValue, MaxColorValue), clamp(c.g, MinColorValue, MaxColorValue), clamp(c.b, MinColorValue, MaxColorValue)}
}

func clamp(x, min, max float64) float64 {
	if x < min {
		x = min
	} else if x > max {
		x = max
	}
	return x
}

func split(s string) (string, string) {
	i := strings.LastIndex(s[:70], " ")
	return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i:])
}

type viewport struct {
	size float64
}

func (c canvas) toViewport(x int, y int, vsize float64) tuple {
	projectionPlane := 1.0
	return tuple{float64(x) * vsize / float64(c.width), float64(y) * vsize / float64(c.height), projectionPlane}
}

func (c canvas) putPixel(x int, y int, color color) {
	sx := (c.width / 2) + x
	sy := (c.height / 2) - y - 1
	c.pixels[sy][sx] = color
}

func traceRay(O tuple, D tuple, tMin float64, tMax float64, sc scene) color {
	closestT := math.Inf(0)
	nilSphere := sphere{}
	closestSphere := nilSphere
	for _, s := range sc.spheres {
		t1, t2 := intersectRaySphere(O, D, s)
		if t1 < closestT && tMin < t1 && t1 < tMax {
			closestT = t1
			closestSphere = s
		}
		if t2 < closestT && tMin < t2 && t2 < tMax {
			closestT = t2
			closestSphere = s
		}
	}

	if closestSphere == nilSphere {
		// Return background color
		return color{255, 255, 255}
	}

	P := O.add(D.multiply(closestT))
	N := P.subtract(closestSphere.center)
	N = N.multiply(1.0 / N.length())
	return closestSphere.color.multiply(computeLighting(P, N, sc))
}

func intersectRaySphere(O tuple, D tuple, s sphere) (float64, float64) {
	r := s.radius
	CO := O.subtract(s.center)
	a := D.dot(D)
	b := 2 * CO.dot(D)
	c := CO.dot(CO) - r*r

	discriminant := float64(b*b - 4*a*c)
	if discriminant < 0 {
		return math.Inf(0), math.Inf(0)
	}
	t1 := float64(-b) + math.Sqrt(discriminant)/float64(2*a)
	t2 := float64(-b) - math.Sqrt(discriminant)/float64(2*a)
	return t1, t2
}

func computeLighting(P tuple, N tuple, sc scene) float64 {
	i := 0.0
	for _, light := range sc.lights {
		if light.ltype == "ambient" {
			i += light.intensity
		} else {
			var L tuple
			if light.ltype == "point" {
				L = light.position.subtract(P)
			} else if light.ltype == "directional" {
				L = light.position
			} else {
				panic("scene created with unknown light type")
			}

			if N.dot(L) > 0 {
				i += light.intensity * N.dot(L) / (N.length() * L.length())
			}
		}
	}
	return i
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	c := canvas{}
	c = c.init(200, 200)

	s := scene{}
	s.v = viewport{size: 1}
	s.spheres = []sphere{sphere{tuple{0, -1, 3}, 1, color{255, 0, 0}}, sphere{tuple{2, 0, 4}, 1, color{0, 0, 255}}, sphere{tuple{-2, 0, 4}, 1, color{0, 255, 0}}, sphere{tuple{0, -5001, 0}, 5000, color{255, 255, 0}}}
	s.lights = []light{light{"ambient", 0.2, tuple{0, 0, 0}}, light{"point", 0.6, tuple{2, 1, 0}}, light{"directional", 0.2, tuple{1, 4, 4}}}

	O := tuple{0, 0, 0}

	for x := -c.width / 2; x < c.width/2; x++ {
		for y := -c.height / 2; y < c.height/2; y++ {
			D := c.toViewport(x, y, s.v.size)
			color := traceRay(O, D, 1, math.Inf(0), s)
			c.putPixel(x, y, color.clamp())
		}
	}

	ppm := toPPM(c)

	f, err := os.Create("latest_render.ppm")
	check(err)
	defer f.Close()
	_, err1 := f.WriteString(ppm)
	check(err1)
	f.Sync()
}
