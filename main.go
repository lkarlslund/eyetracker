package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/lxn/win"
	"gocv.io/x/gocv"
)

func main() {
	hDC := win.GetDC(0)
	defer win.ReleaseDC(0, hDC)
	widthpx := int(win.GetDeviceCaps(hDC, win.HORZRES))
	heightpx := int(win.GetDeviceCaps(hDC, win.VERTRES))
	widthmm := int(win.GetDeviceCaps(hDC, win.HORZSIZE))
	heightmm := int(win.GetDeviceCaps(hDC, win.VERTSIZE))

	fmt.Printf("Monitor detected as %v by %v pixels (%v by %v mm)\n", widthpx, heightpx, widthmm, heightmm)

	webcam, _ := gocv.OpenVideoCaptureWithAPI(0, gocv.VideoCaptureDshow)
	/*
		for x := gocv.VideoCaptureProperties(0); x < 38; x++ {
			res := webcam.Get(x)
			if x == gocv.VideoCaptureFOURCC {
				fmt.Printf("%v: %v\n", x, webcam.CodecString())
			} else {
				fmt.Printf("%v: %v\n", x, res)
			}
		}
	*/
	webcam.Set(gocv.VideoCaptureFOURCC, webcam.ToCodec("MJPG"))
	webcam.Set(gocv.VideoCaptureFrameWidth, 1280)
	webcam.Set(gocv.VideoCaptureFrameHeight, 720)
	webcam.Set(gocv.VideoCaptureFPS, 60)
	/*
		for x := gocv.VideoCaptureProperties(0); x < 38; x++ {
			res := webcam.Get(x)
			if x == gocv.VideoCaptureFOURCC {
				fmt.Printf("%v: %v\n", x, webcam.CodecString())
			} else {
				fmt.Printf("%v: %v\n", x, res)
			}
		}
	*/
	window := gocv.NewWindow("Webcam face tracking demo using Pigo")

	defer window.Close()
	defer webcam.Close()

	img := gocv.NewMat()
	blue := color.RGBA{0, 0, 255, 0}
	red := color.RGBA{255, 0, 0, 0}
	white := color.RGBA{255, 255, 255, 0}
	transparent := color.RGBA{255, 255, 255, 128}

	var facerotationangle float64
	// var distancefromscreen = 40
	// var headrotaztionangle int

	detect := true
	text := false
	lasttime := time.Now()
loop:
	for {
		webcam.Read(&img)

		if detect {
			// Run detections on image
			var leftpupil, leo, lei, rightpupil, reo, rei, nose image.Point

			analyzeangle := -facerotationangle / (2 * math.Phi)
			if analyzeangle < 0 {
				analyzeangle = 1 + analyzeangle
			}
			// analyzeangle = 0
			results := FindFaces(img, 0)

			for _, result := range results {
				// fmt.Printf("%v\n", result)
				// gocv.Rectangle(&img, image.Rect(result.Col-result.Scale/2, result.Row-result.Scale/2, result.Col+result.Scale/2, result.Row+result.Scale/2), blue, 2)
				gocv.Circle(&img, image.Pt(result.Col, result.Row), 2, blue, 1)
				if text {
					gocv.PutText(&img, result.kind, image.Point{result.Col + 12, result.Row + 12}, gocv.FontHersheyComplex, 0.8, white, 2)
				}

				switch result.kind {
				case "LeftEye":
					leftpupil = image.Point{result.Col, result.Row}
				case "RightEye":
					rightpupil = image.Point{result.Col, result.Row}
				case "lp93":
					nose = image.Point{result.Col, result.Row}
				case "lp312":
					leo = image.Point{result.Col, result.Row}
				case "lp38":
					lei = image.Point{result.Col, result.Row}
				case "lp312True":
					reo = image.Point{result.Col, result.Row}
				case "lp38True":
					rei = image.Point{result.Col, result.Row}
				}
			}
			// ang := math.Atan2(float64(righteye.Y-lefteye.Y), float64(righteye.X-lefteye.X))

			lefteyecenter := pixel.L(PointToVec(leo), PointToVec(lei)).Center()
			righteyecenter := pixel.L(PointToVec(reo), PointToVec(rei)).Center()
			lefteyevec := lefteyecenter.Sub(PointToVec(leftpupil))
			righteyevec := righteyecenter.Sub(PointToVec(rightpupil))

			lefteyeouter := PointToVec(leo)
			righteyeouter := PointToVec(reo)

			nosevec := PointToVec(nose)
			eyeline := pixel.L(PointToVec(leo), PointToVec(reo))
			noseintersect := eyeline.Closest(nosevec)

			facerotationangle = righteyeouter.Sub(lefteyeouter).Angle()
			headdirection := (eyeline.Len()/2 - pixel.L(lefteyeouter, noseintersect).Len()) / (eyeline.Len() / 2)

			gocv.Circle(&img, leftpupil, 6, red, 1)
			gocv.Circle(&img, rightpupil, 6, red, 1)

			gocv.Line(&img, leo, reo, blue, 2)
			gocv.Line(&img, nose, VecToPoint(noseintersect), blue, 2)

			// pixel.V(0)
			estimatedspot := image.Pt(0, 0)

			gocv.Circle(&img, estimatedspot, 20, transparent, 10)

			gocv.PutText(&img, fmt.Sprintf("data %.2f %.2f %.2f %v %v", facerotationangle, analyzeangle, headdirection, lefteyevec, righteyevec), image.Point{0, 120}, gocv.FontHersheyComplex, 1, blue, 2)
		}

		thistime := time.Now()
		fps := int(time.Second / thistime.Sub(lasttime))
		lasttime = thistime

		gocv.PutText(&img, fmt.Sprintf("%v fps", fps), image.Point{0, 20}, gocv.FontHersheyComplex, 1, blue, 2)

		window.IMShow(img)
		key := window.WaitKey(5)
		switch key {
		case 32: // space
			detect = !detect
		case 120: // x
			text = !text
		case 27: // esc
			break loop
		case -1:
		default:
			fmt.Printf("Key pressed: %v\n", key)
		}
	}
}
