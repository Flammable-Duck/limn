package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
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
type config struct {
    TemplateDir string `json:"tempates"`
    RootDir string `json:"root"`
    BuildDir string `json:"destination"`
}

type Renderer struct {
    templates *template.Template
    site Site
    rootfs fs.FS
    dest string
    sitedir string
}

func NewRenderer(rootfs fs.FS, sitedir string) *Renderer {
    rndr := &Renderer{rootfs: rootfs, sitedir: sitedir, site:*NewSite()}
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
        filepath.Join(TMPL_DIR, "*/*html")))
    log.Print(rndr.templates.DefinedTemplates())
    return nil
}
func (rndr *Renderer) BuildSiteModel() {
    wd, err := os.Getwd()
    if err != nil { log.Fatal(err) }
    wd = path.Join(wd, rndr.sitedir)
    rndrFunc := func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            log.Fatal(err)
        }
        if d.IsDir() {
            return nil
        }

        page, filename := filepath.Split(path)
        page = strings.TrimSuffix(strings.TrimPrefix(page, ROOT_DIR), "/")
        src := filepath.Join(wd, path)
        if page == "" { page = "/" }

        log.Printf("loading file: %s", d.Name())
        dat, err := os.ReadFile(src)
        if err != nil { return err }

        if filepath.Ext(filename) == ".md" {
            filename := strings.TrimSuffix(filename, "md") + "html"
            rndr.site.AddPage(page)
            rndr.site.Page(page).AddNote(filename, NewNote(dat))
            return nil
        }
        rndr.site.AddAsset(filepath.Join(page, filename), NewAsset(dat))
        return nil
    }

    os.RemoveAll(path.Join(wd, DEST_DIR))

    fs.WalkDir(rndr.fs(), ROOT_DIR, rndrFunc)
    fmt.Println(rndr.site)
}
func (rndr *Renderer) Template() *template.Template {
    return rndr.templates
}
func (rndr *Renderer) URL(w io.Writer, url string) error {
    path, name := filepath.Split(url)
    if path != "/" {
        path = strings.TrimSuffix(path, "/")
    }
    log.Printf("URL: %s%s", path, name)
    var buf bytes.Buffer
    page := rndr.site.Page(path)
    log.Printf("page title: %s", page.Title())
    if name == "" {
        err := page.Render(&buf, rndr.Template(), url)
        if err != nil { return err }
        w.Write(buf.Bytes()); return nil
    }
    note, ok := page.Notes()[name]
    if !ok { return fmt.Errorf("404")}
    err := note.Render(&buf, rndr.Template(), url)
    if err != nil { return err }
    w.Write(buf.Bytes())
    return nil
}
func (rndr *Renderer) WalkSite(f func(path string, c Content) error) error {
    for pName, p := range rndr.site.pages {
        if err := f(pName, p); err != nil {
            return err
        }
        for name, n := range p.notes {
            path := filepath.Join(pName, name)
            if err := f(path, n); err != nil {
                return err
            }
        }
    }
    for dName, d := range rndr.site.assets {
        for name, a := range d {
            path := filepath.Join(dName, name)
            if err := f(path, a); err != nil {
                return err
            }
        }
    }
    return nil
}
