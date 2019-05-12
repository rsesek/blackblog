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
	"path"
	"sort"
	"text/template"

	"github.com/russross/blackfriday"
)

// RenderPost runs the input source through the blackfriday library.
func RenderPost(post *Post, input []byte, page PageParams) ([]byte, error) {
	tpl, err := page.getTemplate("post")
	if err != nil {
		return nil, err
	}

	content := blackfriday.Markdown(
		input,
		blackfriday.HtmlRenderer(page.Blog.GetMarkdownHTMLOptions(), "", ""),
		page.Blog.GetMarkdownExtensions())

	page.Title = post.Title
	params := PostPageParams{
		Post:       post,
		Content:    string(content),
		PageParams: page,
	}

	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, params); err != nil {
		return nil, err
	}

	return wrapPage(buf.Bytes(), params.PageParams)
}

// CreateIndex takes the sorted list of posts and generates HTML output listing
// each one.
func CreateIndex(posts PostList, blog *Blog) ([]byte, error) {
	page := CreatePageParams(blog, nil)

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

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, params)
	if err != nil {
		return nil, err
	}

	return wrapPage(buf.Bytes(), params.PageParams)
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

	// Relative path to the page being rendered.
	URL string
}

// CreatePageParams sets up the parameters for PageParams.
func CreatePageParams(blog *Blog, render *render) PageParams {
	var url, rootPath string
	if render == nil {
		url = "index.html"
		rootPath = ""
	} else {
		post := render.object.(*Post)
		url = post.CreateURL()
		rootPath = depthPath(render)
	}
	return PageParams{
		Blog: blog,
		RootPath: rootPath,
		URL: url,
	}
}

// The directory in which static files live.
const StaticFilesDir = "/static/"

// StaticFileLink returns a link for a static file.
// BUG: This cannot be on type *PageParams for some reason. File a bug?
func (p PageParams) StaticFileLink(file string) string {
	// Strip of the prefix / for StaticFilesDir so that if rendering to files, it
	// does not create an absolute path.
	return path.Join(p.RootPath, StaticFilesDir[1:], file)
}

// IndexPageParams is used to render out the blog post list page.
type IndexPageParams struct {
	PageParams
	Posts PostList
}

func (p IndexPageParams) PostsDescending() PostList {
	return p.PostsDescendingLimit(-1)
}

func (p IndexPageParams) PostsAscending() PostList {
	return p.PostsAscendingLimit(-1)
}

func (p IndexPageParams) PostsDescendingLimit(limit int) PostList {
	sort.Sort(sort.Reverse(p.Posts))
	if limit < len(p.Posts) && limit > 0 {
		return p.Posts[:limit]
	}
	return p.Posts
}

func (p IndexPageParams) PostsAscendingLimit(limit int) PostList {
	sort.Sort(p.Posts)
	if limit < len(p.Posts) && limit > 0 {
		return p.Posts[:limit]
	}
	return p.Posts
}

// PostPageParams is used for displaying a rendered post.
type PostPageParams struct {
	PageParams
	Post    *Post
	Content string // The HTML content rendered from (*Post).GetContents() markdown original.
}

func wrapPage(content []byte, vars PageParams) ([]byte, error) {
	buf := new(bytes.Buffer)

	header, err := vars.getTemplate("header")
	if err != nil {
		return nil, err
	}

	footer, err := vars.getTemplate("footer")
	if err != nil {
		return nil, err
	}

	if err := header.Execute(buf, vars); err != nil {
		return nil, err
	}

	buf.Write(content)

	if err := footer.Execute(buf, vars); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (p *PageParams) getTemplate(name string) (*template.Template, error) {
	name = path.Join(p.Blog.TemplatesDir(), name+".html")
	return template.ParseFiles(name)
}

func generateRedirect(url string) string {
	return fmt.Sprintf(`<html><head><meta http-equiv="refresh" content="0;url=%s"></head></html>`, url)
}
