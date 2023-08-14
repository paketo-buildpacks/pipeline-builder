package main

import (
	"fmt"
	"log"
	"os"

	buildpack "github.com/paketo-buildpacks/pipeline-builder/v2/buildpack"
	"github.com/spf13/pflag"
)

func main() {
	flagSet := pflag.NewFlagSet("update-buildpack-image-id", pflag.ExitOnError)
	image := flagSet.String("image", "", "The exisiting buildpack image")
	newImage := flagSet.String("new-image", "", "The new buildpack image")
	id := flagSet.String("id", "", "The new id of the buildpack")
	version := flagSet.String("version", "", "The new version of the buildpack")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		log.Fatal(fmt.Errorf("unable to parse flags\n%w", err))
	}

	if *image == "" {
		log.Fatal("--image is required")
	}

	if *newImage == "" {
		log.Fatal("--new-image is required")
	}

	if *id == "" {
		log.Fatal("--id is required")
	}

	if *version == "" {
		log.Fatal("--version is required")
	}

	rename, err := buildpack.Rename(*image, *newImage, *id, *version)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rename)
}
