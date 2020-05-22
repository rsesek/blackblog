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
	renderTypeRedirect                    // Link back to the root.
	renderTypeFeed                        // A PostList.
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

func (r *render) String() string {
	var t string
	switch r.t {
	case renderTypeInvalid:
		t = "Invalid"
	case renderTypePost:
		t = "Post"
	case renderTypeDirectory:
		t = "Dir"
	case renderTypeRedirect:
		t = "Redirect"
	case renderTypeFeed:
		t = "Feed"
	default:
		t = "???"
	}
	return fmt.Sprintf("render%s(%p){%v %p}", t, r, r.object, r.parent)
}

// createRenderTree takes a slice of posts and returns the root node of the
// renderTree.
func createRenderTree(posts PostList) (*render, error) {
	root := &render{
		t:      renderTypeDirectory,
		object: make(renderTree),
	}
	root.object.(renderTree)["feed.xml"] = &render{
		t:      renderTypeFeed,
		object: posts,
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
		parent: dir,
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

			// Since this is a subdirectory, accessing index.html should go up
			// to the root. Attach the new render to the newly created directory.
			link := ""
			for ln := node; ln.parent != nil; ln = ln.parent {
				link += "../"
			}
			node.object.(renderTree)["index.html"] = &render{
				t:      renderTypeRedirect,
				object: link,
			}
		}
	}

	return node, nil
}

func visitPosts(root *render) <-chan *Post {
	c := make(chan *Post)

	var visitor func(*render)
	visitor = func(render *render) {
		for _, child := range render.object.(renderTree) {
			if child.t == renderTypeDirectory {
				visitor(child)
			} else if child.t == renderTypePost {
				c <- child.object.(*Post)
			}
		}
	}

	go func() {
		visitor(root)
		close(c)
	}()
	return c
}

// nodeDepth returns the number of edges between the given render |r| and its
// root.
func nodeDepth(r *render) (i int) {
	for ; r.parent != nil; r = r.parent {
		i++
	}
	return
}

// depthPath returns a relative path to the root for a render |r|.
func depthPath(r *render) string {
	depth := nodeDepth(r) - 1
	if depth < 0 {
		depth = 0
	}
	return strings.Repeat("../", depth)
}
