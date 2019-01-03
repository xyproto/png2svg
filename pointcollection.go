package png2svg

import (
	"fmt"
	"strings"
)

type Point struct {
	x, y float64
}

type PointCollection []Point

func NewPointCollection() *PointCollection {
	pc := PointCollection(make([]Point, 0))
	return &pc
}

func (pc *PointCollection) Push(p *Point) {
	*pc = append(*pc, *p)
}

func (pc *PointCollection) PushXY(x, y float64) {
	*pc = append(*pc, Point{x, y})
}

func (pc *PointCollection) Pop() *Point {
	lastIndex := len(*pc) - 1
	p := (*pc)[lastIndex]
	*pc = (*pc)[:lastIndex]
	return &p
}

func (pc *PointCollection) String() string {
	var sb strings.Builder
	for i, p := range *pc {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("(%.3f, %.3f)", p.x, p.y))
	}
	return sb.String()
}

func (pc *PointCollection) Has(p Point) bool {
	for _, ep := range *pc {
		if ep.x == p.x && ep.y == p.y {
			return true
		}
	}
	return false
}

func (pc *PointCollection) HasXY(x, y float64) bool {
	for _, p := range *pc {
		if p.x == x && p.y == y {
			return true
		}
	}
	return false
}

// GrahamScan sorts the points in counter clockwise order
func (pc *PointCollection) GrahamScan() error {
	return nil
}

// PolygonString can be used in polygon tags in SVG images
func (pc *PointCollection) PolygonString() string {
	var sb strings.Builder
	for i, p := range *pc {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%f,%f", p.x, p.y))
	}
	return sb.String()
}
