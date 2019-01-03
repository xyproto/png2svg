package png2svg

import (
	"errors"
	"fmt"
	"math"
	"sort"
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

// Return the point with the smallest Y value, or an error if there are no points.
// Also place the point with the smallest Y value in the first position of the list.
func (pc *PointCollection) BottomPointSwap() (Point, error) {
	if len(*pc) == 0 {
		return Point{}, errors.New("no points")
	}
	if len(*pc) == 1 {
		return (*pc)[0], nil
	}
	minYindex := 0
	minY := (*pc)[minYindex].y
	for i, p := range (*pc)[1:] {
		if p.y < minY {
			minY = p.y
			minYindex = i + 1 // since i loops from 0 when the real index is 1
		}
	}
	if minYindex == 0 {
		// Did not found a smaller Y value than the first one had, return that one
		return (*pc)[0], nil
	}
	// Swap the point with the smallest Y value with the one in position 0
	(*pc)[0], (*pc)[minYindex] = (*pc)[minYindex], (*pc)[0]
	return (*pc)[0], nil
}

// AngleToBottom returns the angle from the bottom-most point to the point
// indicated by the given index. The angle is like it appears on a unit circle, in radians.
func (pc *PointCollection) AngleToBottom(i int) float64 {
	a := (*pc)[0]
	b := (*pc)[i]
	// Return the angle from point a to point b, as indicated by the unit circle
	return math.Atan2(b.y-a.y, b.x-a.x)
}

// DistanceToBottom returns the distance to the bottom-most point, given an index
func (pc *PointCollection) DistanceToBottom(i int) float64 {
	a := (*pc)[0]
	b := (*pc)[i]

	x := (b.x - a.x)
	y := (b.y - a.y)
	// Return the distance from a to b
	return math.Sqrt(x*x + y*y)
}

func (pc PointCollection) Len() int      { return len(pc) }
func (pc PointCollection) Swap(i, j int) { pc[i], pc[j] = pc[j], pc[i] }

func (pc PointCollection) Less(i, j int) bool {
	ppc := &pc
	if ppc.AngleToBottom(i) < ppc.AngleToBottom(j) {
		return true
	}
	if (ppc.AngleToBottom(i) == ppc.AngleToBottom(j)) && (ppc.DistanceToBottom(i) < ppc.DistanceToBottom(j)) {
		return true
	}
	return false
}

// GrahamScan sorts the points in counter clockwise order
func (pc *PointCollection) GrahamScan() error {
	// https://www.geeksforgeeks.org/convex-hull-set-2-graham-scan/

	// 1. Find the bottom-most point
	smallestYpoint, err := pc.BottomPointSwap()
	if err != nil {
		return err
	}

	fmt.Println("FOUND BOTTOM MOST POINT", smallestYpoint)

	fmt.Println("UNSORTED:")

	for i := 1; i < len(*pc); i++ {
		fmt.Println("\tANGLE", pc.AngleToBottom(i), "FOR", (*pc)[i], "DISTANCE", pc.DistanceToBottom(i))
	}

	sort.Sort(*pc)

	fmt.Println("NOW SORTED!")

	for i := 0; i < len(*pc); i++ {
		fmt.Println("\tANGLE", pc.AngleToBottom(i), "FOR", (*pc)[i], "DISTANCE", pc.DistanceToBottom(i))
	}

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
