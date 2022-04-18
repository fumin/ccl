package ccl

import (
	"image"
	"image/color"
	"log"
	"sort"
	"unsafe"
)

const (
	// EmptyBlob refers to points not belonging to an input image.
	EmptyBlob = -1
)

// BlobFromColor returns a blob identifier from its color representation.
func BlobFromColor(r, g, b, a byte) int {
	i := int(r)*256*256*256 + int(g)*256*256 + int(b)*256 + int(a)
	i += EmptyBlob
	return i
}

// ColorFromBlob returns the color representation of a blob.
func ColorFromBlob(v int) color.NRGBA {
	var c color.NRGBA
	if v == EmptyBlob {
		return c
	}

	i := v - EmptyBlob
	c.R = uint8(i / (256 * 256 * 256))
	residue := i % (256 * 256 * 256)
	c.G = uint8(residue / (256 * 256))
	residue = residue % (256 * 256)
	c.B = uint8(residue / 256)
	c.A = uint8(residue % 256)
	return c
}

// Blob is a blob in connected components labeling
type Blob struct {
	// ID is the identifier of this blob
	ID int

	// Size is the size of this blob.
	Size int
}

// CCLImage performs connected components labeling on an image.NRGBA.
// The image is modified in-place, and the blob each pixel belongs to
// can be inferred from BlobFromColor.
func CCLImage(img *image.NRGBA) []Blob {
	if img.Bounds().Min.X != 0 || img.Bounds().Min.Y != 0 {
		log.Fatalf("%+v", img.Bounds())
	}
	imgW, imgH := img.Bounds().Max.X, img.Bounds().Max.Y
	labels := img

	equivalence := make([]int, 0)
	find := func(label int) int {
		parent := label
		for equivalence[parent] != parent {
			parent = equivalence[parent]
		}

		// Squash equivalence links to speed up future lookups.
		for equivalence[label] != label {
			tmp := equivalence[label]
			equivalence[label] = parent
			label = tmp
		}

		return parent
	}

	largestLabel := 0
	pSize := unsafe.Sizeof(img.Pix[0])
	imgPix := unsafe.Pointer(&img.Pix[0])
	labelsPix := unsafe.Pointer(&labels.Pix[0])
	setLabel := func(label int) {
		c := ColorFromBlob(label)
		*(*byte)(labelsPix), *(*byte)(unsafe.Add(labelsPix, 1)), *(*byte)(unsafe.Add(labelsPix, 2)), *(*byte)(unsafe.Add(labelsPix, 3)) = c.R, c.G, c.B, c.A
	}
	aboveBuf := make([]int, imgW)
	cur, above, left := EmptyBlob, EmptyBlob, EmptyBlob
	counts := make([]Blob, 0)
	for y := 0; y < imgH; y++ {
		for x := 0; x < imgW; x++ {
			if *(*byte)(imgPix) == 0 && *(*byte)(unsafe.Add(imgPix, 1)) == 0 && *(*byte)(unsafe.Add(imgPix, 2)) == 0 {
				cur = EmptyBlob
			} else {
				if y == 0 {
					above = EmptyBlob
				} else {
					above = aboveBuf[x]
				}
				if x == 0 {
					left = EmptyBlob
				}
				switch {
				case above == EmptyBlob && left == EmptyBlob:
					cur = largestLabel
					largestLabel++
					// We actually mean equivalence[cur] = cur
					equivalence = append(equivalence, cur)
					counts = append(counts, Blob{ID: cur})
				case above != EmptyBlob && left == EmptyBlob:
					cur = find(above)
				case above == EmptyBlob && left != EmptyBlob:
					cur = find(left)
				default:
					equivalence[find(left)] = find(above)
					cur = find(left)
				}
				counts[cur].Size++
			}

			setLabel(cur)
			aboveBuf[x] = cur
			left = cur
			imgPix = unsafe.Add(imgPix, 4*pSize)
			labelsPix = unsafe.Add(labelsPix, 4*pSize)
		}
		// if y%1000 == 0 || y == imgH-1 {
		// 	log.Printf("1st pass y %d/%d, labels %d", y, imgH, largestLabel)
		// }
	}

	for _, blob := range counts {
		other := find(blob.ID)
		if other == blob.ID {
			continue
		}
		counts[other].Size += blob.Size
		counts[blob.ID].Size = 0
	}
	sort.Slice(counts, func(i, j int) bool { return counts[i].Size > counts[j].Size })
	mapping := make([]int, len(counts))
	numNotEmpty := 0
	for i, b := range counts {
		if b.Size == 0 {
			mapping[b.ID] = mapping[find(b.ID)]
		} else {
			mapping[b.ID] = i
			counts[i].ID = i
			numNotEmpty++
		}
	}
	counts = counts[:numNotEmpty]

	labelsPix = unsafe.Pointer(&labels.Pix[0])
	for y := 0; y < imgH; y++ {
		for x := 0; x < imgW; x++ {
			blob := BlobFromColor(*(*byte)(labelsPix), *(*byte)(unsafe.Add(labelsPix, 1)), *(*byte)(unsafe.Add(labelsPix, 2)), *(*byte)(unsafe.Add(labelsPix, 3)))
			if blob != EmptyBlob {
				setLabel(mapping[blob])
			}

			labelsPix = unsafe.Add(labelsPix, 4)
		}
		// if y%1000 == 0 || y == imgH-1 {
		// 	log.Printf("2nd pass y %d/%d", y, imgH)
		// }
	}

	return counts
}

// CollectBlobs collects the blobs from an image returned by CCLImage.
func CollectBlobs(img *image.NRGBA) []Blob {
	blobs := make([]Blob, 0)
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			i := img.PixOffset(x, y)
			blobID := BlobFromColor(img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3])
			if blobID == EmptyBlob {
				continue
			}

			if blobID >= len(blobs) {
				short := blobID + 1 - len(blobs)
				blobs = append(blobs, make([]Blob, short)...)
			}
			blobs[blobID].ID = blobID
			blobs[blobID].Size++
		}

		// if y%1000 == 0 || y == img.Rect.Max.Y-1 {
		// 	log.Printf("CollectBlobs row %d/%d, %d blobs", y, img.Bounds().Max.Y, len(blobs))
		// }
	}
	return blobs
}

var palette = []color.NRGBA{
	color.NRGBA{R: 255, A: 255},
	color.NRGBA{G: 255, A: 255},
	color.NRGBA{B: 255, A: 255},
	color.NRGBA{R: 255, G: 255, A: 255},
	color.NRGBA{R: 255, B: 255, A: 255},
}

// Visualize visualizes the largest blobs in an image.
func Visualize(src *image.NRGBA) {
	if src.Bounds().Min.X != 0 || src.Bounds().Min.Y != 0 {
		log.Fatalf("%+v", src.Bounds())
	}
	imgW, imgH := src.Bounds().Max.X, src.Bounds().Max.Y
	srcPix := unsafe.Pointer(&src.Pix[0])
	dst := src
	dstPix := unsafe.Pointer(&dst.Pix[0])
	for y := 0; y < imgH; y++ {
		for x := 0; x < imgW; x++ {
			blob := BlobFromColor(*(*byte)(srcPix), *(*byte)(unsafe.Add(srcPix, 1)), *(*byte)(unsafe.Add(srcPix, 2)), *(*byte)(unsafe.Add(srcPix, 3)))
			switch {
			case blob == EmptyBlob:
			case blob < len(palette):
				*(*byte)(dstPix), *(*byte)(unsafe.Add(dstPix, 1)), *(*byte)(unsafe.Add(dstPix, 2)), *(*byte)(unsafe.Add(dstPix, 3)) = palette[blob].R, palette[blob].G, palette[blob].B, palette[blob].A
			default:
				*(*byte)(dstPix), *(*byte)(unsafe.Add(dstPix, 1)), *(*byte)(unsafe.Add(dstPix, 2)), *(*byte)(unsafe.Add(dstPix, 3)) = 255, 255, 255, 255
			}

			srcPix = unsafe.Add(srcPix, 4)
			dstPix = unsafe.Add(dstPix, 4)
		}
		// if y%1000 == 0 || y == imgH-1 {
		// 	log.Printf("Visualize %d/%d", y, imgH)
		// }
	}
}
