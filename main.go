package main

import (
	"flag"
	"io/fs"
	"limn/render"
	"log"
	"os"
)

var rootfs fs.FS
var siteDir string

func init() {
    // var err error
    flag.StringVar(&siteDir, "dir", ".", "project directory")
    flag.Parse()
    rootfs = os.DirFS(".")
    if _, err := fs.ReadDir(rootfs, siteDir); err != nil {
        log.Fatal(err)
    }
}

func main() {
    // log.Print("Printing Site Tree...")
    // builder.PrintSiteTree(root)
    rndr := render.NewRenderer(rootfs, siteDir)
    rndr.BuildSiteModel()
    // var buf *bytes.Buffer
    // rndr.URL(buf, "/notes/note_6.md")
}
