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
