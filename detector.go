package main

import (
	"io/ioutil"
	"log"

	pigo "github.com/esimov/pigo/core"
	"gocv.io/x/gocv"
)

var (
	cascade          []byte
	puplocCascade    []byte
	faceClassifier   *pigo.Pigo
	puplocClassifier *pigo.PuplocCascade
	flpcs            map[string][]*pigo.FlpCascade
	imgParams        *pigo.ImageParams
	err              error
)

var (
	eyeCascades  = []string{"lp46", "lp44", "lp42", "lp38", "lp312"}
	mouthCascade = []string{"lp93", "lp84", "lp82", "lp81"}
)

func init() {
	// Ensure that the face detection classifier is loaded only once.
	cascade, err = ioutil.ReadFile("cascade/facefinder")
	if err != nil {
		log.Fatalf("Error reading the cascade file: %v", err)
	}
	p := pigo.NewPigo()

	// Unpack the binary file. This will return the number of cascade trees,
	// the tree depth, the threshold and the prediction from tree's leaf nodes.
	faceClassifier, err = p.Unpack(cascade)
	if err != nil {
		log.Fatalf("Error unpacking the cascade file: %s", err)
	}

	// Ensure that we load the pupil localization cascade only once
	puplocCascade, err := ioutil.ReadFile("cascade/puploc")
	if err != nil {
		log.Fatalf("Error reading the puploc cascade file: %s", err)
	}
	puplocClassifier, err = puplocClassifier.UnpackCascade(puplocCascade)
	if err != nil {
		log.Fatalf("Error unpacking the puploc cascade file: %s", err)
	}

	flpcs, err = puplocClassifier.ReadCascadeDir("cascade/lps")
	if err != nil {
		log.Fatalf("Error unpacking the facial landmark detection cascades: %s", err)
	}
}

var (
	gs = gocv.NewMat()
)

type Detection struct {
	pigo.Detection
	kind string
}

func FindFaces(gi gocv.Mat, angle float64) []Detection {
	gocv.CvtColor(gi, &gs, gocv.ColorBGRToGray)

	rows := gs.Rows()
	cols := gs.Cols()
	pixels := gs.ToBytes()

	imgParams = &pigo.ImageParams{
		Pixels: pixels,
		Rows:   rows,
		Cols:   cols,
		Dim:    cols,
	}

	cParams := pigo.CascadeParams{
		MinSize:     rows / 2,
		MaxSize:     rows / 3 * 2,
		ShiftFactor: 0.1,
		ScaleFactor: 1.10,
		ImageParams: *imgParams,
	}

	// Run the classifier over the obtained leaf nodes and return the detection results.
	// The result contains quadruplets representing the row, column, scale and detection score.
	results := faceClassifier.RunCascade(cParams, angle)

	// Calculate the intersection over union (IoU) of two clusters.
	results = faceClassifier.ClusterDetections(results, 0.0)

	var dets []Detection

	for i := 0; i < len(results); i++ {
		// left eye
		puploc := &pigo.Puploc{
			Row:      results[i].Row - int(0.085*float32(results[i].Scale)),
			Col:      results[i].Col - int(0.185*float32(results[i].Scale)),
			Scale:    float32(results[i].Scale) * 0.4,
			Perturbs: 63,
		}
		leftEye := puplocClassifier.RunDetector(*puploc, *imgParams, angle, false)
		if leftEye.Row > 0 && leftEye.Col > 0 {
			dets = append(dets, Detection{pigo.Detection{
				Row:   leftEye.Row,
				Col:   leftEye.Col,
				Scale: int(leftEye.Scale),
				Q:     results[i].Q,
			},
				"LeftEye",
			})
		}

		// right eye
		puploc = &pigo.Puploc{
			Row:      results[i].Row - int(0.085*float32(results[i].Scale)),
			Col:      results[i].Col + int(0.185*float32(results[i].Scale)),
			Scale:    float32(results[i].Scale) * 0.4,
			Perturbs: 63,
		}

		rightEye := puplocClassifier.RunDetector(*puploc, *imgParams, angle, false)
		if rightEye.Row > 0 && rightEye.Col > 0 {
			dets = append(dets, Detection{pigo.Detection{
				Row:   rightEye.Row,
				Col:   rightEye.Col,
				Scale: int(rightEye.Scale),
				Q:     results[i].Q,
			}, "RightEye"})
		}

		// Traverse all the eye cascades and run the detector on each of them.
		for _, eye := range eyeCascades {
			for _, flpc := range flpcs[eye] {
				flp := flpc.GetLandmarkPoint(leftEye, rightEye, *imgParams, puploc.Perturbs, false)
				if flp.Row > 0 && flp.Col > 0 {
					dets = append(dets, Detection{pigo.Detection{
						Row:   flp.Row,
						Col:   flp.Col,
						Scale: int(flp.Scale),
						Q:     results[i].Q,
					}, eye})
				}

				flp = flpc.GetLandmarkPoint(leftEye, rightEye, *imgParams, puploc.Perturbs, true)
				if flp.Row > 0 && flp.Col > 0 {
					if flp.Row > 0 && flp.Col > 0 {
						dets = append(dets, Detection{pigo.Detection{
							Row:   flp.Row,
							Col:   flp.Col,
							Scale: int(flp.Scale),
							Q:     results[i].Q,
						}, eye + "True"})
					}
				}
			}
		}

		// Traverse all the mouth cascades and run the detector on each of them.
		for _, mouth := range mouthCascade {
			for _, flpc := range flpcs[mouth] {
				flp := flpc.GetLandmarkPoint(leftEye, rightEye, *imgParams, puploc.Perturbs, false)
				if flp.Row > 0 && flp.Col > 0 {
					if flp.Row > 0 && flp.Col > 0 {
						dets = append(dets, Detection{pigo.Detection{
							Row:   flp.Row,
							Col:   flp.Col,
							Scale: int(flp.Scale),
							Q:     results[i].Q,
						}, mouth})
					}
				}
			}
		}
		flp := flpcs["lp84"][0].GetLandmarkPoint(leftEye, rightEye, *imgParams, puploc.Perturbs, true)
		if flp.Row > 0 && flp.Col > 0 {
			if flp.Row > 0 && flp.Col > 0 {
				dets = append(dets, Detection{pigo.Detection{
					Row:   flp.Row,
					Col:   flp.Col,
					Scale: int(flp.Scale),
					Q:     results[i].Q,
				}, "lp84"})
			}
		}
	}

	return dets
}
