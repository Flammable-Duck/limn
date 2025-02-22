package render

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"limn/markdown"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
    TMPL_DIR = "templates"
    ROOT_DIR = "root"
    DEST_DIR = "build"
)

type Renderer struct {
    templates *template.Template
    rootfs fs.FS
    dest string
    sitedir string
}

func NewRenderer(rootfs fs.FS, sitedir string) *Renderer {
    rndr := &Renderer{rootfs: rootfs, sitedir: sitedir}
    if err := rndr.checkTree(); err != nil {
        log.Fatal(err)
    }
    if err := rndr.initTemplates(); err != nil {
        log.Fatal(err)
    }
    return rndr
}

func (rndr *Renderer) fs() (subfs fs.FS) {
    var err error
    subfs, err = fs.Sub(rndr.rootfs, rndr.sitedir)
    if err != nil { log.Fatal(err) }
    return
}

func (rndr *Renderer) subFs(dir string) (subfs fs.FS) {
    var err error
    subfs, err = fs.Sub(rndr.rootfs, filepath.Join(rndr.sitedir, dir))
    if err != nil { log.Fatal(err) }
    return
}

func (rndr *Renderer) checkTree() error {
    log.Printf("%s", rndr.rootfs)
    if tmpldir, err := fs.Stat(rndr.fs(), TMPL_DIR); err != nil ||
    !tmpldir.IsDir() {
        return fmt.Errorf("templates dir not found in site tree.")
    }
    if rootdir, err := fs.Stat(rndr.fs(), ROOT_DIR); err != nil ||
    !rootdir.IsDir() {
        return fmt.Errorf("root dir not found in site tree.")
    }
    log.Print("checkTree: tree ok")
    return nil
}

func (rndr *Renderer) initTemplates() error {
    rndr.templates = template.Must(template.ParseFS(rndr.fs(),
        filepath.Join(TMPL_DIR, "*html")))
    log.Print("initTemplates: templates loaded")
    return nil
}

func (rndr *Renderer) Render() {
    wd, err := os.Getwd()
    if err != nil { log.Fatal(err) }
    wd = path.Join(wd, rndr.sitedir)
    rndrFunc := func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            log.Fatal(err)
        }

        dest, _ := strings.CutPrefix(path, ROOT_DIR)
        dest = filepath.Join(wd, DEST_DIR, dest)
        src := filepath.Join(wd, path)

        if d.IsDir() {
            log.Printf("dir: %s/", d.Name())
            if err := os.Mkdir(dest,  0750); err != nil {
                log.Fatal(err)
            }
            return nil
        }

        switch filepath.Ext(d.Name()) {
        case ".md":
            log.Printf("markdown: %s", d.Name())
            // markdown.LoadContent
            dest = strings.TrimSuffix(dest, filepath.Ext(dest)) + ".html"
            fout, err := os.Create(dest)
            if err != nil {
                log.Fatal(err)
            }
            defer fout.Close()
            err = markdown.RenderMarkdown(rndr.templates, fout, src)
            if err != nil {
                log.Fatal(err)
            }
        default:
            log.Printf("other: %s", d.Name())
            log.Printf(">> %s", dest)
            fin, err := os.Open(src)
            if err != nil {
                log.Fatal(err)
            }
            defer fin.Close()
            fout, err := os.Create(dest)
            if err != nil {
                log.Fatal(err)
            }
            defer fout.Close()

            io.Copy(fout, fin)
        }

        return nil
    }

    os.RemoveAll(path.Join(wd, DEST_DIR))

    fs.WalkDir(rndr.fs(), ROOT_DIR, rndrFunc)
}
