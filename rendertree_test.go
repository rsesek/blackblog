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
)

func TestRootTree(t *testing.T) {
	one := &Post{URLFragment: "test_post"}
	root, err := createRenderTree([]*Post{one})

	if err != nil {
		t.Fatal("Unexpected error creating render tree", err)
	}

	if root.t != renderTypeDirectory {
		t.Fatalf("Root should be of type %v, got %v", renderTypeDirectory, root.t)
	}

	if root.parent != nil {
		t.Errorf("Root should not have a parent, has %v", root.parent)
	}

	contents, ok := root.object.(renderTree)
	if !ok {
		t.Errorf("Root object should be a renderTree, is %v", root.object)
	}

	if len(contents) != 1 {
		t.Errorf("Root's renderTree should have 1 object, has %d", len(contents))
	}

	for k, v := range contents {
		e := "test_post.html"
		if k != e {
			t.Errorf("Single key should be %q, got %q", e, k)
		}
		if v.object != one {
			t.Errorf("Single value should be %v, got %v", one, v)
		}
		if v.parent != root {
			t.Errorf("Single value parent should be %v, got %v", root, v.parent)
		}
	}
}

func TestTwoDirs(t *testing.T) {
	post := &Post{URLFragment: "test_post", Date: "14 October 2012"}
	root, err := createRenderTree([]*Post{post})

	if err != nil {
		t.Fatal("Unexpected error creating render tree", err)
	}

	contents, ok := root.object.(renderTree)
	if !ok {
		t.Errorf("Root object should be a render tree, is %v", root.object)
	}

	if len(contents) != 1 {
		t.Errorf("Root's renderTree should have 1 object, has %d", len(contents))
	}

	year, ok := contents["2012"]
	if !ok {
		t.Fatalf("Year directory not present")
	}

	if year.t != renderTypeDirectory {
		t.Errorf("Year should be a directory, is %v", year.t)
	}
	if year.parent != root {
		t.Errorf("Year should have parent %v, is %v", root, year.parent)
	}
	contents, ok = year.object.(renderTree)
	if !ok {
		t.Fatalf("Year should be a renderTree, got %v", contents)
	}

	month, ok := contents["10"]
	if !ok {
		t.Fatalf("Month directory not present")
	}

	if month.t != renderTypeDirectory {
		t.Errorf("Month should be a directory, is %v", month.t)
	}
	if month.parent != year {
		t.Errorf("Month should have parent %v, is %v", year, month.parent)
	}
	contents, ok = month.object.(renderTree)
	if !ok {
		t.Fatalf("Month should be a renderTree, got %v", contents)
	}

	postRender, ok := contents["test_post.html"]
	if !ok {
		t.Fatalf("Test post not present")
	}
	if postRender.t != renderTypePost {
		t.Errorf("Test post should be a post, got %v", postRender.t)
	}
	if postRender.parent != month {
		t.Errorf("Test post should have parent %v, is %v", month, postRender.parent)
	}
	if testPost, ok := postRender.object.(*Post); !ok {
		t.Errorf("Test post should be a post, got %v", testPost)
	} else if testPost != post {
		t.Errorf("Post should equal %v, got %v", post, testPost)
	}
}

func TestVisitor(t *testing.T) {
	root := &render{
		t: renderTypeDirectory,
		object: renderTree{
			"a": &render{
				t:      renderTypePost,
				object: &Post{Title: "First"},
			},
			"b": &render{
				t:      renderTypePost,
				object: &Post{Title: "Second"},
			},
			"c": &render{
				t: renderTypeDirectory,
				object: renderTree{
					"d": &render{t: renderTypeRedirect},
					"e": &render{
						t:      renderTypePost,
						object: &Post{Title: "Third"},
					},
					"f": &render{
						t: renderTypeDirectory,
						object: renderTree{
							"g": &render{
								t:      renderTypePost,
								object: &Post{Title: "Fourth"},
							},
						},
					},
				},
			},
		},
	}

	expectation := map[string]bool{
		"First":  false,
		"Second": false,
		"Third":  false,
		"Fourth": false,
	}

	c := visitPosts(root)
	for p := range c {
		expectation[p.Title] = true
	}

	for k, v := range expectation {
		if !v {
			t.Errorf("Did not visit post with title %q", k)
		}
	}
}

func TestNodeDepth(t *testing.T) {
	r := &render{}
	d := nodeDepth(r)

	if d != 0 {
		t.Errorf("Node depth for root should be 0, got %d", d)
	}

	c1 := &render{parent: r}
	d = nodeDepth(c1)
	if d != 1 {
		t.Errorf("Node depth for c1 should be 1, got %d", d)
	}

	c2 := &render{parent: c1}
	d = nodeDepth(c2)
	if d != 2 {
		t.Errorf("Node depth for c2 should be 2, got %d", d)
	}
}

func TestDepthPath(t *testing.T) {
	r := &render{}
	p := depthPath(r)
	e := ""

	if p != e {
		t.Errorf("Depth path for root should be %q, got %q", e, p)
	}

	c1 := &render{parent: r}
	p = depthPath(c1)
	e = "../"
	if p != e {
		t.Errorf("Depth path for c1 should be %q, got %q", e, p)
	}

	c2 := &render{parent: c1}
	p = depthPath(c2)
	e = "../../"
	if p != e {
		t.Errorf("Depth path for c2 should be %q, got %q", e, p)
	}
}
