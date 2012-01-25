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
	"fmt"
	"flag"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/russross/blackfriday"
)

var (
	flagSource = flag.String("root", "", "The root directory of all Markdown posts")
	flagPort = flag.String("port", "", "The port to bind to for running the standalone HTTP server")
	flagDest = flag.String("dest", "", "The output directory for running in comiple mode")
)

func main() {
	flag.Parse()

	if *flagSource == "" {
		fmt.Fprintf(os.Stderr, "No -root blog directory specified\n")
		os.Exit(1)
	}

	if *flagPort == "" && *flagDest == "" {
		fmt.Fprintf(os.Stderr, "No -port or -dest flag specified\n")
		os.Exit(2)
	}

	post, err := NewPostFromPath(*flagSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.String())
		os.Exit(3)
	}

	data, err := post.GetContents()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.String())
		os.Exit(4)
	}

	output := blackfriday.Markdown(
		data,
		blackfriday.HtmlRenderer(
			blackfriday.HTML_USE_SMARTYPANTS |
				blackfriday.HTML_SMARTYPANTS_LATEX_DASHES,
			"",
			""),
		0)
	fmt.Printf("%s", output)
}

// GetPostsInDirectory recursively examines the directory at the path and finds
// any Markdown (.md) files and returns the corresponding Post objects.
func GetPostsInDirectory(dirPath string) []*Post {
	fd, err := os.Open(dirPath)
	defer fd.Close()
	if err != nil {
		return nil
	}

	files, err := fd.Readdir(-1)
	if err != nil {
		return nil
	}

	var results []*Post
	for _, file := range files {
		filePath := path.Join(dirPath, file.Name)
		if file.IsDirectory() {
			subfiles := GetPostsInDirectory(filePath)
			if subfiles != nil {
				results = append(results, subfiles...)
			}
		} else if strings.HasSuffix(file.Name, ".md") {
			if post, err := NewPostFromPath(filePath); post != nil && err == nil {
				results = append(results, post)
			}
		}
	}

	return results
}

// PostURLMap keys Post objects by their final URL.
type PostURLMap map[string]*Post

// SortPosts creates URLs for each Post and returns a map that links the URL to
// the post and a slice of the URLs in sorted order.
func SortPosts(posts []*Post) (postMap PostURLMap, sorted []string) {
	postMap = make(PostURLMap, len(posts))
	for _, post := range posts {
		url := post.CreateURL()
		postMap[url] = post
		sorted = append(sorted, url)
	}
	sort.Strings(sorted)
	return
}
