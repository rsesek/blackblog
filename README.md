# Blackblog

Blackblog is a simple blogging platform written in Go. It uses the
[Blackfriday Markdown](https://github.com/russross/blackfriday) library to
format posts.

You can run Blackblog in two modes: standalone server and compiler. In server
mode, your blog will be served dynamically using a built-in web server. This is
useful when writing because your words will automatically be updated when you
reload blog pages. In compile mode, your blog is rendered out to a directory of
static HTML files, which can then be hosted by an web server.

## Beta Software

Blackblog is beta software. It is used to publish my personal blog and is
feature complete, but it is not yet well-tested beyond that and its unit tests.

## Installation

To use Blackblog, you must first have [Go](http://golang.org) installed on your
machine. Then, install both the Blackfriday Markdown library and the Blackblog
program:

    $ go get github.com/russross/blackfriday
    $ go get github.com/rsesek/blackblog

To get started, create a new blog in a directory `myblog` and run the built-in
server:

    $ blackblog newblog myblog
    $ blackblog serve myblog

Point your web browser to the location it prints. Then try editing
`myblog/posts/welcome.md` and see your updates immediately as you refresh the
page in your browser.

To add new posts, simply create a `file.md` in `myblog/posts/`.

    $ vim myblog/posts/first_post.md

You can customize the title and other parameters by editing the configuration
file. Note that this requires a server restart:

    $ vim myblog/blackblog.json

You can also render your blog as static files, to the directory specified in the
configuration file (mentioned above). To do so:

    $ blackblog render myblog

And then just publish it on the Internet by uploading it to your website:

    $ scp -r ./myblog/out/ example.com:~/public_html/blog

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

## Customizing the Appearance

Blackblog comes with a very basic style that you will most likely wish to
customize for your own blog. Start out by looking in `myblog/blackblog.json` and
copying the directory specified in `TemplatesDir` to your `myblog` directory.

Then, update the configuration file by changing `TemplatesDir` and
`StaticFilesDir`. Since you've copied these to `myblog`, the values should be
`./templates/` and `./templates/static/`, respecitvely.

From there, you can edit the HTML template files and the CSS file in your blog.
Try running Blackblog in server mode when editing templates, which will allow
you to just reload the pages to see the changes you're making.
