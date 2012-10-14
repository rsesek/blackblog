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
	"os"
	"path"
	"strings"
)

type renderType int

const (
	renderTypeInvalid   renderType = iota // Invalid render type.
	renderTypePost                        // A Post object.
	renderTypeDirectory                   // A renderTree.
)

// A renderTree maps a URL fragment to a render object for the current level in
// the tree.
type renderTree map[string]*render

// render represents some element in the renderTree.
type render struct {
	t      renderType
	object interface{}
	parent *render
}

// createRenderTree takes a slice of posts and returns the root node of the
// renderTree.
func createRenderTree(posts []*Post) (*render, error) {
	root := &render{
		t:      renderTypeDirectory,
		object: make(renderTree),
	}
	for _, p := range posts {
		if err := insertPost(p, root); err != nil {
			return nil, err
		}
	}
	return root, nil
}

// insertPost places the given post into the renderTree root at its appropriate
// depth for the URL.
func insertPost(post *Post, root *render) error {
	url := post.CreateURL()
	dir, err := findOrCreateDirNode(url, root)
	if err != nil {
		return err
	}

	filename := path.Base(url)
	dir.object.(renderTree)[filename] = &render{
		t:      renderTypePost,
		object: post,
	}
	return nil
}

func findOrCreateDirNode(url string, root *render) (*render, error) {
	parts := strings.Split(url, string(os.PathSeparator))

	// Loop over the parts of the URL, finding or creating a directory node for
	// each path component.
	node := root
	for _, part := range parts[0 : len(parts)-1] {
		// Test if this part is already in the tree.
		if child, ok := node.object.(renderTree)[part]; ok {
			// The part is in the tree, make sure it is a directory node.
			if child.t == renderTypeDirectory {
				node = child
			} else {
				return nil, fmt.Errorf("trying to find dir for %q, encountered non-directory node in path", url)
			}
		} else {
			// A node was not found here, so create it.
			rt := node.object.(renderTree)
			rt[part] = &render{
				t:      renderTypeDirectory,
				object: make(renderTree),
				parent: node,
			}
			node = rt[part]
		}
	}

	return node, nil
}

// writeRenderTree takes a root render object and writes out a rendered site
// to the given destination path.
func writeRenderTree(dest string, root *render) error {
	if root.t != renderTypeDirectory {
		return fmt.Errorf("writeRenderTree for %q: not a directory", dest)
	}

	// Iterate over this renderTree's subnodes.
	for part, render := range root.object.(renderTree) {
		p := path.Join(dest, part)
		if render.t == renderTypeDirectory {
			// For directories, ensure that the parent directory exists. If it
			// does not, create it and add a redirect index.html file.
			if err := os.Mkdir(p, 0755); err != nil && !os.IsExist(err) {
				return err
			}
			createRedirectFile(p, func() (depth int) {
				for n := render; n.parent != nil; n = n.parent {
					depth++
				}
				return
			}())
			// Recurse on its subnodes.
			writeRenderTree(p, render)
		} else if render.t == renderTypePost {
			// For posts, just render the content into the template.
			post := render.object.(*Post)
			content, err := post.GetContents()
			if err != nil {
				return err
			}

			html := RenderPost(post, content)
			f, err := os.Create(p)
			if err != nil {
				return err
			}
			f.Write(html)
			f.Close()
		}
	}

	return nil
}
