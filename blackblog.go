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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	// Flags that allow overriding configuration defaults.
	serverPort = flag.Int("port", 0, "Override the port on which the standalone HTTP server will run.")
	outputDir = flag.String("output", "", "Override the output directory when rendering to static files.")

	commandDocs = map[string]string{
		cmdNewBlog:      "Create a new blog with some sample data in the specified directory.",
		cmdServer:       "Run a standalone web server for the given blog.",
		cmdStaticOutput: "Render the blog out to static HTML files.",
	}
	commandOrder = []string{
		cmdNewBlog,
		cmdServer,
		cmdStaticOutput,
	}
)

const (
	cmdNewBlog      = "newblog"
	cmdServer       = "serve"
	cmdStaticOutput = "render"
)

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		usage()
		os.Exit(1)
	}

	// Load the blog configuration.
	blogPath, _ := os.Getwd()
	if len(args) >= 2 {
		blogPath = args[1]
	}

	blog, err := ReadBlog(blogPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not read blog configuration:", err)
		os.Exit(2)
	}

	// Process flags to override configuration values.
	if *serverPort != 0 {
		blog.Port = *serverPort
	}
	if *outputDir != "" {
		blog.OutputDir = *outputDir
	}

	// Execute the specified command.
	switch args[0] {
	case cmdNewBlog:
		fmt.Fprintf(os.Stderr, "NOT IMPLEMENTED")
		os.Exit(200)
	case cmdServer:
		if err := StartBlogServer(blog); err != nil {
			fmt.Fprintln(os.Stderr, "Could not start blog server:", err)
			os.Exit(3)
		}
	case cmdStaticOutput:
		if err := WriteStaticBlog(blog); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s command [path/to/blog]:\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  Commands:\n")
	for _, cmdName := range commandOrder {
		fmt.Fprintf(os.Stderr, "    %s\t%s\n", cmdName, commandDocs[cmdName])
	}
	fmt.Fprintf(os.Stderr, "\n  Flags:\n")
	flag.PrintDefaults()
}

// GetPostsInDirectory recursively examines the directory at the path and finds
// any Markdown (.md) files and returns the corresponding Post objects.
func GetPostsInDirectory(dirPath string) (posts PostList, err error) {
	err = filepath.Walk(dirPath, func(file string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(file, ".md") {
			if post, err := NewPostFromPath(file); post != nil && err == nil {
				posts = append(posts, post)
			} else {
				return err
			}
		}
		return nil
	})
	return
}
