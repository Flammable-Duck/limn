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
	// "strings"
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
    templates map[string]*template.Template
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

// blog.questionable.services/article/approximating-html-template-inheritance
func (rndr *Renderer) initTemplates() error {
    rndr.templates = make(map[string]*template.Template)
    // template.Must(template.ParseFS(rndr.fs(),
    //     filepath.Join(TMPL_DIR, "*html")))
    layouts, err := fs.Glob(rndr.rootfs,
        filepath.Join(TMPL_DIR, "layouts/*.html"))
    if err != nil {
        return err
    }
    log.Printf("layouts: %v", layouts)

    includes, err := fs.Glob(rndr.rootfs,
        filepath.Join(TMPL_DIR, "includes/*.html"))
    if err != nil {
        return err
    }

    for _, layout := range layouts {
        files := append(includes, layout)
        rndr.templates[filepath.Base(layout)] = template.Must(
            template.ParseFiles(files...))
    }

    log.Print("initTemplates: templates loaded")
    return nil
}

func (rndr *Renderer) BuildSiteModel() {
    wd, err := os.Getwd()
    if err != nil { log.Fatal(err) }
    wd = path.Join(wd, rndr.sitedir)
    // var site Site = *NewSite()
    rndrFunc := func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            log.Fatal(err)
        }
        if d.IsDir() {
            return nil
        }

        // filename := filepath.Base(path)
        // page, _ := strings.CutPrefix(filepath.Dir(path), ROOT_DIR)
        page, filename := filepath.Split(path)
        src := filepath.Join(wd, path)
        if page == "" { page = "/" }

        log.Printf("\npath: '%s'\n page: '%s'\n filename: '%s'\n src: %s",
            path, page, filename, src)

        if filepath.Ext(filename) == ".md" {
            log.Printf("markdown: %s", d.Name())
            mdat, err := os.ReadFile(src)
            if err != nil { return err }
            
            rndr.site.Page(page).AddNote(filename, NewNote(mdat))
            return nil
        }

        return nil
    }

    os.RemoveAll(path.Join(wd, DEST_DIR))

    fs.WalkDir(rndr.fs(), ROOT_DIR, rndrFunc)
    fmt.Println(rndr.site)
}

func (rndr *Renderer) URL(w io.Writer, url string) error {
    path, name := filepath.Split(url)
    log.Printf("URL: %s + %s", path, name)
    var buf *bytes.Buffer
    page := rndr.site.Page(path)
    if name == "" {
        err := page.Render(buf, rndr.templates[page.Template()])
        if err != nil { return err }
        w.Write(buf.Bytes()); return nil
    }
    note := page.Notes()[name]
    err := note.Render(buf, rndr.templates[page.Template()])
    if err != nil { return err }
    w.Write(buf.Bytes())
    return nil
}

func (rndr *Renderer) renderTemplate(
    w io.Writer, name string, data interface{}) error {
    var b bytes.Buffer
    tmpl, ok := rndr.templates[name]
    // TODO: if template not defined (name == ""), use generic template instead
    if !ok {
        return fmt.Errorf("template '%s' not found", name)
    }

    log.Printf("using template %s (%s)", tmpl.Name(), name)
    err := tmpl.ExecuteTemplate(&b, "base", data)
    if err != nil {
        return err
    }
    _, err = b.WriteTo(w)
    return err
}
