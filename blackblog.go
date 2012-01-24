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
	"flag"
	"io/ioutil"
	"os"

	"github.com/russross/blackfriday"
)

var (
	flagSource = flag.String("root", "", "The root directory of all Markdown posts")
	flagPort = flag.String("port", "", "The port to bind to for running the standalone HTTP server")
	flagDest = flag.String("dest", "", "The output directory for running in comiple mode")
)

func main() {
	flag.Parse()

	if *flagSource == "" {
		fmt.Fprintf(os.Stderr, "No -root blog directory specified\n")
		os.Exit(1)
	}

	if *flagPort == "" && *flagDest == "" {
		fmt.Fprintf(os.Stderr, "No -port or -dest flag specified\n")
		os.Exit(2)
	}

	fd, err := os.Open(*flagSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.String())
		os.Exit(3)
	}

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.String())
		os.Exit(4)
	}


	output := blackfriday.Markdown(
		data,
		blackfriday.HtmlRenderer(
			blackfriday.HTML_USE_SMARTYPANTS |
				blackfriday.HTML_SMARTYPANTS_LATEX_DASHES,
			"",
			""),
		0)
	fmt.Printf("%s", output)
}
