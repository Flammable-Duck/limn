package limn_test

import (
	"io/fs"
	"limn/render"
	"log"
	"os"
	"testing"
)

const SITEDIR = "/example_site"

var rootfs fs.FS

func init() {
    // var err error
    rootfs = os.DirFS(".")
    if _, err := fs.ReadDir(rootfs, SITEDIR); err != nil {
        log.Fatal(err)
    }
}

func TestRenderTree(t *testing.T) {
    // log.Print("Printing Site Tree...")
    // builder.PrintSiteTree(root)
    rndr := render.NewRenderer(rootfs, SITEDIR)
    rndr.Render()
}
