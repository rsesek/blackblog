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
	"flag"
	"fmt"
	"os"
	"path"
	"runtime"
	"text/template"
	"time"
)

var (
	// Flags that allow overriding configuration defaults.
	serverPort = flag.Int("port", 0, "Override the port on which the standalone HTTP server will run.")
	outputDir  = flag.String("output", "", "Override the output directory when rendering to static files.")

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
	if err != nil && args[0] != cmdNewBlog {
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
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: blackblog newblog path/for/blog")
			os.Exit(3)
		}
		if err := newBlog(args[1]); err != nil {
			fmt.Fprintln(os.Stderr, "Error creating new blog:", err)
			os.Exit(3)
		}
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

func newBlog(at string) error {
	if err := os.Mkdir(at, 0755); err != nil {
		return err
	}

	if err := os.Mkdir(path.Join(at, "posts"), 0755); err != nil {
		return err
	}
	if err := os.Mkdir(path.Join(at, "out"), 0755); err != nil {
		return err
	}

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return errors.New("cannot find installation directory")
	}

	data := struct {
		InstallDir string
		Date       time.Time
	}{
		InstallDir: path.Dir(thisFile),
		Date:       time.Now(),
	}

	config, err := template.New("config").Parse(defaultConfig)
	if err != nil {
		return err
	}

	post, err := template.New("post").Parse(defaultPost)
	if err != nil {
		return err
	}

	f, err := os.Create(path.Join(at, ConfigFileName))
	if err == nil {
		err = config.Execute(f, data)
	}
	f.Close()

	if err != nil {
		return err
	}

	f, err = os.Create(path.Join(at, "posts", "welcome.md"))
	if err == nil {
		err = post.Execute(f, data)
	}
	f.Close()
	return err
}

var (
	defaultConfig = `{
	"Title": "A Black Blog",
	"PostsDir": "./posts",
	"TemplatesDir": "{{.InstallDir}}/templates",
	"StaticFilesDir": "{{.InstallDir}}/templates/static",
	"OutputDir": "./out",
	"Port": 8066
}`

	defaultPost = `~~~ Title: Welcome to Blackblog
~~~ Date: {{.Date.Format "January _2 2006"}}
~~~ URL: welcome

This is the first and only post in your Blackblog. Feel free to delete it.`
)
