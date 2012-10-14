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
	if testPost, ok := postRender.object.(*Post); !ok {
		t.Errorf("Test post should be a post, got %v", testPost)
	} else if testPost != post {
		t.Errorf("Post should equal %v, got %v", post, testPost)
	}
}
