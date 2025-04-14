package main

import (
	"limn/build"
	"limn/render"
	"fmt"
	"os"
    "path/filepath"

	"github.com/alecthomas/kong"
)


type Context struct {
    Debug bool
}

type BuildCmd struct {
    Path string `arg:"" default:"pwd" type:"existingdir" help:"path to site"`
}
func (b *BuildCmd) Run(ctx *BuildCmd) error {
    if ctx.Path == "" {
        ctx.Path = "."
    }

    rootfs := os.DirFS(ctx.Path)
    rndr := render.NewRenderer(rootfs, ".")
    rndr.BuildSiteModel()

    build.BuildSite(rndr, ctx.Path)
    return nil
}

type ShowCmd struct {
    Path string `arg:"" default:"pwd" type:"existingdir" help:"path to site"`
}
func (show *ShowCmd) Run(ctx *ShowCmd) error {
    if ctx.Path == "" {
        ctx.Path = "."
    }
    rootfs := os.DirFS(ctx.Path)
    rndr := render.NewRenderer(rootfs, ".")
    rndr.BuildSiteModel()

    fmt.Printf("config path : %s\n", filepath.Join(rndr.Wd(), "config.json"))
    fmt.Println(rndr.Template().DefinedTemplates())
    fmt.Println(rndr.Site())
    return nil
}

var cli struct {
    Debug bool `help:"Enable debug mode."`
    Build BuildCmd `cmd:"" help:"build site"`
    Show ShowCmd `cmd:"" help:"see site structure"`
}

func main() {
    ctx := kong.Parse(&cli, kong.Description("Static site generator"))
    err := ctx.Run(&Context{Debug: cli.Debug})
    ctx.FatalIfErrorf(err)
}
