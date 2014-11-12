package main

import (
	// "bufio"
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	// "image/draw"
	// "image/jpeg"
	"image/png"
	// "io/ioutil"
	// "io"
	// "math"
	"os"
	// "os/exec"
	"time"
)

var fname string
var dim = int(1000)

func init() {
	flag.StringVar(&fname, "f", "", "byte map")
}

func main() {
	flag.Parse()
	if fname != "" {
		for {
			outfname := "./tmp.png"

			// randBytes, err := ioutil.ReadFile(fname)
			f, err := os.Open(fname)
			if err != nil {
				panic(1)
			}
			defer f.Close()

			offset := int64(0)
			oldBytes := make([]byte, 4*2*dim*dim)
			randBytes := make([]byte, 4*2*dim*dim)

			for {
				n, err := f.ReadAt(randBytes, offset)
				if err != nil || n != 4*2*dim*dim {
					break
				}

				breite := 2 * dim
				hoehe := dim // (len(randBytes) >> 2) / breite

				fmt.Println(breite, hoehe, breite*hoehe, len(randBytes), len(randBytes)>>2)

				dst := image.NewNRGBA(image.Rect(0, 0, breite, hoehe))

				bytes2Pixel(dst, &randBytes)
				err = writeFile(dst, outfname)
				if err != nil {
					break
				}
				os.Rename(outfname, "randPad.png")
				fmt.Println(offset, bytes.Compare(oldBytes, randBytes))
				copy(oldBytes, randBytes)

				offset += int64(4 * 2 * dim * dim)
				time.Sleep(time.Second / 33)

			}

		}
	}
}

func bytes2Pixel(dst *image.NRGBA, randBytes *[]byte) {
	b := dst.Bounds()
	idx := 0
	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			ncol := color.NRGBA{uint8((*randBytes)[idx]), uint8((*randBytes)[idx+1]), uint8((*randBytes)[idx+2]), uint8((*randBytes)[idx+3])}
			dst.Set(x, y, ncol)
			idx += 4
		}
	}
}

func writeFile(pic *image.NRGBA, fname string) error {
	// Write to a .png file

	f, err := os.Create(fname)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()

	err = png.Encode(f, pic)
	if err != nil {
		fmt.Println(err)
		return err
	}

	//fmt.Println("RandomPattern", fname)
	return nil
}
