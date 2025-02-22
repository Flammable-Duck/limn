package markdown

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"

	"github.com/yuin/goldmark-emoji"
	east "github.com/yuin/goldmark-emoji/ast"
	"github.com/yuin/goldmark-emoji/definition"
)

const emojiPath string = "/assets/emotes"

// TODO: load this from a yaml config
var Emojis []definition.Emoji = []definition.Emoji{
    definition.NewEmoji("snicker", nil, "snicker"),
    definition.NewEmoji("headphones", nil, "headphones"),
    definition.NewEmoji("bang", nil, "bang"),
    definition.NewEmoji("music", nil, "mp3"),
    definition.NewEmoji("pirateship", nil, "ship", "boat", "sailboat"),
    definition.NewEmoji("bigduck", nil, "duck"),
    definition.NewEmoji("sparkle", nil, "sparkle"),
    definition.NewEmoji("laughing", nil, "lol", "lmao", "grin"),
    definition.NewEmoji("new", nil, "new"),
}

func renderEmojis(w util.BufWriter, source []byte, n *east.Emoji,
    config *emoji.RendererConfig) {
    emojiTemplate := `<span class="emojiContainer">
    <img class="emoji" alt=":%s:" src="` + emojiPath + `/%s.gif">
    </span>`
    fmt.Fprintf(w, emojiTemplate, n.Value.Name, n.Value.Name)
}

var md = goldmark.New(
    goldmark.WithExtensions(
        extension.GFM,
        meta.Meta,
			emoji.New(
				emoji.WithEmojis(
					definition.NewEmojis(Emojis...),
				),
				emoji.WithRenderingMethod(emoji.Func),
				emoji.WithRendererFunc(renderEmojis),
			),
        ),
    goldmark.WithParserOptions(
        parser.WithAutoHeadingID(),
        ),
    goldmark.WithRendererOptions(
            html.WithUnsafe(),
        ),
    )


type Content struct {
    name string
    template string
    title string
    body string
    date int64
    pinned bool
}

func (c *Content) Name() string { return c.name }
func (c *Content) Title() string { return c.title }
func (c *Content) Date() int64 { return c.date }
func (c *Content) Body() template.HTML { return template.HTML(c.body) }
func (c *Content) Pinned() bool { return c.pinned }
func (c *Content) IsIndex() bool {
    return filepath.Base(c.Name()) == "index.md"
}

func RenderMarkdown(templates *template.Template, w io.Writer, src string,) error {
    content := LoadMarkdown(src)
    if tmpl := templates.Lookup(content.template); tmpl == nil {
        log.Printf("'%s' does not specify a valid template: '%s'", content.Title(), content.template)
        tmpl = template.Must(template.New("blank").Parse("{{.Body}}"))
        return tmpl.Execute(w, &content)
    } else {
    }
    log.Printf("title '%s' tmpl: '%s'", content.Title(), content.template)
    return templates.ExecuteTemplate(w, "content", &content)
}

func LoadMarkdown(path string) (c Content) {
    c.name = path
    file, err := os.ReadFile(path)
    if err != nil {
        log.Println(err)
        c.body = "Unable to load " + path
        return
    }

    var buf bytes.Buffer
    context := parser.NewContext()
    err = md.Convert(file, &buf, parser.WithContext(context))
    if err != nil {
        c.body = "Failed to parse" + path
        log.Println(err)
        return
    }
    c.body = buf.String()

    metaData := meta.Get(context)
    if title, ok := metaData["Title"].(string); ok {
        c.title = title
        log.Printf("%s", title)
    }

    if template, ok := metaData["Template"].(string); ok {
        c.template = strings.ToLower(template)
        log.Printf("template: %s", template)
    }

    // log.Printf("date: %T", metaData["Date"])
    if date, ok := metaData["Date"].(int); ok {
        c.date = int64(date)
    }

    if pinned, ok := metaData["Pinned"].(bool); ok {
        c.pinned = pinned
        log.Printf("%t", pinned)
    }
    return
}

func SortContent(c []Content) []Content {
    sort.Slice(c, func(i, j int) bool {
        return c[i].Date() > c[j].Date()
    })
    sort.Slice(c, func(i, j int) bool {
        if c[i].Pinned() { return true } else { return false }
    })
    sort.Slice(c, func(i, j int) bool {
        if c[i].IsIndex() { return true } else { return false }
    })
    return c
}
