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
	"bytes"
	"crypto/md5"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// A Post contains the metadata about a post.
type Post struct {
	// The full path to the Markdown file containing this post.
	Filename string

	// The title of the post, from the metadata.
	Title string

	// The URL fragment of the post, from the metadata.
	URLFragment string

	// The date the post was published, from the metadata.
	Date       string
	dateParsed time.Time

	// The MD5 checksum of the file's contents.
	checksum []byte
}

// GetPostsInDirectory recursively examines the directory at the path and finds
// any Markdown (.md) files and returns the corresponding Post objects.
func GetPostsInDirectory(dirPath string) (posts PostList, err error) {
	err = filepath.Walk(dirPath, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
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

// NewPostFromPath creates a new Post object from the file at the given path
// with the metadata updated.
func NewPostFromPath(path string) (*Post, error) {
	p := &Post{Filename: path}
	p.UpdateMetadata()
	if len(p.checksum) < 1 {
		return nil, errors.New("Could not checksum blog post")
	}
	return p, nil
}

// Open returns an opened file handle for the Post.
func (p *Post) Open() (*os.File, error) {
	return os.Open(p.Filename)
}

// IsUpToDate checks that the in-memory data about the Post matches that of the
// file on disk.
func (p *Post) IsUpToDate() bool {
	file, err := p.Open()
	if err != nil {
		return false
	}
	defer file.Close()
	return p.isUpToDateInternal(file)
}

func (p *Post) isUpToDateInternal(file *os.File) bool {
	if p.checksum == nil {
		return false
	}
	return bytes.Equal(p.checksum, computeChecksum(file))
}

func computeChecksum(file *os.File) []byte {
	digest := md5.New()
	io.Copy(digest, file)
	return digest.Sum(nil)
}

// GetContents returns the Markdown content of a post, excluding metadata.
func (p *Post) GetContents() ([]byte, error) {
	return p.parse(parseContents)
}

// UpdateMetadata re-reads the file on disk and updates the in-memory metadata.
func (p *Post) UpdateMetadata() {
	p.parse(parseLazily)
}

type parseOptions uint

const (
	// parseLazily only reloads the metadata if the post is out-of-date.
	parseLazily parseOptions = iota
	// parseReloadMetadata unconditionally reloads the post metadata.
	parseReloadMetadata
	// parseContents unconditionally reloads the post metadata and returns
	// the post contents.
	parseContents
)

// parse parses the Post according to `opts`. Returns the post contents if
// `opts` is `parseContents`, otherwise only returns an error if one occurs.
func (p *Post) parse(opts parseOptions) ([]byte, error) {
	file, err := p.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if opts == parseLazily && p.isUpToDateInternal(file) {
		return nil, nil
	}

	p.checksum = computeChecksum(file)
	file.Seek(0, 0)

	var contents []byte
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')

		// If an error occurred, return the error except for EOF.
		if err != nil && err != io.EOF {
			return nil, err
		}

		// Skip lines that are for metadata.
		if len(line) >= 2 && string(line[0:2]) == "~~" {
			p.parseMetadataLine(line)
			continue
		}

		if opts == parseReloadMetadata {
			return nil, nil
		}

		// Store the rest of the file.
		contents = append(contents, line...)

		// If this is the end of file, exit the loop to return the result.
		if err == io.EOF {
			break
		}
	}
	return contents, nil
}

func (p *Post) parseMetadataLine(line string) {
	if len(line) < 2 || line[0:2] != "~~" {
		return
	}
	line = line[2:]

	pieces := strings.SplitN(line, ":", 2)
	if len(pieces) != 2 {
		return
	}

	val := strings.TrimSpace(pieces[1])

	switch strings.ToLower(strings.TrimSpace(pieces[0])) {
	case "title":
		p.Title = val
	case "url":
		p.URLFragment = val
	case "date":
		p.Date = val
	}
}

// CreateURL constructs the URL of a post based on its metadata.
func (p *Post) CreateURL() string {
	// First, create the file's basename.
	basename := p.URLFragment
	if basename == "" && p.Title != "" {
		basename = regexp.MustCompile("[^A-Za-z0-9_]").ReplaceAllString(p.Title, "_")
		basename = regexp.MustCompile("_{1,}").ReplaceAllString(basename, "_")
	} else if basename == "" {
		basename = path.Base(p.Filename)
		ext := path.Ext(basename)
		basename = basename[:len(basename)-len(ext)]
	}

	basename = strings.ToLower(basename)
	if strings.HasSuffix(basename, "_") {
		basename = basename[:len(basename)-1]
	} else if strings.HasSuffix(basename, "/") {
		basename += "index"
	}
	url := basename + ".html"

	// Next, try and get the date of the post to include subdirectories.
	if p.dateParsed = parseDate(p.Date); !p.dateParsed.IsZero() {
		year, month, _ := p.dateParsed.Date()
		url = path.Join(strconv.FormatInt(int64(year), 10), strconv.Itoa(int(month)), url)
	}

	return url
}

func (p *Post) CreatePermalink(b *Blog) string {
	component := p.CreateURL()
	base := b.URL()
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	return base + component
}

func (p *Post) GetDate() *time.Time {
	if p.Date == "" {
		return nil
	}
	return &p.dateParsed
}

func parseDate(input string) time.Time {
	if input == "" {
		return time.Time{}
	}

	t, err := time.Parse("2006-01-02", input)
	if err == nil {
		return t
	}

	t, err = time.Parse("_2 January 2006", input)
	if err == nil {
		return t
	}

	t, err = time.Parse("January _2, 2006", input)
	if err == nil {
		return t
	}

	t, err = time.Parse("January _2 2006", input)
	if err == nil {
		return t
	}

	return time.Time{}
}

// sort.Interface implementation:

type PostList []*Post

func (pl PostList) Len() int {
	return len(pl)
}

func (pl PostList) Less(i, j int) bool {
	// Call CreateURL first to ensure the date is parsed.
	ui, uj := pl[i].CreateURL(), pl[j].CreateURL()
	di, dj := pl[i].dateParsed, pl[j].dateParsed

	if !di.IsZero() && !dj.IsZero() {
		return di.Before(dj)
	}
	return ui < uj
}

func (pl PostList) Swap(i, j int) {
	pl[i], pl[j] = pl[j], pl[i]
}
