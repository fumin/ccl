package ccl

import (
	"fmt"
)

func Example() {
	data := [][]int{
		{0, 1, 0, 0, 0, 0, 0, 1, 1},
		{1, 1, 1, 0, 1, 0, 0, 1, 0},
		{0, 1, 0, 0, 1, 0, 0, 1, 0},
		{0, 1, 1, 1, 1, 0, 0, 1, 0},
		{0, 0, 0, 1, 0, 0, 0, 1, 0},
		{0, 1, 0, 1, 1, 1, 1, 1, 0},
		{1, 1, 1, 0, 0, 0, 0, 0, 1},
		{1, 0, 1, 1, 1, 0, 0, 0, 1},
		{1, 1, 1, 0, 0, 0, 0, 1, 1},
		{1, 0, 1, 0, 1, 0, 1, 1, 1},
	}
	bmp := newBitmap(data)

	labelSizes := HoshenKopelman(bmp)

	for y := 0; y < len(bmp.labels); y++ {
		for x := 0; x < len(bmp.labels[y]); x++ {
			fmt.Printf(" % d", bmp.labels[y][x])
		}
		fmt.Printf("\n")
	}

	fmt.Println("label sizes:", labelSizes)

	// Output:
	//  -1  0 -1 -1 -1 -1 -1  0  0
	//   0  0  0 -1  0 -1 -1  0 -1
	//  -1  0 -1 -1  0 -1 -1  0 -1
	//  -1  0  0  0  0 -1 -1  0 -1
	//  -1 -1 -1  0 -1 -1 -1  0 -1
	//  -1  1 -1  0  0  0  0  0 -1
	//   1  1  1 -1 -1 -1 -1 -1  2
	//   1 -1  1  1  1 -1 -1 -1  2
	//   1  1  1 -1 -1 -1 -1  2  2
	//   1 -1  1 -1  3 -1  2  2  2
	// label sizes: [23 13 7 1]
}

type bitmap struct {
	data   [][]int
	labels [][]int
	x, y   int
}

func newBitmap(data [][]int) *bitmap {
	bmp := &bitmap{}
	bmp.data = data

	bmp.labels = make([][]int, len(data))
	for y := 0; y < len(data); y++ {
		bmp.labels[y] = make([]int, len(data[y]))
		for x := range bmp.labels[y] {
			bmp.labels[y][x] = NullLabel
		}
	}

	return bmp
}

func (bmp *bitmap) Reset() {
	bmp.x, bmp.y = 0, 0
}

func (bmp *bitmap) Next() bool {
	if !bmp.next() {
		return false
	}
	for bmp.data[bmp.y][bmp.x] == 0 {
		if !bmp.next() {
			return false
		}
	}
	return true
}

func (bmp *bitmap) next() bool {
	bmp.x++
	if bmp.x >= len(bmp.data[bmp.y]) {
		bmp.y++
		bmp.x = 0
	}
	if bmp.y >= len(bmp.data) {
		return false
	}
	return true
}

func (bmp *bitmap) Neighbors() []int {
	x, y := bmp.x, bmp.y
	neighbors := make([]int, 0)
	if y > 0 && bmp.data[y-1][x] != 0 {
		neighbors = append(neighbors, bmp.labels[y-1][x])
	}
	if x > 0 && bmp.data[y][x-1] != 0 {
		neighbors = append(neighbors, bmp.labels[y][x-1])
	}
	return neighbors
}

func (bmp *bitmap) GetLabel() int {
	return bmp.labels[bmp.y][bmp.x]
}

func (bmp *bitmap) SetLabel(label int) {
	bmp.labels[bmp.y][bmp.x] = label
}

func (bmp *bitmap) Size() int {
	return 1
}
