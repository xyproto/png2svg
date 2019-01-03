package png2svg

import (
	"fmt"
)

func ExamplePointCollection() {
	pc := NewPointCollection()

	// Create some sort of "circle"
	pc.Push(&Point{0, 1})
	pc.Push(&Point{2, 3})
	pc.Push(&Point{4, 5})
	pc.Push(&Point{2, 7})
	pc.Push(&Point{0, 10})
	pc.Push(&Point{-2, 7})
	pc.Push(&Point{-4, 5})
	pc.Push(&Point{-2, 3})

	fmt.Println(pc)

	// Output:
	// (0.000, 1.000), (2.000, 3.000), (4.000, 5.000), (2.000, 7.000), (0.000, 10.000), (-2.000, 7.000), (-4.000, 5.000), (-2.000, 3.000)
}

func ExamplePointCollection_GrahamScan() {
	pc := NewPointCollection()

	// Create some sort of "circle", but smallest Y must not be first
	pc.Push(&Point{2, 3})
	pc.Push(&Point{2, 7})
	pc.Push(&Point{-2, 7})
	pc.Push(&Point{-4, 5})
	pc.Push(&Point{-2, 3})
	pc.Push(&Point{4, 5})
	pc.Push(&Point{0, 1})
	pc.Push(&Point{0, 10})

	fmt.Println(pc)
	pc.GrahamScan()
	fmt.Println(pc)

	// Output:
	// (2.000, 3.000), (4.000, 5.000), (2.000, 7.000), (0.000, 10.000), (-2.000, 7.000), (-4.000, 5.000), (-2.000, 3.000), (0.000, 1.000)
	// FOUND BOTTOM MOST POINT {0 1}
	// (0.000, 1.000), (4.000, 5.000), (2.000, 7.000), (0.000, 10.000), (-2.000, 7.000), (-4.000, 5.000), (-2.000, 3.000), (2.000, 3.000)

}
