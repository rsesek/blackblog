# Blackblog

Blackblog is a simple blogging platform written in Go, that uses the
[Blackfriday Markdown](https://github.com/russross/blackfriday) library to
format posts.

You can run Blackblog in two modes: standalone server and compiler. In server
mode, you specify a path to the blog posts and it dynamically serves pages on
the fly, acting as a HTTP web server. In compile mode, it takes a directory
of posts and outputs a new directory of static HTML files.

## Starting a Post

Posts use pure Markdown formatting, but have some additional metadata at the
beginning of the file. Metadata lines begin with two tilde `~~` characters and
can only occur at the top of the file. The first line that does not start with
two tildes ends the metadata section and starts the post in Markdown format.

Currently, the following metadata attributes are supported:

* **Title**: The name of the post, which is unique from the first heading.
* **URL**: The URL fragment for the blog post.
* **Date**: The date and time at which the post was published.

Example:

    ~~ Title: How To Use Blackblog
    ~~ Date: 24 January 2012
    ~~ URL: using-blackblog

    # Using Blackblog
    Blackblog lorem ipsum dolor sit amet.

The URL metadata will be used to construct a URL of the form:
`/YYYY/MM/url.html`.
