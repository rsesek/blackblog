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
	"io"
	"path"
	"os"
	"testing"
)

func TestCopyDir(t *testing.T) {
	src := "./tests/"
	dest := path.Join(os.TempDir(), "TestCopyDir_Dest")
	defer os.RemoveAll(dest)

	err := copyDir(dest, src)
	if err != nil {
		t.Fatalf("Error copying directory: %v", err)
	}

	f, err := os.Open(path.Join(dest, "recurse", "copy_test"))
	if err != nil {
		t.Fatalf("Cannot open test file: %v", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		t.Fatalf("Error Stat()ing file: %v", err)
	}

	if info.Mode() != 0754 {
		t.Errorf("Mode should be 0754, got %v", info.Mode())
	}

	data := make([]byte, 4)
	n, err := f.Read(data)
	if n != len(data) {
		t.Errorf("Data should be of length %d, not %d", len(data), n)
	}
	if err != nil {
		t.Errorf("Error should be nil, got %v", err)
	}

	n, err = f.Read(data)
	if n != 0 && err != io.EOF {
		t.Errorf("Should have gotten EOF, got %v", err)
	}

	contents := string(data)
	if contents != "Foo\n" {
		t.Errorf("Contents of file should be %q, got %q", "Foo", contents)
	}
}
