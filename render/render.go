package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
    "limn/config"
    "limn/markdown"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var funcMap = template.FuncMap{
    "SiteWithURL": UrlWrap[*Site],
    "PageWithURL": UrlWrap[*Page],
    "NoteWithURL": UrlWrap[*Note],
}

type Renderer struct {
    templates *template.Template
    rootfs fs.FS
    site Site
    mdRndr *markdown.MarkdownRenderer
    config config.Config
    dest string
    sitedir string
}

func NewRenderer(rootfs fs.FS, sitedir string) *Renderer {
    rndr := &Renderer{rootfs: rootfs, sitedir: sitedir, site:*NewSite()}
    if cfg, err := fs.Stat(rndr.fs(), "config.json"); err != nil ||
    cfg.IsDir() {
        rndr.config = config.DefaultConfig()
    } else if err := rndr.loadConfig(); err != nil {
        log.Fatal(err)
    }
    log.Printf("config:\n---\n%s\n---", rndr.config)
    rndr.mdRndr = markdown.NewRenderer(&rndr.config)
    if err := rndr.checkTree(); err != nil {
        log.Fatal(err)
    }
    if err := rndr.initTemplates(); err != nil {
        log.Fatal(err)
    }
    return rndr
}
func (rndr *Renderer) checkTree() error {
    log.Printf("%s", rndr.rootfs)
    if tmpldir, err := fs.Stat(rndr.fs(), rndr.config.Paths.Templates); err != nil ||
    !tmpldir.IsDir() {
        return fmt.Errorf("templates dir not found in site tree.")
    }
    if rootdir, err := fs.Stat(rndr.fs(), rndr.config.Paths.Root); err != nil ||
    !rootdir.IsDir() {
        return fmt.Errorf("root dir not found in site tree.")
    }
    log.Print("checkTree: tree ok")
    return nil
}
func (rndr *Renderer) loadConfig() error {
    wd, err := os.Getwd()
    if err != nil { return err }
    wd = path.Join(wd, rndr.sitedir)
    data, err := os.ReadFile(filepath.Join(wd, "config.json"))
    cfg, err := config.ReadConfig(data)
    if err != nil { return err }
    rndr.config = cfg
    return nil
}
func (rndr *Renderer) initTemplates() error {
    rndr.templates = template.Must(
        template.New("templates").Funcs(funcMap).ParseFS(rndr.fs(),
        filepath.Join(rndr.config.Paths.Templates, "*/*html")))
    log.Print(rndr.templates.DefinedTemplates())
    return nil
}
func (rndr *Renderer) fs() (subfs fs.FS) {
    var err error
    subfs, err = fs.Sub(rndr.rootfs, rndr.sitedir)
    if err != nil { log.Fatal(err) }
    return
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

        src := filepath.Join(wd, path)
        url := URL(strings.TrimPrefix(path, rndr.config.Paths.Root))

        log.Printf("loading file: %s", d.Name())
        dat, err := os.ReadFile(src)
        if err != nil { return err }

        if filepath.Ext(url.NoteName()) == ".md" {
            url = URL( strings.TrimSuffix(
                url.String(), "md") + "html")
            rndr.site.AddPage(url.PageName())
            md, meta := rndr.mdRndr.LoadMarkdown(dat)
            rndr.site.Page(
                url.PageName()).AddNote(url.NoteName(),
                NewNote(md, meta))
            return nil
        }
        rndr.site.AddAsset(url.Path(), NewAsset(dat))
        return nil
    }

    os.RemoveAll(path.Join(wd, rndr.config.Paths.Build))

    fs.WalkDir(rndr.fs(), rndr.config.Paths.Root, rndrFunc)
    fmt.Println(rndr.site)
}
func (rndr *Renderer) Template() *template.Template {
    return rndr.templates
}
func (rndr *Renderer) Site() *Site {
    return &rndr.site
}
func (rndr *Renderer) URL(w io.Writer, path string) error {
    url := URL(path)
    buf := bytes.NewBuffer([]byte{})
    switch filepath.Ext(url.Path()) {
    case ".html":
        page, ok := rndr.site.Pages()[url.PageName()]
        if !ok {
            return fmt.Errorf("path %s page %s not found.",
                url.Path(), url.PageName())
        }
        var err error
        if url.NoteName() == "index.html" {
            err = page.Render(buf, rndr.Template(), page.Template())
        } else {
            ctnt, ok := page.Notes()[url.NoteName()]
            if !ok {
                return fmt.Errorf("Page %s Note %s not found.",
                    url.PageName(), url.NoteName())
            }
            err = ctnt.Render(buf, rndr.Template(), ctnt.Template())
        }
        if err != nil {
            return err
        }
    default:
        asset, ok := rndr.site.assets[url.PageName()][url.NoteName()]
        if !ok {
            return fmt.Errorf("path %s content %s not found.",
                url.PageName(), url.NoteName())
        }
        asset.Render(buf, rndr.Template(), asset.Template())
        w.Write(buf.Bytes())
        return nil
    }
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
