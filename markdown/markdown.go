package markdown

import (
	"bytes"
	"fmt"
	"html/template"
	"limn/config"
	"log"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	"github.com/yuin/goldmark-emoji"
	east "github.com/yuin/goldmark-emoji/ast"
	"github.com/yuin/goldmark-emoji/definition"
)

type wrapImageTransformer struct {}
func (t *wrapImageTransformer) Transform(
    node *ast.Document, reader text.Reader, pc parser.Context) {
    var images []*ast.Image
    ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
        if imageNode, ok := n.(*ast.Image); ok && entering {
            images = append(images, imageNode)
		}
        return ast.WalkContinue, nil
    })
    for _, imageNode := range images {
        if imageNode.Parent() != nil {
            imgWrapper := ast.NewParagraph()
            imgWrapper.SetAttribute([]byte("class"), "imgWrapper")
            parent := imageNode.Parent()
            sibiling := imageNode.PreviousSibling()
            imgWrapper.AppendChild(imgWrapper, imageNode)
            parent.InsertAfter(parent, sibiling, imgWrapper)
        }
    }
}

type MarkdownRenderer struct {
    md goldmark.Markdown
    config *config.Config
    emojis map[string]string

}
func NewRenderer(cfg *config.Config) ( *MarkdownRenderer) {
    mdRndr := &MarkdownRenderer{config: cfg, emojis: make(map[string]string)}
    for _, emoji := range mdRndr.config.Emojis {
        mdRndr.emojis[emoji.Name] = emoji.Path
    }
    mdRndr.initMD()
    return mdRndr
}
func (mdRndr *MarkdownRenderer) EmojiDef() []definition.Emoji {
    emojiDef := []definition.Emoji{}
    for name := range mdRndr.emojis {
        emojiDef = append(emojiDef, definition.NewEmoji(name, nil, name))
    }
    return emojiDef
}
func (mdRndr *MarkdownRenderer) initMD() {
    mdRndr.md = goldmark.New(
    goldmark.WithExtensions(
        extension.GFM,
        meta.Meta,
        emoji.New(
            emoji.WithEmojis(
                definition.NewEmojis(mdRndr.EmojiDef()...),
            ),
            emoji.WithRenderingMethod(emoji.Func),
            emoji.WithRendererFunc(mdRndr.renderEmojis),
        ),
        ),
    goldmark.WithParserOptions(
        parser.WithAutoHeadingID(),
        parser.WithASTTransformers(
            util.Prioritized(&wrapImageTransformer{}, 9999),
            ),
        ),
    goldmark.WithRendererOptions(
            html.WithUnsafe(),
        ),
    )
}
func (mdRndr *MarkdownRenderer) renderEmojis(
    w       util.BufWriter,
    source  []byte, n *east.Emoji,
    config  *emoji.RendererConfig) {

    emojiTemplate := `<span class="emojiContainer">
    <img class="emoji" alt=":%s:" src="` + mdRndr.emojis[n.Value.Name] + `">
    </span>`

    fmt.Fprintf(w, emojiTemplate, n.Value.Name)
}
func (mdRndr *MarkdownRenderer) LoadMarkdown(data []byte) (body template.HTML, metadata map[string]interface{}) {
    var buf bytes.Buffer
    context := parser.NewContext()
    err := mdRndr.md.Convert(data, &buf, parser.WithContext(context))
    if err != nil {
        body = template.HTML("Failed to parse markdown.")
        log.Println(err)
        return
    }

    body = template.HTML(buf.String())
    metadata = meta.Get(context)
    return
}
