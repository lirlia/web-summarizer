package domain

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ledongthuc/pdf"
	"github.com/m-mizutani/goerr"
)

type fileType string

const (
	fileTypeHtml fileType = "html"
	fileTypePdf  fileType = "pdf"
)

func NewContentsClient(url string) *contentsClient {
	return &contentsClient{url: url}
}

type contentsClient struct {
	url string
}

func (r *contentsClient) FileType() fileType {
	// check url has period
	if !strings.Contains(r.url, ".") {
		return fileTypeHtml
	}

	lowerURL := strings.ToLower(r.url)

	switch {
	case strings.HasSuffix(lowerURL, string(fileTypeHtml)):
		return fileTypeHtml
	case strings.HasSuffix(lowerURL, string(fileTypePdf)):
		return fileTypePdf
	}

	return fileTypeHtml
}

func (r *contentsClient) GetContents(ctx context.Context) (string, error) {
	switch r.FileType() {
	case fileTypeHtml:
		return r.htmlContents(ctx)
	case fileTypePdf:
		return r.pdfContents(ctx)
	}

	return "", goerr.New(fmt.Sprintf("Unsupported file type: %s", r.FileType()), nil)
}

func (r *contentsClient) htmlContents(ctx context.Context) (string, error) {

	agent := NewHeadlessAgent()
	buf, err := agent.Get(ctx, r.url)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(buf)))
	if err != nil {
		return "", goerr.Wrap(err)
	}

	var content strings.Builder
	doc.Find("p, h1, h2, h3, h4, h5").Each(func(i int, s *goquery.Selection) {
		content.WriteString(s.Text() + "\n")
	})

	return content.String(), nil
}

func (r *contentsClient) pdfContents(ctx context.Context) (string, error) {

	opts := []AgentOption{
		WithHeaders(map[string]string{
			"Accept": "application/pdf",
		}),
	}

	agent := NewHTTPAgent(opts...)
	buf, err := agent.Get(ctx, r.url)
	if err != nil {
		return "", err
	}

	// Parse PDF data in memory
	reader := bytes.NewReader(buf)
	pdfReader, err := pdf.NewReader(reader, int64(len(buf)))
	if err != nil {
		return "", goerr.Wrap(err)
	}

	var result strings.Builder
	numPages := pdfReader.NumPage()
	for i := 1; i <= numPages; i++ {
		page := pdfReader.Page(i)
		if page.V.IsNull() {
			continue
		}
		content, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		result.WriteString(content)
	}

	return result.String(), nil
}
