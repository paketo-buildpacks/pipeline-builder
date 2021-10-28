package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"

	_ "github.com/paketo-buildpacks/pipeline-builder/octo/statik"
	"github.com/rakyll/statik/fs"
)

// Prints the contents of all the statik files as well as their names and modification times
//
// This can be useful to see what has changed between two versions.
//
// For example:
//   1. `git co <branch>`
//   2. `go run cmd/statik/main.go > old.txt`
//   3. `git co <other-branch>`
//   4. `go run cmd/statik/main.go > new.txt`
//   5. `diff -u old.txt new.txt`
func main() {
	fmt.Println("Contents of statik files")
	fmt.Println()

	statik, err := fs.New()
	if err != nil {
		fmt.Println(err)
		panic("unable to open static fs")
	}

	maxLen := 0
	files := []string{}
	data := map[string]time.Time{}
	err = fs.Walk(statik, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("unable to walk\n%w", err)
		}

		_, ok := data[path]
		if ok {
			return fmt.Errorf("unable to continue, duplicate file %s", path)
		}

		if len(path) > maxLen {
			maxLen = len(path)
		}
		files = append(files, path)
		data[path] = info.ModTime()

		fmt.Println()
		fmt.Println("#########################################", path)

		fp, err := statik.Open(path)
		if err != nil {
			return fmt.Errorf("unable to open file %s", path)
		}

		b, err := ioutil.ReadAll(fp)
		if err != nil {
			return fmt.Errorf("unable to read file %s", path)
		}

		fmt.Println(string(b))

		return nil
	})
	if err != nil {
		fmt.Println(err)
		panic("unable to walk statik fs")
	}
	fmt.Println()
	fmt.Println("Listing files & modification times stored in statik")

	sort.Strings(files)
	for _, path := range files {
		gap := ""
		for i := 0; i < maxLen-len(path); i++ {
			gap += " "
		}
		fmt.Printf("%s  %s  %s\n", path, gap, data[path])
	}
}
