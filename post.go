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
	"bufio"
	"crypto/md5"
	"io"
	"os"
)

// A Post contains the metadata about a post.
type Post struct {
	// The full path to the Markdown file containing this post.
	Filepath string

	// The title of the post, from the metadata.
	Title string

	// The URL fragment of the post, from the metadata.
	URLFragment string

	// The date the post was published, from the metadata.
	Date string

	// The MD5 checksum of the file's contents.
	checksum []byte
}

// NewPostFromPath creates a new Post object from the file at the given path
// with the metadata updated.
func NewPostFromPath(path string) (*Post, os.Error) {
	p := &Post{Filepath: path}
	p.UpdateMetadata()
	if len(p.checksum) < 1 {
		return nil, os.NewError("Could not checksum blog post")
	}
	return p, nil
}

// Open returns an opened file handle for the Post.
func (p *Post) Open() (*os.File, os.Error) {
	return os.Open(p.Filepath)
}

// IsUpToDate checks that the in-memory data about the Post matches that of the
// file on disk.
func (p *Post) IsUpToDate() bool {
	file, err := p.Open()
	defer file.Close()
	if err != nil {
		return false
	}
	return p.isUpToDateInternal(file)
}

func (p *Post) isUpToDateInternal(file *os.File) bool {
	return false
}

func (p *Post) computeChecksum(file *os.File) []byte {
	digest := md5.New()
	io.Copy(digest, file)
	return digest.Sum()
}

// GetContents returns the Markdown content of a post, excluding metadata.
func (p *Post) GetContents() ([]byte, os.Error) {
	file, err := p.Open()
	defer file.Close()
	if err != nil {
		return nil, err
	}

	var result []byte
	wasPrefix := false
	reader := bufio.NewReader(file)
	for {
		line, isPrefix, err := reader.ReadLine()

		// If an error occurred, return the error except for EOF.
		if err != nil {
			if err == os.EOF {
				err = nil
			}
			return result, err
		}

		// Skip lines that are for metadata.
		if wasPrefix || len(line) >= 2 && string(line[0:2]) == "~~" {
			wasPrefix = isPrefix
			continue
		}

		if !isPrefix {
			line = append(line, '\n')
		}

		// Store the rest of the file.
		result = append(result, line...)
	}
	return result, nil
}

// UpdateMetadata re-reads the file on disk and updates the in-memory metadata.
func (p *Post) UpdateMetadata() {
	file, err := p.Open()
	defer file.Close()
	if err != nil || p.isUpToDateInternal(file) {
		return
	}
	p.checksum = p.computeChecksum(file)

	file.Seek(0, 0)
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err == os.EOF || len(line) < 2 || line[0:2] != "~~" {
			break
		}
	}
}
