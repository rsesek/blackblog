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
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"text/template"

	"github.com/russross/blackfriday"
)

// RenderPost runs the input source through the blackfriday library.
func RenderPost(post *Post, input []byte, page PageParams) []byte {
	tpl, err := page.getTemplate("post")
	if err != nil {
		return nil
	}

	content := blackfriday.Markdown(
		input,
		blackfriday.HtmlRenderer(
			blackfriday.HTML_USE_SMARTYPANTS|
				blackfriday.HTML_USE_XHTML|
				blackfriday.HTML_SMARTYPANTS_LATEX_DASHES,
			"",
			""),
		0)

	page.Title = post.Title
	params := PostPageParams{
		Post:       post,
		Content:    string(content),
		PageParams: page,
	}

	buf := bytes.NewBuffer([]byte{})
	tpl.Execute(buf, params)

	result, err := wrapPage(buf.Bytes(), params.PageParams)
	return result
}

func makeParentDirIfNecessary(dir string) error {
	parent, _ := path.Split(dir)
	return os.MkdirAll(parent, 0755)
}

// CreateIndex takes the sorted list of posts and generates HTML output listing
// each one.
func CreateIndex(posts PostList, page PageParams) ([]byte, error) {
	tpl, err := page.getTemplate("index")
	if err != nil {
		return nil, err
	}

	sort.Sort(posts)

	page.Title = "Posts"
	params := IndexPageParams{
		Posts:      posts,
		PageParams: page,
	}

	buf := bytes.NewBuffer([]byte{})
	err = tpl.Execute(buf, params)
	if err != nil {
		return nil, err
	}

	content, err := wrapPage(buf.Bytes(), params.PageParams)
	return content, nil
}

// PageParams contains the varaibles passed to the basic header/footer
// page templates.
type PageParams struct {
	// The blog configuration object.
	Blog *Blog

	// The title of the blog post.
	Title string

	// Relative path linking up to the root of the blog.
	RootPath string
}

// The directory in which static files live.
const StaticFilesDir = "/static/"

// IndexPageParams is used to render out the blog post list page.
type IndexPageParams struct {
	PageParams
	Posts PostList
}

// PostPageParams is used for displaying a rendered post.
type PostPageParams struct {
	PageParams
	Post    *Post
	Content string // The HTML content rendered from (*Post).GetContents() markdown original.
}

func wrapPage(content []byte, vars PageParams) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})

	header, err := vars.getTemplate("header")
	if err != nil {
		return nil, err
	}

	footer, err := vars.getTemplate("footer")
	if err != nil {
		return nil, err
	}

	header.Execute(buf, vars)
	buf.Write(content)
	footer.Execute(buf, vars)

	return buf.Bytes(), nil
}

func (p *PageParams) getTemplate(name string) (*template.Template, error) {
	name = path.Join(p.Blog.TemplatesDir, name+".html")
	file, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	tpl := template.New(name)
	_, err = tpl.Parse(string(file))
	if err != nil {
		return nil, err
	}

	return tpl, nil
}

func generateRedirect(url string) string {
	return fmt.Sprintf(`<html><head><meta http-equiv="refresh" content="0;url=%s"></head></html>`, url)
}
