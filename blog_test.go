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
	"testing"

	"github.com/russross/blackfriday"
)

func verifyConfig(b *Blog, t *testing.T) {
	if b == nil {
		t.Fatalf("Blog should not be nil")
	}

	expectations := []struct {
		actual, expected, field string
	}{
		{b.Title, "Head of a Cow", "Title"},
		{b.PostsDir, "./", "PostsDir"},
		{b.TemplatesDir, "../templates", "TemplatesDir"},
		{b.StaticFilesDir, "../templates/static/", "StaticFilesDir"},
		{b.OutputDir, "../out/", "OutputDir"},
		{b.configPath, "tests/blackblog.json", "configPath"},
	}

	for _, e := range expectations {
		if e.actual != e.expected {
			t.Errorf("%s should be %q, got %q", e.field, e.expected, e.actual)
		}
	}

	port := 8066
	if b.Port != port {
		t.Errorf("Port should be %d, got %d", port, b.Port)
	}
}

func TestReadWithoutSuffix(t *testing.T) {
	b, e := ReadBlog("./tests")
	if e != nil {
		t.Fatalf("Unexpected error reading blog: %v", e)
	}

	verifyConfig(b, t)
}

func TestReadWithSuffix(t *testing.T) {
	b, e := ReadBlog("./tests/blackblog.json")
	if e != nil {
		t.Fatalf("Unexpected error reading blog: %v", e)
	}

	verifyConfig(b, t)
}

func TestGetPaths(t *testing.T) {
	blog := &Blog{
		configPath: "/abs/path/blackblog.json",
		OutputDir:  "../blog_out",
		PostsDir:   "./posts",
	}

	var a, e string

	e = "/abs/blog_out"
	a = blog.GetOutputDir()
	if a != e {
		t.Errorf("GetOutputDir() should return %q, got %q", e, a)
	}

	e = "/abs/path/posts"
	a = blog.GetPostsDir()
	if a != e {
		t.Errorf("GetPostsDir() should return %q, got %q", e, a)
	}
}

func TestExtensionsAndOptions(t *testing.T) {
	blog := &Blog{
		MarkdownExtensions:  []string{"EXTENSION_FOOTNOTES", "EXTENSION_NO_INTRA_EMPHASIS"},
		MarkdownHTMLOptions: []string{"HTML_USE_SMARTYPANTS", "HTML_USE_XHTML", "HTML_SMARTYPANTS_LATEX_DASHES", "HTML_SAFELINK", "HTML_TOC"},
	}
	blog.parseOptions()

	extensions := blog.GetMarkdownExtensions()
	if extensions != blackfriday.EXTENSION_FOOTNOTES|blackfriday.EXTENSION_NO_INTRA_EMPHASIS {
		t.Errorf("Incorrect extensions-to-flags conversion")
	}

	options := blog.GetMarkdownHTMLOptions()
	if options != blackfriday.HTML_USE_SMARTYPANTS|blackfriday.HTML_USE_XHTML|blackfriday.HTML_SMARTYPANTS_LATEX_DASHES|blackfriday.HTML_SAFELINK|blackfriday.HTML_TOC {
		t.Errorf("Incorrect HTML-options-to-flags conversion")
	}

	// If there are no HTML options in a blog, then use the old defaults.
	blog = &Blog{}
	blog.parseOptions()
	if blog.GetMarkdownExtensions() != 0 {
		t.Errorf("Default markdown extensions should be empty, got %#x", blog.GetMarkdownExtensions())
	}
	expected := blackfriday.HTML_USE_SMARTYPANTS | blackfriday.HTML_USE_XHTML | blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	if blog.GetMarkdownHTMLOptions() != expected {
		t.Errorf("Default markdown HTML options should be %#x, got %#x", expected, blog.GetMarkdownHTMLOptions())
	}
}
