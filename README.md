# Blackblog

Blackblog is a simple blogging platform written in Go. It uses the
[Blackfriday Markdown](https://github.com/russross/blackfriday) library to
format posts.

You can run Blackblog in two modes: standalone server and compiler. In server
mode, you specify a path to the blog posts and it dynamically serves pages on
the fly, acting as a HTTP web server. In compile mode, it takes a directory
of posts and outputs a new directory of static HTML files.

Blackblog is built for Go 1.

## Alpha Software

Blackblog is alpha software. It is used to publish my personal blog, but it is
not yet feature complete.

## Installation

    $ go get github.com/russross/blackfriday
    $ go get github.com/rsesek/blackblog

    To create a new blog and run the built-in server:
    $ blackblog newblog myblog
    $ blackblog serve myblog

    To add new posts, simply create a file.md in myblog/posts/.
    $ vim myblog/posts/first_post.md

    If you keep the server running, it will automatically reload the page as you
    write new content and refresh your web browser.

    Edit the title and other parameters (requires a server restart):
    $ vim myblog/blackblog.json

    To publish your blog as static files:
    $ blackblog render myblog
    $ scp -r ./myblog/out example.com:~/public_html/blog

This installs the two Go packages you need, starts a new blog and copies the
templates so they can be customized for your blog. It then compiles the blog
and uploads it to your server.

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
