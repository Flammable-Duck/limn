package render

import (
	"fmt"
	"html/template"
	"io"
	"limn/markdown"
)

type Content interface {
    Render(io.Writer, *template.Template) error
    Template() string
}

type html struct {}

type Asset struct {
    data []byte
}
var _ Content = &Asset{}

type Note struct {
    html
    body template.HTML
    metadata map[string]interface{}
}
var _ Content = &Note{}

type Page struct {
    html
    notes map[string]*Note
}
var _ Content = &Page{}

type Site struct {
    pages map[string]*Page
}

func (n *html) Render(w io.Writer, tmpl *template.Template) error {
    return tmpl.Execute(w, n)
}

func NewSite() *Site {
    return &Site{pages: make(map[string]*Page)}
}
func (s *Site) Page(path string) *Page {
    _, ok := s.pages[path]
    if !ok {
        s.pages[path] = NewPage()
    }
    return s.pages[path]
}
func (s Site) String() (str string) {
    for name, page := range s.pages {
        str = fmt.Sprintf("%s\n%-15s #%s\n%s", str, name, page.Title(), page)
    }
    return str
}

func NewPage() *Page {
    return &Page{notes: make(map[string]*Note)}
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
    index, ok := p.Notes()["index.md"]
    if !ok { return "" }
    return index.Title()
}
func (p *Page) Template() string {
    index, ok := p.Notes()["index.md"]
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

func NewAsset(data []byte) (a *Asset) {
    return &Asset{data: data}
}
func (a *Asset) Render(w io.Writer, _ *template.Template) (err error) {
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
