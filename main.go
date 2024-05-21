package main

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/aztec"
	"github.com/boombuler/barcode/code128"
	"github.com/boombuler/barcode/code39"
	"github.com/boombuler/barcode/code93"
	"github.com/boombuler/barcode/ean"
	"github.com/boombuler/barcode/qr"

	"github.com/julienschmidt/httprouter"
)

func logError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func Index(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(200)
	fmt.Fprintf(w, "Barcode Server\n"+
		"A library to generate barcodes & qrcodes using http requests\n\n"+
		"GET /generate/<mode>/<size>?data=<data>\n\n"+
		"mode  - barcode mode (one of: ean, code39, code93, code128, aztec, qr)\n"+
		"data  - data to encode\n"+
		"scale - output image scale")
}

func Health(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	fmt.Fprintf(w, "OK")
}

func convertTo24BitDepth(img image.Image) *image.RGBA {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return rgba
}

func Generate(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	query := req.URL.Query()
	value := query.Get("data")
	size := ps.ByName("size")
	mode := ps.ByName("name")

	var code barcode.Barcode
	var err error

	switch mode {
	case "ean":
		code, err = ean.Encode(value)
	case "code39":
		code, err = code39.Encode(value, true, true)
	case "code93":
		code, err = code93.Encode(value, true, true)
	case "code128":
		code, err = code128.Encode(value)
	case "aztec":
		code, err = aztec.Encode([]byte(value), aztec.DEFAULT_EC_PERCENT, aztec.DEFAULT_LAYERS)
	case "qr":
		code, err = qr.Encode(value, qr.M, qr.Auto)
	default:
		w.WriteHeader(404)
		fmt.Fprintf(w, "404 page not found")
		return
	}

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "%s", err)
		return
	}

	sizeParts := strings.Split(size, "x")
	if len(sizeParts) != 2 {
		w.WriteHeader(400)
		fmt.Fprintf(w, "invalid size")
		return
	}

	width, err := strconv.Atoi(sizeParts[0])
	height, err := strconv.Atoi(sizeParts[1])
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "invalid size parameters")
		return
	}

	code, err = barcode.Scale(code, width, height)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "%s", err)
		return
	}

	// Convert to 24-bit depth
	img := convertTo24BitDepth(code)

	w.Header().Add("content-type", "image/png")
	w.WriteHeader(200)
	logError(png.Encode(w, img))
}

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/health", Health)
	router.GET("/generate/:name/:size", Generate)

	port := os.Getenv("PORT")

	if len(port) == 0 {
		port = "8080"
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	logError(http.ListenAndServe(port, router))
}
