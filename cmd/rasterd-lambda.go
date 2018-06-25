package main

import (
	"github.com/akrylysov/algnhsa"
	"github.com/whosonfirst/go-rasterzen/http"
	"github.com/whosonfirst/go-whosonfirst-cache-s3"
	"log"
	gohttp "net/http"
	"os"
)

func main() {

	// https://docs.aws.amazon.com/lambda/latest/dg/env_variables.html

	s3_dsn := os.Getenv("RASTERZEN_S3_DSN")
	cache_opts := os.Getenv("RASTERZEN_CACHE_OPTIONS")

	opts, err := s3.NewS3CacheOptionsFromString(cache_opts)

	if err != nil {
		log.Fatal(err)
	}

	c, err := s3.NewS3Cache(s3_dsn, opts)

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

	opts := new(algnhsa.Options)
	opts.BinaryContentTypes = []string{ "image/png" }

	algnhsa.ListenAndServe(mux, nil)
}
