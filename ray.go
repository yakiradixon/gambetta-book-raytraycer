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

func (t tuple) subtract(t1 tuple) tuple {
	return tuple{t.x - t1.x, t.y - t1.y, t.z - t1.z}
}

func (t tuple) dot(t1 tuple) float64 {
	return (t.x * t1.x) + (t.y * t1.y) + (t.z * t1.z)
}

type color struct {
	r int
	g int
	b int
}

type scene struct {
	v viewport
	spheres []sphere
}

type sphere struct {
	center tuple
	radius float64
	color color
}

type canvas struct {
	width int
	height int
	pixels [][]color
}

func (c canvas) init(w, h int) *canvas {
	pixels := make([][]color, h+1)
	for x, pixel := range pixels {
		pixels[x] = make([]color, w)
		for y, _ := range pixel {
			pixels[x][y] = color{0, 0, 0}
		}
	}
	return &canvas{w, h, pixels}
}

func toPPM(c *canvas) string {
	return  ppmHeader(c) + ppmPixelData(c) + ppmFooter()
}

func ppmHeader(c *canvas) string {
	PPMFlavor := "P3"
	MaxColorValue := 255
	return fmt.Sprintf("%s\n%d %d\n%d\n", PPMFlavor, c.width, c.height, MaxColorValue)
}

func ppmPixelData(c *canvas) string {
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


func writePixelDataFor(psb *strings.Builder, c int, r *string) {
	MinColorValue := 0
	MaxColorValue := 255
	(*psb).WriteString(fmt.Sprintf("%d ", clamp(int(math.Round(float64(c*MaxColorValue))), MinColorValue, MaxColorValue)))
	*r += (*psb).String()
	(*psb).Reset()
}

func clamp(x, min, max int) int {
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
	size int
}

func (c *canvas) toViewport(x int, y int, v viewport) tuple {
	// 1 is the projection plane D, projection plane Z
	return tuple{float64(x*v.size/c.width), float64(y*v.size/c.height), 1}
}

func (c *canvas) putPixel(x int, y int, color color) {
	sx := (c.width / 2) + x
	sy := (c.height/2) - y - 1
	// fmt.Printf("in putPixel:  %s %s", sx, sy)
	c.pixels[sy][sx] = color
}

func traceRay(O tuple, D tuple, tMin float64, tMax float64, sc scene) color {
	closestT := tMax
	closestSphere := sphere{}
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
		nilSphere := sphere{}
		if closestSphere == nilSphere {
			// Return background color
			return color{255, 255, 255}
		}
	}
	return closestSphere.color
}

func intersectRaySphere(O tuple, D tuple, s sphere) (float64, float64) {
	r := s.radius
	CO := O.subtract(s.center)
	a := D.dot(D)
	b := 2*CO.dot(D)
	c := CO.dot(CO) - r * r

	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return math.Inf(0), math.Inf(0)
	}
	t1 := (-float64(b) + math.Sqrt(float64(discriminant))) / float64(2*a)
	t2 := (float64(b) - math.Sqrt(float64(discriminant))) / float64(2*a)
	return t1, t2
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	c := &canvas{}
	c = c.init(600, 600)

	s := scene{}
	s.v = viewport{1}
	s.spheres = []sphere{sphere{tuple{0, -1, 3}, 1, color{255, 0, 0}}, sphere{tuple{2, 0, 4}, 1, color{0, 0, 255}}, sphere{tuple{-2, 0 ,4}, 1, color{0, 255, 0}}}

	O := tuple{0, 0, 0}

	for x := -c.width/2; x < c.width/2; x++ {
		for y := -c.height/2; y < c.height/2; y++ {
			// fmt.Printf("in main loop x: %s y: %s", x, y)
			D := c.toViewport(x, y, s.v)
			fmt.Println(D)
			color := traceRay(O, D, 1, math.Inf(0), s)
			c.putPixel(x, y, color)
		}
	}

	ppm := toPPM(c)

	f, err := os.Create("rs.ppm")
	check(err)
	defer f.Close()
	_, err1 := f.WriteString(ppm)
	check(err1)
	f.Sync()
}
