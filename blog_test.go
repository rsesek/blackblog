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

func verifyConfig(b *Blog, t *testing.T) {
	if b == nil {
		t.Fatalf("Blog should not be nil")
	}

	expectations := []struct {
		actual, expected, field string
	}{
		{b.Title, "Head of a Cow", "Title"},
		{b.PostsDir, "./", "PostsDir"},
		{b.TemplatesDir, "../templates", "TemplatesDir"},
		{b.StaticFilesDir, "", "StaticFilesDir"},
		{b.OutputDir, "../out/", "OutputDir"},
		{b.configPath, "./tests/blackblog.json", "configPath"},
	}

	for _, e := range expectations {
		if e.actual != e.expected {
			t.Errorf("%s should be %q, got %q", e.field, e.expected, e.actual)
		}
	}

	port := 8066
	if b.Port != port {
		t.Errorf("Port should be %d, got %d", port, b.Port)
	}
}

func TestReadWithoutSuffix(t *testing.T) {
	b, e := ReadBlog("./tests")
	if e != nil {
		t.Fatalf("Unexpected error reading blog: %v", e)
	}

	verifyConfig(b, t)
}

func TestReadWithSuffix(t *testing.T) {
	b, e := ReadBlog("./tests/blackblog.json")
	if e != nil {
		t.Fatalf("Unexpected error reading blog: %v", e)
	}

	verifyConfig(b, t)
}
