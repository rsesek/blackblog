//
// Blackblog
// Copyright 2012 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/russross/blackfriday"
)

const ConfigFileName = "blackblog.json"

// Blog is a structure that contains the configuration of a blackblog. This is
// stored as a JSON file, in the blog root directory, named `blackblog.json`.
type Blog struct {
	// The configuration data.
	config configFile

	// Parsed values of the string versions in the config.
	markdownExtensions  int
	markdownHTMLOptions int

	// Path to the configuration file (including "blackblog.json").
	configPath string
}

type configFile struct {
	// The name of the blog, used in page titles.
	Title string

	// Path to the directory containing the Markdown files used for posts.
	PostsDir string

	// Path to the templates directory, used to format the blog.
	TemplatesDir string

	// Static files that are copied to the OutputDir or that are served in server
	// mode to support the templates.
	StaticFilesDir string

	// When rendering the blog to static files, the directory to place the
	// output.
	OutputDir string

	// When running as a server, the port on which the server is bound.
	Port int

	// A list of string EXTENSION_ constants to pass to Blackfriday Markdown.
	MarkdownExtensions []string

	// A list of HTML_ options to pass to the Blackfriday Markdown HTML renderer.
	MarkdownHTMLOptions []string
}

func (b *Blog) Title() string {
	return b.config.Title
}

func (b *Blog) Port() int {
	return b.config.Port
}

func (b *Blog) TemplatesDir() string {
	return b.getPath(b.config.TemplatesDir)
}

func (b *Blog) StaticFilesDir() string {
	if b.config.StaticFilesDir == "" {
		return ""
	}
	return b.getPath(b.config.StaticFilesDir)
}

func (b *Blog) GetPostsDir() string {
	return b.getPath(b.config.PostsDir)
}

func (b *Blog) GetOutputDir() string {
	return b.getPath(b.config.OutputDir)
}

func (b *Blog) getPath(part string) string {
	return path.Join(path.Dir(b.configPath), part)
}

func (b *Blog) GetMarkdownExtensions() int {
	return b.markdownExtensions
}

func (b *Blog) GetMarkdownHTMLOptions() int {
	return b.markdownHTMLOptions
}

// ReadBlog reads the blog configuration from the specified file path. This
// does not need to end in `blackblog.json`.
func ReadBlog(p string) (*Blog, error) {
	if !strings.HasSuffix(p, ConfigFileName) {
		p = path.Join(p, ConfigFileName)
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := json.NewDecoder(f)
	var config configFile
	if err = d.Decode(&config); err != nil {
		return nil, err
	}

	blog := &Blog{
		config:     config,
		configPath: path.Clean(p),
	}
	if err := blog.parseOptions(); err != nil {
		return nil, err
	}

	return blog, nil
}

func (b *Blog) parseOptions() error {
	if err := parseFlags(&b.markdownExtensions, markdownExtensions, b.config.MarkdownExtensions); err != nil {
		return fmt.Errorf("Markdown Extensions: %v", err)
	}
	options := b.config.MarkdownHTMLOptions
	if options == nil {
		// The default options that were specified before the configuration allowed
		// specification.
		options = []string{"HTML_USE_SMARTYPANTS", "HTML_USE_XHTML", "HTML_SMARTYPANTS_LATEX_DASHES"}
	}
	if err := parseFlags(&b.markdownHTMLOptions, markdownHTMLOptions, options); err != nil {
		return fmt.Errorf("Markdown HTML Options: %v", err)
	}
	return nil
}

func parseFlags(flags *int, allowedFlags map[string]int, specifiedFlags []string) error {
	for _, flag := range specifiedFlags {
		value, ok := allowedFlags[flag]
		if !ok {
			return fmt.Errorf("Unknown flag %q", flag)
		}
		*flags |= value
	}
	return nil
}

var (
	markdownExtensions = map[string]int{
		"EXTENSION_NO_INTRA_EMPHASIS":          blackfriday.EXTENSION_NO_INTRA_EMPHASIS,
		"EXTENSION_TABLES":                     blackfriday.EXTENSION_TABLES,
		"EXTENSION_FENCED_CODE":                blackfriday.EXTENSION_FENCED_CODE,
		"EXTENSION_AUTOLINK":                   blackfriday.EXTENSION_AUTOLINK,
		"EXTENSION_STRIKETHROUGH":              blackfriday.EXTENSION_STRIKETHROUGH,
		"EXTENSION_LAX_HTML_BLOCKS":            blackfriday.EXTENSION_LAX_HTML_BLOCKS,
		"EXTENSION_SPACE_HEADERS":              blackfriday.EXTENSION_SPACE_HEADERS,
		"EXTENSION_HARD_LINE_BREAK":            blackfriday.EXTENSION_HARD_LINE_BREAK,
		"EXTENSION_TAB_SIZE_EIGHT":             blackfriday.EXTENSION_TAB_SIZE_EIGHT,
		"EXTENSION_FOOTNOTES":                  blackfriday.EXTENSION_FOOTNOTES,
		"EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK": blackfriday.EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK,
		"EXTENSION_HEADER_IDS":                 blackfriday.EXTENSION_HEADER_IDS,
		"EXTENSION_TITLEBLOCK":                 blackfriday.EXTENSION_TITLEBLOCK,
	}

	markdownHTMLOptions = map[string]int{
		"HTML_SKIP_HTML":                blackfriday.HTML_SKIP_HTML,
		"HTML_SKIP_STYLE":               blackfriday.HTML_SKIP_STYLE,
		"HTML_SKIP_IMAGES":              blackfriday.HTML_SKIP_IMAGES,
		"HTML_SKIP_LINKS":               blackfriday.HTML_SKIP_LINKS,
		"HTML_SANITIZE_OUTPUT":          blackfriday.HTML_SANITIZE_OUTPUT,
		"HTML_SAFELINK":                 blackfriday.HTML_SAFELINK,
		"HTML_NOFOLLOW_LINKS":           blackfriday.HTML_NOFOLLOW_LINKS,
		"HTML_HREF_TARGET_BLANK":        blackfriday.HTML_HREF_TARGET_BLANK,
		"HTML_TOC":                      blackfriday.HTML_TOC,
		"HTML_OMIT_CONTENTS":            blackfriday.HTML_OMIT_CONTENTS,
		"HTML_COMPLETE_PAGE":            blackfriday.HTML_COMPLETE_PAGE,
		"HTML_USE_XHTML":                blackfriday.HTML_USE_XHTML,
		"HTML_USE_SMARTYPANTS":          blackfriday.HTML_USE_SMARTYPANTS,
		"HTML_SMARTYPANTS_FRACTIONS":    blackfriday.HTML_SMARTYPANTS_FRACTIONS,
		"HTML_SMARTYPANTS_LATEX_DASHES": blackfriday.HTML_SMARTYPANTS_LATEX_DASHES,
		"HTML_FOOTNOTE_RETURN_LINKS":    blackfriday.HTML_FOOTNOTE_RETURN_LINKS,
	}
)
