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
	"text/template"
	"sort"

	"github.com/russross/blackfriday"
)

// RenderPost runs the input source through the blackfriday library.
func RenderPost(post *Post, input []byte) []byte {
	tpl, err := getTemplate("post")
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

	buf := bytes.NewBuffer([]byte{})
	tpl.Execute(buf, map[string]interface{}{
		"Post":    post,
		"Content": string(content),
	})

	result, err := wrapPage(buf.Bytes(), map[string]string{
		"Title":    post.Title,
		"RootPath": getRootPath(post.CreateURL()),
	})
	return result
}

func makeParentDirIfNecessary(dir string) error {
	parent, _ := path.Split(dir)
	return os.MkdirAll(parent, 0755)
}

// CreateIndex takes the sorted list of posts and generates HTML output listing
// each one.
func CreateIndex(posts PostList) ([]byte, error) {
	tpl, err := getTemplate("index")
	if err != nil {
		return nil, err
	}

	sort.Sort(posts)

	buf := bytes.NewBuffer([]byte{})
	err = tpl.Execute(buf, struct{Posts PostList}{posts})
	if err != nil {
		return nil, err
	}

	content, err := wrapPage(buf.Bytes(), map[string]string{
		"Title":    "Posts",
		"RootPath": "",
	})
	return content, nil
}

func wrapPage(content []byte, vars interface{}) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})

	header, err := getTemplate("header")
	if err != nil {
		return nil, err
	}

	footer, err := getTemplate("footer")
	if err != nil {
		return nil, err
	}

	header.Execute(buf, vars)
	buf.Write(content)
	footer.Execute(buf, vars)

	return buf.Bytes(), nil
}

func getTemplate(name string) (*template.Template, error) {
	name = path.Join(*flagTemplates, name+".html")
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
