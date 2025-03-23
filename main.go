package main

import (
	"flag"
	"io/fs"
	"limn/build"
	"limn/render"
	"log"
	"os"
)

var rootfs fs.FS
var siteDir string

func init() {
    flag.StringVar(&siteDir, "dir", ".", "project directory")
    flag.Parse()
    rootfs = os.DirFS(".")
    if _, err := fs.ReadDir(rootfs, siteDir); err != nil {
        log.Fatal(err)
    }
}

func main() {
    rndr := render.NewRenderer(rootfs, siteDir)
    rndr.BuildSiteModel()
    siteFunc := func(p string, c render.Content) error {
        log.Printf("%-40s%s\n", p, c.Template())
        return nil
    }
    rndr.WalkSite(siteFunc)
    build.BuildSite(rndr, siteDir)
    // buf := bytes.NewBuffer([]byte{})
    // err := rndr.URL(buf, "/contact.html")
    // if err != nil { log.Fatal(err) }
    // log.Println(buf.String())
}
