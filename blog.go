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
	"encoding/json"
	"os"
	"path"
	"strings"
)

const ConfigFileName = "blackblog.json"

// Blog is a structure that contains the configuration of a blackblog. This is
// stored as a JSON file, in the blog root directory, named `blackblog.json`.
type Blog struct {
	// The name of the blog, used in page titles.
	Title string

	// Path to the directory containing the Markdown files used for posts.
	PostsDir string

	// Path to the templates directory, used to format the blog.
	TemplatesDir string

	// Static files that are copied to the OutputDir or that are served in server
	// mode to support the templates.
	StaticFilesDir string

	// When rendering the blog to static files, the directory to place the
	// output.
	OutputDir string

	// When running as a server, the port on which the server is bound.
	Port int

	// Path to the configuration file (including "blackblog.json").
	configPath string
}

func (b *Blog) GetPostsDir() string {
	return b.getPath(b.PostsDir)
}

func (b *Blog) GetOutputDir() string {
	return b.getPath(b.OutputDir)
}

func (b *Blog) getPath(part string) string {
	return path.Join(path.Dir(b.configPath), part)
}

// ReadBlog reads the blog configuration from the specified file path. This
// does not need to end in `blackblog.json`.
func ReadBlog(p string) (blog *Blog, err error) {
	if !strings.HasSuffix(p, ConfigFileName) {
		p = path.Join(p, ConfigFileName)
	}

	r, err := os.Open(p)
	if err != nil {
		return
	}

	blog = new(Blog)
	d := json.NewDecoder(r)
	if err = d.Decode(blog); err == nil {
		blog.configPath = path.Clean(p)
	}
	return
}
