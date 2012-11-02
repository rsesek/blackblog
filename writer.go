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
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

// WriteStaticBlog takes a given blog and renders its output as static HTML
// files, according to the configuration.
func WriteStaticBlog(blog *Blog) error {
	posts, err := GetPostsInDirectory(blog.GetPostsDir())
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
		f, err = os.Create(path.Join(blog.GetOutputDir(), "index.html"))
	}
	if err != nil {
		return errors.New("Creating index: " + err.Error())
	}
	defer f.Close()
	f.Write(index)
	return nil
}

// writeRenderTree takes a root render object and writes out a rendered site
// to the given destination path.
func writeRenderTree(blog *Blog, root *render) error {
	if root.t != renderTypeDirectory {
		return fmt.Errorf("writeRenderTree for %q: not a directory", blog.OutputDir)
	}

	// Iterate over this renderTree's subnodes.
	for part, render := range root.object.(renderTree) {
		p := path.Join(blog.GetOutputDir(), part)
		switch render.t {
		case renderTypeDirectory:
			// For directories, ensure that the parent directory exists. If it
			// does not, create it and add a redirect index.html file.
			if err := os.Mkdir(p, 0755); err != nil && !os.IsExist(err) {
				return err
			}
			// Recurse on its subnodes.
			if err := writeRenderTree(blog, render); err != nil {
				return err
			}
		case renderTypePost:
			// For posts, just render the content into the template.
			post := render.object.(*Post)
			content, err := post.GetContents()
			if err != nil {
				return err
			}

			// Try to render the post.
			html, err := RenderPost(post, content, PageParams{
				Blog:     blog,
				RootPath: depthPath(render),
			})
			if err != nil {
				return err
			}

			// Try to write the post.
			f, err := os.Create(p)
			if err != nil {
				return err
			}
			f.Write(html)
			f.Close()
		case renderTypeRedirect:
			f, err := os.Create(p)
			if err != nil {
				return err
			}
			fmt.Fprint(f, generateRedirect(render.object.(string)))
			f.Close()
		default:
			return fmt.Errorf("writeRenderTree for %q: unknown renderType %v", p, render.t)
		}
	}

	return nil
}

// copyDir dittos the source directory tree to the destination.
func copyDir(dst, src string) error {
	// Make sure the destination exists.
	if err := os.Mkdir(dst, 0755); err != nil && !os.IsExist(err) {
		return err
	}

	src = path.Clean(src)
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Strip off the source directory prefix.
		newP := path.Join(dst, p[len(src):])

		if info.IsDir() {
			if err := os.Mkdir(newP, 0755); !os.IsExist(err) {
				return err
			}
		} else {
			df, err := os.OpenFile(newP, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
			if err != nil {
				return err
			}
			defer df.Close()

			sf, err := os.Open(p)
			if err != nil {
				return err
			}
			defer sf.Close()

			if _, err := io.Copy(df, sf); err != nil {
				return err
			}
		}

		return nil
	})
}
