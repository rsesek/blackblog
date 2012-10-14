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
	"strings"
	"text/template"

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
func CreateIndex(filepath string, postMap PostURLMap, sortOrder []string) {
	tpl, err := getTemplate("index")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating index.html: %v\n", err)
	}

	posts := make([]map[string]string, len(sortOrder))
	for i, url := range sortOrder {
		posts[i] = map[string]string{
			"URL":   url,
			"Date":  postMap[url].Date,
			"Title": postMap[url].Title,
		}
	}

	buf := bytes.NewBuffer([]byte{})
	tpl.Execute(buf, map[string]interface{}{"Posts": posts})

	fd, err := os.Create(filepath)
	defer fd.Close()
	if err != nil {
		return
	}

	content, err := wrapPage(buf.Bytes(), map[string]string{
		"Title":    "Posts",
		"RootPath": "",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating index.html: %v\n", err)
		return
	}
	fd.Write(content)
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

// createRedirectFile creates a file index.html at |at| that redirects up
// |depth| levels.
func createRedirectFile(at string, depth int) error {
	url := strings.Repeat("../", depth)
	content := fmt.Sprintf(`<html><head><meta http-equiv="refresh" content="0;url=%s"></head></html>`, url)
	f, err := os.Create(path.Join(at, "index.html"))
	if err != nil {
		return err
	}
	fmt.Fprint(f, content)
	f.Close()
	return nil
}
