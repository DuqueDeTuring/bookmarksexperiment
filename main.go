package main

import (
	"bytes"
	"fmt"
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"io"
	"os"
)




// ASTs

type TagsAST struct {
	gast.BaseInline
}

func (n *TagsAST) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

var KindTags = gast.NewNodeKind("Tags")

func (n *TagsAST) Kind() gast.NodeKind {
	return KindTags
}

func NewTags() *TagsAST {
	return &TagsAST{}
}

// /AST

//var tagsListRegexp = regexp.MustCompile(`\w+}`)
type tagsDelimiterProcessor struct {

}

func (p *tagsDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == '{'
}

func (p *tagsDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *tagsDelimiterProcessor) OnMatch(consumes int) gast.Node {
	return NewTags()
}

var defaultTagsDelimiterProcessor = &tagsDelimiterProcessor{}

type tagsParser struct {}

var defaultTagsParser = &tagsParser{}

func NewTagsParser() parser.InlineParser {
	return defaultTagsParser
}

func (t *tagsParser) Trigger() []byte {
	return []byte{'{'}
}

func (t *tagsParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	before := block.PrecendingCharacter()
	line, segment := block.PeekLine()
	node := parser.ScanDelimiter(line, before, 1, defaultTagsDelimiterProcessor)
	if node == nil {
		return nil
	}

	//m := tagsListRegexp.FindSubmatch(line)
	//if m == nil {
	//	return nil
	//}
	//fmt.Print(m)
	//
	//l := len(m[0])
	//node.Segment = segment.WithStop(segment.Start + l)
	node.Segment = segment.WithStop(segment.Start + node.OriginalLength)
	//block.Advance(l)
	block.Advance(node.OriginalLength)
	pc.PushDelimiter(node)
	return node
}

func (t *tagsParser) CloseBlock(parent gast.Node, pc parser.Context) {

}

type TagsHTMLRenderer struct {
	html.Config
}

func NewTagsHTMLRenderer(opts ...html.Option) renderer.NodeRenderer {
	r := &TagsHTMLRenderer{Config: html.NewConfig(),}
	for _, opt := range opts {
		opt.SetHTMLOption(&r.Config)
	}
	return r
}

func (r *TagsHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindTags, r.renderTags)
}

var TagsAttributeFilter = html.GlobalAttributeFilter

func (r *TagsHTMLRenderer) renderTags(w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<del")
			html.RenderAttributes(w, n, TagsAttributeFilter)
			_ =  w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<del>")
		}
	} else {
		_, _ = w.WriteString("</del>")
	}
	return gast.WalkContinue, nil
}

type tags struct {

}

var Tags = &tags{}

func (e *tags) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewTagsParser(), 500)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewTagsHTMLRenderer(), 500)))
}

func main() {
	file, err := os.Open("bookmarks.md")
	if err != nil {
		panic(err)
	}
	source, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer

	md := goldmark.New(
		goldmark.WithExtensions(Tags),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	if err := md.Convert(source, &buf); err != nil {
		panic(err)
	}

	out, err := os.Create("bookmarks.html")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	buf.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n <meta charset=\"utf-8\" />\n <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\" />\n\n <title></title>\n \n</head>\n\n<body>")

	n, err := buf.WriteTo(out)
	if err != nil {
		panic(err)
	}
	buf.WriteString("\n</body>\n</html>\n\n")

	fmt.Println("n:", n)
}
