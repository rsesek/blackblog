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
	"sync"
	"time"
)

var (
	serverPort = flag.Int("port", 0, "The port on which the standalone HTTP server will run.")

	serverPollWait = flag.Int("server-poll-time", 60, "The time in seconds that the server waits before polling the directory for changes.")
)

type blogServer struct {
	root string // Path to the root of the blog.

	mu    *sync.RWMutex
	posts PostList
	r     *render
}

// RunAsServer checks if the program has been configured to run as a web server.
func RunAsServer() bool {
	return *serverPort != 0
}

// StartBlogServer runs the program's web server given the blog located
// at |blogRoot|.
func StartBlogServer(blogRoot string) error {
	if !RunAsServer() {
		return errors.New("No --port specified to start the server")
	}

	server := &blogServer{
		root: blogRoot,
		mu:   new(sync.RWMutex),
	}

	err := server.buildPosts()
	if err != nil {
		return err
	}
	go server.pollPostChanges()

	fmt.Printf("Starting blog server on port %d\n", *serverPort)
	return http.ListenAndServe(fmt.Sprintf(":%d", *serverPort), server)
}

func (b *blogServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	url := strings.Trim(req.URL.Path, "/")

	b.mu.RLock()
	defer b.mu.RUnlock()

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

func (b *blogServer) pollPostChanges() {
	for {
		time.Sleep(time.Duration(*serverPollWait) * time.Second)
		if err := b.buildPosts(); err != nil {
			panic(err.Error())
		}
	}
}

func (b *blogServer) buildPosts() (err error) {
	newPosts, err := GetPostsInDirectory(b.root)
	if err != nil {
		return
	}

	b.mu.RLock()
	rebuild := len(newPosts) != len(b.posts)
	if !rebuild {
		for _, p := range b.posts {
			if !p.IsUpToDate() {
				rebuild = true
				break
			}
		}
	}
	b.mu.RUnlock()

	if rebuild {
		b.mu.Lock()
		defer b.mu.Unlock()

		b.posts, err = GetPostsInDirectory(b.root)
		if err != nil {
			return
		}
		b.r, err = createRenderTree(b.posts)
		if err != nil {
			return
		}
	}
	return nil
}
