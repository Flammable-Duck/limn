package render

import (
	"fmt"
	"html/template"
	"io"
	"limn/markdown"
	"log"
	"path/filepath"
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

type _URL string;
func (path _URL) Path() string {
    return string(path)
}
func (path _URL) PageName() string {
    pageName, _ := filepath.Split(string(path))
    return pageName
}
func (path _URL) NoteName() string {
    _, noteName := filepath.Split(path.Path())
    log.Printf("####>>>>####>>>> %s, %s", path, noteName)
    if noteName == "" || filepath.Ext(path.Path()) == "" {
        return "index.html"
    }
    return noteName
}
type UrlWrapper[S Content] struct {
    Content S
    _URL
}
func UrlWrap[S Content](c S, path string) UrlWrapper[S] {
    return UrlWrapper[S]{c, _URL(path)}
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
func (s Site) String() (str string) {
    for name, page := range s.pages {
        str = fmt.Sprintf("%s\n%-15s #%s\n%s", str, name, page.Title(), page)
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
        str = fmt.Sprintf("%s- %-15s | %s\n",
            str, name, note.Title())
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

func NewNote(data []byte) *Note {
    n := &Note{}
    n.body, n.metadata = markdown.LoadMarkdown(data)
    return n
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
func (n *Note) Render(w io.Writer, tmpl *template.Template, path string) error {
    wrapper := UrlWrap(n, path)
    err := tmpl.ExecuteTemplate(w, n.Template(), wrapper)
    if err != nil {
        return fmt.Errorf("%s: %s", n.Title(), err.Error())
    }
    return nil
}
