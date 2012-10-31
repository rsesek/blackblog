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
	"errors"
	"os"
	"path"
)

// WriteStaticBlog takes a given blog and renders its output as static HTML
// files, according to the configuration.
func WriteStaticBlog(blog *Blog) error {
	posts, err := GetPostsInDirectory(blog.PostsDir)
	if err != nil {
		return errors.New("Get posts: " + err.Error())
	}

	renderTree, err := createRenderTree(posts)
	if err != nil {
		return errors.New("Render posts:" + err.Error())
	}

	if err := writeRenderTree(blog, renderTree); err != nil {
		return errors.New("Write files: " + err.Error())
	}

	index, err := CreateIndex(posts, PageParams{Blog: blog})
	var f *os.File
	if err == nil {
		f, err = os.Create(path.Join(blog.OutputDir, "index.html"))
	}
	if err != nil {
		return errors.New("Creating index: " + err.Error())
	}
	defer f.Close()
	f.Write(index)
	return nil
}
