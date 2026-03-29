package main

import (
	"image"
	"image/color"
	"image/gif"
	"math"
	"math/cmplx"
	"os"
	"sync"
)

const (
	width     = 800  // для GIF лучше не больше 600-800
	height    = 800
	frames    = 1200  // количество кадров
	delay     = 2    // задержка между кадрами (10 = 100ms)
	maxIter0  = 300  // базовое количество итераций
	zoomSpeed = 1.01 // скорость приближения
)

func mandelbrot(c complex128, maxIter int) float64 {
	z := complex(0.0, 0.0)
	for i := 0; i < maxIter; i++ {
		z = z*z + c
		abs := cmplx.Abs(z)
		if abs > 2 {
			return float64(i) + 1 - math.Log2(math.Log2(abs))
		}
	}
	return float64(maxIter)
}

func getColor(iter float64, maxIter int) color.RGBA {
	if iter >= float64(maxIter) {
		return color.RGBA{0, 0, 0, 255}
	}
	t := iter / float64(maxIter)
	r := uint8(127 * (1 + math.Sin(6.28*(t+0.0))))
	g := uint8(127 * (1 + math.Sin(6.28*(t+0.33))))
	b := uint8(127 * (1 + math.Sin(6.28*(t+0.66))))
	return color.RGBA{r, g, b, 255}
}

// рендер кадра с параллельной обработкой строк
func renderFrame(width, height int, zoom float64, centerX, centerY float64) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, width, height), nil)
	maxIter := int(maxIter0 + math.Log(zoom)*80)

	// создаём палитру для GIF
	palette := []color.Color{}
	for i := 0; i < 256; i++ {
		palette = append(palette, color.RGBA{uint8(i), uint8((i*7)%256), uint8((i*5)%256), 255})
	}
	img.Palette = palette

	var wg sync.WaitGroup
	for py := 0; py < height; py++ {
		wg.Add(1)
		go func(py int) {
			defer wg.Done()
			for px := 0; px < width; px++ {
				x := centerX + (float64(px)/float64(width)-0.5)*(3.0/zoom)
				y := centerY + (float64(py)/float64(height)-0.5)*(3.0/zoom)
				c := complex(x, y)
				iter := mandelbrot(c, maxIter)
				col := getColor(iter, maxIter)
				// находим ближайший цвет в палитре
				img.SetColorIndex(px, py, uint8(col.R))
			}
		}(py)
	}
	wg.Wait()
	return img
}

func main() {
	centerX := -0.743643887037151
	centerY := 0.13182590420533
	zoom := 2.0

	anim := &gif.GIF{}
	for i := 0; i < frames; i++ {
		img := renderFrame(width, height, zoom, centerX, centerY)
		anim.Image = append(anim.Image, img)
		anim.Delay = append(anim.Delay, delay)

		zoom *= zoomSpeed
	}

	f, err := os.Create("mandelbrot.gif")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	gif.EncodeAll(f, anim)
}
