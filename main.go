package main

import (
	"bufio"
	"image"
	"image/color"
	"image/png"
	"os"
)

const (
	totalWidth  = 1024
	totalHeight = 768
)

func main() {
	rect := image.Rect(0, 0, totalWidth, totalHeight)
	img := image.NewRGBA(rect)

	for x := 0; x < totalWidth; x++ {
		for y := 0; y < totalHeight; y++ {
			img.Set(x, y, color.RGBA{
				uint8(float64(x) / totalWidth * 255),
				uint8(float64(y) / totalHeight * 255),
				0,
				255,
			})
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
