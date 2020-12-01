package main

import (
	"image"

	"github.com/faiface/pixel"
)

func PointToVec(p image.Point) pixel.Vec {
	return pixel.V(float64(p.X), float64(p.Y))
}

func VecToPoint(v pixel.Vec) image.Point {
	return image.Point{int(v.X), int(v.Y)}
}
