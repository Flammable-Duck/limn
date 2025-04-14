package render

import (
    "fmt"
    "html/template"
    "io"
    "log"
    "path/filepath"
    "strings"
)


type Content interface {
    Render(io.Writer, *template.Template, string) error
    Template() string
}

type Asset struct {
    data []byte
}
var _ Content = &Asset{}

type Note struct {
    body template.HTML
    metadata map[string]interface{}
}
var _ Content = &Note{}

type Page struct {
    notes map[string]*Note
}
var _ Content = &Page{}

type URL string;
func (path URL) String() string {
    return string(path)
}
func (path URL) Path() string {
    return path.String()
}
func (path URL) PageName() string {
    dir, name := filepath.Split(path.String())
    if filepath.Ext(name) == "" {
        dir = path.String()
    }
    if !strings.HasSuffix(dir, "/") {
        dir = fmt.Sprintf("%s/", dir)
    }
    return dir
}
func (path URL) NoteName() string {
    noteName := filepath.Base(path.String())
    if noteName == "" || filepath.Ext(path.Path()) == "" {
        return "index.html"
    }
    return noteName
}
type UrlWrapper[S Content] struct {
    Content S
    URL
}
func UrlWrap[S Content](c S, path string) UrlWrapper[S] {
    return UrlWrapper[S]{c, URL(path)}
}

type Site struct {
    pages map[string]*Page
    assets map[string]map[string]*Asset
}
func NewSite() *Site {
    return &Site{
        pages: make(map[string]*Page),
        assets: make(map[string]map[string]*Asset),
    }
}
func (s *Site) AddPage(name string) {
    _, ok := s.pages[name]
    if !ok {
        s.pages[name] = &Page{notes: make(map[string]*Note)}
    }
}
func (s *Site) Page(name string) *Page {
    page, ok := s.pages[name]
    if !ok {
        log.Fatalf("Page '%s' not found", name)
    }
    return page
}
func (s *Site) AddAsset(path string, a *Asset) {
    dirName, name := filepath.Split(path)
    if _, ok := s.assets[dirName]; !ok {
        s.assets[dirName] = make(map[string]*Asset)
    }
    s.assets[dirName][name] = a
}
func (s *Site) Title() string {
    p, ok := s.pages["/"]
    if !ok { return "" }
    return p.Title()
}
func (s *Site) Pages() map[string]*Page {
    return s.pages
}
func (s *Site) Template() string {
    return "base"
}
func (s *Site) Render(w io.Writer, tmpl *template.Template, path string) error {
    wrapper := UrlWrap(s, path)
    err := tmpl.ExecuteTemplate(w, s.Template(), wrapper)
    if err != nil {
        return fmt.Errorf("%s: %s", s.Title(), err.Error())
    }
    return nil
}
func (s Site) String() (str string) {
    for name, page := range s.pages {
        template := page.Template()
        if template != "" {
            template = fmt.Sprintf("[%s]", template)
        }
        str = fmt.Sprintf("%s\n%-20s #%s %s\n%s",
            str, name, page.Title(), template, page)
    }
    return str
}

func (p *Page) AddNote(path string, n *Note) {
    if p.notes == nil {
        p.notes = make(map[string]*Note)
    }
    p.notes[path] = n
}
func (p *Page) Notes() map[string]*Note {
    return p.notes
}
func (p *Page) Title() string {
    index, ok := p.Notes()["index.html"]
    if !ok { return "" }
    return index.Title()
}
func (p *Page) Template() string {
    index, ok := p.Notes()["index.html"]
    if !ok { return "" }
    return index.Template()
}
func (p Page) String() (str string) {
    for name, note := range p.Notes() {
        template := note.Template()
        if template != "" {
            template = fmt.Sprintf("[%s]", template)
        }
        str = fmt.Sprintf("%s- %-25s | %s %s\n",
            str, name, note.Title(), template)
    }
    return
}
func (p *Page) Render(w io.Writer, tmpl *template.Template, path string) error {
    wrapper := UrlWrap(p, path)
    err := tmpl.ExecuteTemplate(w, p.Template(), wrapper)
    if err != nil {
        return fmt.Errorf("%s: %s", p.Title(), err.Error())
    }
    return nil
}

func NewAsset(data []byte) (a *Asset) {
    return &Asset{data: data}
}
func (a *Asset) Render(w io.Writer, _ *template.Template, _ string) (err error) {
    _, err = w.Write(a.data)
    return
}
func (a *Asset) Template() string {
    return "raw"
}

func NewNote(body template.HTML, meta map[string]interface{}) *Note {
    return &Note{body: body, metadata: meta}
}
func (n *Note) Body() template.HTML {
    return n.body
}
func (n *Note) Title() string {
    t, ok := n.metadata["Title"]
    if !ok { return "" }
    title, ok := t.(string)
    if !ok { return "" }
    return title
}
func (n *Note) Template() string {
    t, ok := n.metadata["Template"]
    if !ok { return "" }
    template, ok := t.(string)
    if !ok { return "" }
    return template
}
func (n *Note) CoverImgURL() string {
    d, ok := n.metadata["CoverImg"]
    if !ok { return "" }
    url, ok := d.(string)
    if !ok { return "" }
    return url
}
func (n *Note) Render(w io.Writer, tmpl *template.Template, path string) error {
    wrapper := UrlWrap(n, path)
    err := tmpl.ExecuteTemplate(w, n.Template(), wrapper)
    if err != nil {
        return fmt.Errorf("%s: %s", n.Title(), err.Error())
    }
    return nil
}
