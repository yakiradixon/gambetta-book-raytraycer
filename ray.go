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

func add(t1, t2 tuple) tuple {
	return tuple{t1.x + t2.x, t1.y + t2.y, t1.z + t2.z}
}

func subtract(t1, t2 tuple) tuple {
	return tuple{t1.x - t2.x, t1.y - t2.y, t1.z - t2.z}
}

func multiply(v float64, t1 tuple) tuple {
	return tuple{v * t1.x, v * t1.y, v * t1.z}
}

func divide(v float64, t1 tuple) tuple {
	return tuple{t1.x / v, t1.y / v, t1.z / v}
}

func dot(t1, t2 tuple) float64 {
	return (t1.x * t2.x) + (t1.y * t2.y) + (t1.z * t2.z)
}

func length(t1 tuple) float64 {
	return math.Sqrt(dot(t1, t1))
}

type color struct {
	r float64
	g float64
	b float64
}

func addColor(c1, c2 color) color {
	return color{c1.r + c2.r, c1.g + c2.g, c1.b + c2.b}
}

func multiplyColor(v float64, c color) color {
	return color{v * c.r, v * c.g, v * c.b}
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
	specular float64
	reflective float64
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

func traceRay(origin tuple, direction tuple, tMin float64, tMax float64, recursionDepth int, sc scene) color {
	closestSphere, closest_t := closestIntersection(origin, direction, tMin, tMax, sc)
	nilSphere := sphere{}
	if closestSphere == nilSphere {
		// Return background color
		return color{255, 255, 255}
	}

	point := add(origin, multiply(closest_t, direction))
	normal := subtract(point, closestSphere.center)
	normal = multiply(1.0 / length(normal), normal)

	view := multiply(-1, direction)
	lighting := computeLighting(point, normal, view, closestSphere.specular, sc)
	localColor := multiplyColor(lighting, closestSphere.color)

	if recursionDepth == 0 || closestSphere.reflective <= 0.0 {
		return localColor
	}
	reflectedRay := reflectRay(view, normal)
	reflectedColor := traceRay(point, reflectedRay, 0.001, math.Inf(0), recursionDepth - 1, sc)
	return addColor(multiplyColor(1.0 - closestSphere.reflective, localColor), multiplyColor(closestSphere.reflective, reflectedColor))
}

func closestIntersection(origin tuple, direction tuple, tMin float64, tMax float64, sc scene) (sphere, float64) {
	closest_t := math.Inf(0)
	nilSphere := sphere{}
	closestSphere := nilSphere
	for _, sp := range sc.spheres {
		t1, t2 := intersectRaySphere(origin, direction, sp)
		if t1 < closest_t && tMin < t1 && t1 < tMax {
			closest_t = t1
			closestSphere = sp
		}
		if t2 < closest_t && tMin < t2 && t2 < tMax {
			closest_t = t2
			closestSphere = sp
		}
	}

	return closestSphere, closest_t
}

func intersectRaySphere(origin tuple, direction tuple, s sphere) (float64, float64) {
	oc := subtract(origin, s.center)
	a := dot(direction, direction)
	b := 2 * dot(oc, direction)
	c := dot(oc, oc) - s.radius * s.radius

	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return math.Inf(0), math.Inf(0)
	}
	t1 := (-b + math.Sqrt(discriminant))/(2*a)
	t2 := (-b - math.Sqrt(discriminant))/(2*a)
	return t1, t2
}

func computeLighting(point tuple, normal tuple, view tuple, specular float64, sc scene) float64 {
	i := 0.0

	for _, light := range sc.lights {
		if light.ltype == "ambient" {
			i += light.intensity
		} else {
			var L tuple
			var min, max float64
			min = 0.001
			if light.ltype == "point" {
				L = subtract(light.position, point)
				max = 1
			} else if light.ltype == "directional" {
				L = light.position
				max = math.Inf(0)
			} else {
				panic("scene created with unknown light type")
			}

			shadowSphere, _ := closestIntersection(point, L, min, max, sc)
			nilSphere := sphere{}
			if shadowSphere != nilSphere {
				continue
			}

			if dot(normal, L) > 0 {
				i += light.intensity * dot(normal, L) / (length(normal) * length(L))
			}

			if specular != -1 {
				R := reflectRay(L, normal)
				if dot(R, view) > 0 {
					i += light.intensity * math.Pow(dot(R, view) / (length(R) * length(view)), specular)
				}
			}
		}
	}
	return i
}

func reflectRay(ray, normal tuple) tuple {
	return subtract(multiply(2.0 * dot(normal, ray), normal), ray)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	c := canvas{}
	c = c.init(600, 600)
	s := scene{}
	s.v = viewport{size: 1}

	sp1 := sphere{tuple{0, -1, 3}, 1, color{255, 0, 0}, 500, 0.2}
	sp2 := sphere{tuple{2, 0, 4}, 1, color{0, 0, 255}, 500, 0.3}
	sp3 := sphere{tuple{-2, 0, 4}, 1, color{0, 255, 0}, 10, 0.4}
	sp4 := sphere{tuple{0, -5001, 0}, 5000, color{255, 255, 0}, 1000, 0.5}
	s.spheres = []sphere{sp1, sp2, sp3, sp4}

	l1 := light{"ambient", 0.2, tuple{0, 0, 0}}
	l2 := light{"point", 0.6, tuple{2, 1, 0}}
	l3 := light{"directional", 0.2, tuple{1, 4, 4}}

	s.lights = []light{l1, l2, l3}

	O := tuple{0, 0, 0}

	recursionDepth := 3

	for x := -c.width / 2; x < c.width/2; x++ {
		for y := -c.height / 2; y < c.height/2; y++ {
			D := c.toViewport(x, y, s.v.size)
			color := traceRay(O, D, 1, math.Inf(0), recursionDepth, s)
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
