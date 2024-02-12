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

	"github.com/russross/blackfriday/v2"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
)

const ConfigFileName = "blackblog.json"

// Blog is a structure that contains the configuration of a blackblog. This is
// stored as a JSON file, in the blog root directory, named `blackblog.json`.
type Blog struct {
	// The configuration data.
	config configFile

	// For V2 configs, the Markdown renderer.
	md goldmark.Markdown

	// For V1 configs, parsed values of the string versions in the config.
	markdownExtensions  blackfriday.Extensions
	markdownHTMLOptions blackfriday.HTMLFlags

	// Path to the configuration file (including "blackblog.json").
	configPath string
}

const configVersion = 2

type configFile struct {
	ConfigVersion int

	// The name of the blog, used in page titles.
	Title string

	// The full base URL of the blog. Used for producing permalinks.
	URL string

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

	GoldmarkConfig GoldmarkConfig
}

type GoldmarkConfig struct {
	Extension struct {
		// Enable table extension.
		Table bool

		// Typographic substitutions / "Smartypants".
		Typographer struct {
			Disable bool
		}
	}

	Parse struct {
	}

	Render struct {
		// Output XHTML instead of HTML5.
		XHTML bool

		// Allow raw HTML.
		Unsafe bool
	}
}

func (b *Blog) Title() string {
	return b.config.Title
}

func (b *Blog) URL() string {
	return b.config.URL
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

func (b *Blog) GetMarkdownExtensions() blackfriday.Extensions {
	return b.markdownExtensions
}

func (b *Blog) GetMarkdownHTMLOptions() blackfriday.HTMLFlags {
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

	if want, got := configVersion, config.ConfigVersion; want != got && got != 0 {
		return nil, fmt.Errorf("ConfigVersion not compatible, need %d and got %d", want, got)
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
	if b.config.ConfigVersion == configVersion {
		gc := b.config.GoldmarkConfig

		// Extensions.
		exts := make([]goldmark.Extender, 0)

		if !gc.Extension.Typographer.Disable {
			exts = append(exts, extension.NewTypographer())
		}
		if gc.Extension.Table {
			exts = append(exts, extension.NewTable())
		}

		// Parse options.
		// Nothing yet...

		// Render options.
		ropts := make([]renderer.Option, 0)

		if gc.Render.XHTML {
			ropts = append(ropts, html.WithXHTML())
		}
		if gc.Render.Unsafe {
			ropts = append(ropts, html.WithUnsafe())
		}

		// Assemble!
		b.md = goldmark.New(
			goldmark.WithExtensions(exts...),
			goldmark.WithRendererOptions(ropts...))
		return nil
	}
	for _, flag := range b.config.MarkdownExtensions {
		value, ok := markdownExtensions[flag]
		if !ok {
			return fmt.Errorf("Unknown Markdown extensions: %v", flag)
		}
		b.markdownExtensions |= value
	}

	if b.config.MarkdownHTMLOptions == nil {
		// The default options that were specified before the configuration allowed
		// specification.
		b.markdownHTMLOptions = blackfriday.Smartypants | blackfriday.UseXHTML | blackfriday.SmartypantsLatexDashes | blackfriday.SmartypantsDashes
	} else {
		for _, flag := range b.config.MarkdownHTMLOptions {
			value, ok := markdownHTMLOptions[flag]
			if !ok {
				return fmt.Errorf("Unknown Markdown HTML option: %v", flag)
			}
			b.markdownHTMLOptions |= value
		}
	}
	return nil
}

var (
	markdownExtensions = map[string]blackfriday.Extensions{
		// Legacy names:
		"EXTENSION_NO_INTRA_EMPHASIS":          blackfriday.NoIntraEmphasis,
		"EXTENSION_TABLES":                     blackfriday.Tables,
		"EXTENSION_FENCED_CODE":                blackfriday.FencedCode,
		"EXTENSION_AUTOLINK":                   blackfriday.Autolink,
		"EXTENSION_STRIKETHROUGH":              blackfriday.Strikethrough,
		"EXTENSION_LAX_HTML_BLOCKS":            blackfriday.LaxHTMLBlocks,
		"EXTENSION_SPACE_HEADERS":              blackfriday.SpaceHeadings,
		"EXTENSION_HARD_LINE_BREAK":            blackfriday.HardLineBreak,
		"EXTENSION_TAB_SIZE_EIGHT":             blackfriday.TabSizeEight,
		"EXTENSION_FOOTNOTES":                  blackfriday.Footnotes,
		"EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK": blackfriday.NoEmptyLineBeforeBlock,
		"EXTENSION_HEADER_IDS":                 blackfriday.HeadingIDs,
		"EXTENSION_TITLEBLOCK":                 blackfriday.Titleblock,

		"NoExtensions":           blackfriday.NoExtensions,
		"NoIntraEmphasis":        blackfriday.NoIntraEmphasis,
		"Tables":                 blackfriday.Tables,
		"FencedCode":             blackfriday.FencedCode,
		"Autolink":               blackfriday.Autolink,
		"Strikethrough":          blackfriday.Strikethrough,
		"LaxHTMLBlocks":          blackfriday.LaxHTMLBlocks,
		"SpaceHeadings":          blackfriday.SpaceHeadings,
		"HardLineBreak":          blackfriday.HardLineBreak,
		"TabSizeEight":           blackfriday.TabSizeEight,
		"Footnotes":              blackfriday.Footnotes,
		"NoEmptyLineBeforeBlock": blackfriday.NoEmptyLineBeforeBlock,
		"HeadingIDs":             blackfriday.HeadingIDs,
		"Titleblock":             blackfriday.Titleblock,
		"AutoHeadingIDs":         blackfriday.AutoHeadingIDs,
		"BackslashLineBreak":     blackfriday.BackslashLineBreak,
		"DefinitionLists":        blackfriday.DefinitionLists,
	}

	markdownHTMLOptions = map[string]blackfriday.HTMLFlags{
		// Legacy names:
		"HTML_SKIP_HTML":                blackfriday.SkipHTML,
		"HTML_SKIP_IMAGES":              blackfriday.SkipImages,
		"HTML_SKIP_LINKS":               blackfriday.SkipLinks,
		"HTML_SAFELINK":                 blackfriday.Safelink,
		"HTML_NOFOLLOW_LINKS":           blackfriday.NofollowLinks,
		"HTML_HREF_TARGET_BLANK":        blackfriday.HrefTargetBlank,
		"HTML_TOC":                      blackfriday.TOC,
		"HTML_COMPLETE_PAGE":            blackfriday.CompletePage,
		"HTML_USE_XHTML":                blackfriday.UseXHTML,
		"HTML_USE_SMARTYPANTS":          blackfriday.Smartypants,
		"HTML_SMARTYPANTS_FRACTIONS":    blackfriday.SmartypantsFractions,
		"HTML_SMARTYPANTS_LATEX_DASHES": blackfriday.SmartypantsLatexDashes,
		"HTML_FOOTNOTE_RETURN_LINKS":    blackfriday.FootnoteReturnLinks,

		"HTMLFlagsNone":           blackfriday.HTMLFlagsNone,
		"SkipHTML":                blackfriday.SkipHTML,
		"SkipImages":              blackfriday.SkipImages,
		"SkipLinks":               blackfriday.SkipLinks,
		"Safelink":                blackfriday.Safelink,
		"NofollowLinks":           blackfriday.NofollowLinks,
		"NoreferrerLinks":         blackfriday.NoreferrerLinks,
		"NoopenerLinks":           blackfriday.NoopenerLinks,
		"HrefTargetBlank":         blackfriday.HrefTargetBlank,
		"CompletePage":            blackfriday.CompletePage,
		"UseXHTML":                blackfriday.UseXHTML,
		"FootnoteReturnLinks":     blackfriday.FootnoteReturnLinks,
		"Smartypants":             blackfriday.Smartypants,
		"SmartypantsFractions":    blackfriday.SmartypantsFractions,
		"SmartypantsDashes":       blackfriday.SmartypantsDashes,
		"SmartypantsLatexDashes":  blackfriday.SmartypantsLatexDashes,
		"SmartypantsAngledQuotes": blackfriday.SmartypantsAngledQuotes,
		"SmartypantsQuotesNBSP":   blackfriday.SmartypantsQuotesNBSP,
		"TOC":                     blackfriday.TOC,
	}
)
