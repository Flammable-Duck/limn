package markdown

import (
	"bytes"
	"fmt"
	"html/template"
	"log"

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

func LoadMarkdown(data []byte) (body template.HTML, metadata map[string]interface{}) {
    var buf bytes.Buffer
    context := parser.NewContext()
    err := md.Convert(data, &buf, parser.WithContext(context))
    if err != nil {
        body = template.HTML("Failed to parse markdown.")
        log.Println(err)
        return
    }

    body = template.HTML(buf.String())
    metadata = meta.Get(context)
    return
}
