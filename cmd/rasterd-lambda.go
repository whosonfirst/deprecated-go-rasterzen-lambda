package main

import (
	"flag"
	"github.com/akrylysov/algnhsa"
	"github.com/whosonfirst/go-rasterzen/http"
	"github.com/whosonfirst/go-whosonfirst-cache-s3"
	"log"
	gohttp "net/http"
)

func main() {

	s3_dsn := flag.String("s3-dsn", "", "A valid go-whosonfirst-aws DSN string")
	s3_opts := flag.String("s3-opts", "", "A valid go-whosonfirst-cache-s3 options string")

	flag.Parse()

	opts, err := s3.NewS3CacheOptionsFromString(*s3_opts)

	if err != nil {
		log.Fatal(err)
	}

	c, err := s3.NewS3Cache(*s3_dsn, opts)

	if err != nil {
		log.Fatal(err)
	}

	mux := gohttp.NewServeMux()

	png_handler, err := http.PNGHandler(c)

	if err != nil {
		log.Fatal(err)
	}

	svg_handler, err := http.SVGHandler(c)

	if err != nil {
		log.Fatal(err)
	}

	geojson_handler, err := http.GeoJSONHandler(c)

	if err != nil {
		log.Fatal(err)
	}

	mux.Handle("/geojson/", geojson_handler)
	mux.Handle("/png/", png_handler)
	mux.Handle("/svg/", svg_handler)

	algnhsa.ListenAndServe(mux, nil)
}
