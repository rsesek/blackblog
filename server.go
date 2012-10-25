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
	"net/http"
	"strings"
)

var (
	serverPort = flag.Int("port", 0, "The port on which the standalone HTTP server will run.")
)

type blogServer struct {
	posts PostList
	r     *render
}

func (b *blogServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	url := strings.Trim(req.URL.Path, "/")

	if url == "" {
		b.serveNode(rw, req, b.r)
		return
	}

	parts := strings.Split(url, "/")
	node := b.r
	for _, part := range parts {
		if child, ok := node.object.(renderTree)[part]; ok {
			node = child
		} else {
			http.NotFound(rw, req)
			return
		}
	}

	b.serveNode(rw, req, node)
}

func (b *blogServer) serveNode(rw http.ResponseWriter, req *http.Request, render *render) {
	switch render.t {
	case renderTypePost:
		post := render.object.(*Post)
		data, err := post.GetContents()
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			fmt.Fprint(rw, err.Error())
			return
		}
		content := RenderPost(post, data)
		rw.Write(content)
	case renderTypeRedirect:
		fallthrough
	case renderTypeDirectory:
		// The root element should generate a post list.
		if render.t == renderTypeDirectory && render.parent == nil {
			index, err := CreateIndex(b.posts)
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(rw, err.Error())
				return
			}
			rw.Write(index)
			return
		}

		// Other directories when accessed directly should fallback to the
		// redirect.
		if render.t == renderTypeDirectory {
			render = render.object.(renderTree)["index.html"]
		}

		http.Redirect(rw, req, render.object.(string), http.StatusMovedPermanently)
	default:
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "Unknown render: %v", render)
	}
}

func RunAsServer() bool {
	return *serverPort != 0
}

func StartBlogServer(posts PostList) error {
	if !RunAsServer() {
		return errors.New("No --port specified to start the server")
	}

	root, err := createRenderTree(posts)
	if err != nil {
		return err
	}

	fmt.Printf("Starting blog server on port %d\n", *serverPort)
	return http.ListenAndServe(fmt.Sprintf(":%d", *serverPort), &blogServer{
		posts: posts,
		r:     root,
	})
}

func newBlogServer(r *render) http.Handler {
	return &blogServer{
		r: r,
	}
}
