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
	"testing"
	"time"
)

func TestGetContents(t *testing.T) {
	post, _ := NewPostFromPath("./tests/simple_post.md")
	s, _ := post.GetContents()
	expected := "\nThis is a simple post!\n\nWith two lines.\n"
	if string(s) != expected {
		t.Errorf("simple_post contents incorrect. Expected '%s', got '%s'", expected, s)
	}
}

type parseMetadata struct {
	input string
	post  Post
}

func TestParseMetadataLine(t *testing.T) {
	var results = []parseMetadata{
		{"~~ Title: This is a title", Post{Title: "This is a title"}},
		{"~~ Title: ~Weird Data~", Post{Title: "~Weird Data~"}},
		{"~~ Unknown: Field", Post{}},
		{"~~ uRl: foo_bar.html", Post{URLFragment: "foo_bar.html"}},
		{"~~ Date: 12/13/1344", Post{Date: "12/13/1344"}},
		{"~~Date: 13 January 2012     ", Post{Date: "13 January 2012"}},
	}

	for _, r := range results {
		var p Post
		p.parseMetadataLine(r.input)
		rp := r.post
		if p.Title != rp.Title || p.Date != rp.Date || p.URLFragment != rp.URLFragment {
			t.Errorf("Parse error for input '%s', expected '%v', got '%v'", r.input, r.post, p)
		}
	}
}

func TestFullMetadata(t *testing.T) {
	post, err := NewPostFromPath("./tests/simple_post.md")
	if err != nil {
		t.Errorf("Error reading post: %v", err)
	}

	expected := "Simple Post"
	if post.Title != expected {
		t.Errorf("post.Title mismatch, expected '%s', got '%s'", expected, post.Title)
	}

	expected = "simple_post"
	if post.URLFragment != expected {
		t.Errorf("post.URLFragment mismatch, expected '%s', got '%s'", expected, post.URLFragment)
	}

	expected = "24 January 2012"
	if post.Date != expected {
		t.Errorf("post.Date mismatch, expected '%s', got '%s'", expected, post.Date)
	}
}

func TestIsOutOfDate(t *testing.T) {
	post, err := NewPostFromPath("./tests/update_test.md")
	if err != nil {
		t.Errorf("Error reading post: %v", err)
	}

	if !post.IsUpToDate() {
		t.Errorf("Post %s is unexpectedly out of date", post.Filename)
	}

	post.Filename = "./tests/update_test_out_of_date.md"
	if post.IsUpToDate() {
		t.Errorf("Post %s is unexpectedly up-to-date", post.Filename)
	}
}

type createURL struct {
	url  string
	post Post
}

func TestCreateURL(t *testing.T) {
	results := []createURL{
		{"2012/1/test.html", Post{URLFragment: "test", Date: "25 January 2012"}},
		{"2012/12/test.html.html", Post{URLFragment: "test.html", Date: "12 December 2012"}},
		{"2012/1/foobar.html", Post{Title: "Foobar", Date: "1 January 2012"}},
		{"2012/4/some_post.html", Post{Title: "Some Post", Date: "4 April 2012"}},
		{"2012/3/a_post.html", Post{Filename: "some/path/a_post.md", Date: "March 3, 2012"}},
		{"2013/4/nyc_meetup_april_2013.html", Post{Title: "NYC Meetup, April 2013", Date: "April 16, 2013"}},
		{"escaped_test_page.html", Post{Title: `Escaped, Test! "Page"`}},
		{"test_post.html", Post{Filename: "/some/test/test_post.md"}},
		{"test_test_test.html", Post{Title: "Test tEsT TEST"}},
		{"foobar.html", Post{URLFragment: "foobar"}},
	}

	for _, r := range results {
		actual := r.post.CreateURL()
		if r.url != actual {
			t.Errorf("Create URL mismatch, expected '%s', got '%s' for %v", r.url, actual, r.post)
		}
	}
}

type parseDateResult struct {
	in  string
	out time.Time
}

func TestParseDate(t *testing.T) {
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		t.Fatal(err)
	}
	results := []parseDateResult{
		{"13 September 2012", time.Date(2012, 9, 13, 0, 0, 0, 0, loc)},
		{"1 April 2012", time.Date(2012, 4, 1, 0, 0, 0, 0, loc)},
		{"October 12 2011", time.Date(2011, 10, 12, 0, 0, 0, 0, loc)},
		{"August 2 2011", time.Date(2011, 8, 2, 0, 0, 0, 0, loc)},
		{"March 2, 2012", time.Date(2012, 3, 2, 0, 0, 0, 0, loc)},
	}

	for _, r := range results {
		actual := parseDate(r.in)
		if actual.IsZero() {
			t.Errorf("Failed to parse input '%s'", r.in)
		} else if actual.Year() != r.out.Year() || actual.Month() != r.out.Month() || actual.Day() != r.out.Day() {
			t.Errorf("Date parse fail. Input '%s', expected '%v', got '%v'", r.in, r.out, actual)
		}
	}
}
