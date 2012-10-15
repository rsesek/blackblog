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
	"errors"
	"net/http"
)

var (
	serverPort = flag.Int("port", 0, "The port on which the standalone HTTP server will run.")
)

type blogServer struct {
	r *render
}

func (b *blogServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "%v\n", b.r)
}

func RunAsServer() bool {
	return *serverPort != 0
}

func StartBlogServer(posts []*Post) error {
	if !RunAsServer() {
		return errors.New("No --port specified to start the server")
	}

	root, err := createRenderTree(posts)
	if err != nil {
		return err
	}

	fmt.Printf("Starting blog server on port %d", *serverPort)
	return http.ListenAndServe(fmt.Sprintf(":%d", *serverPort), &blogServer{root})
}

func newBlogServer(r *render) http.Handler {
	return &blogServer{
		r: r,
	}
}
