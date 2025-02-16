// Copyright 2025 Blue Static <https://www.bluestatic.org>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	kindNewline = ast.NewNodeKind("newline")
	gmExts      = &goldmarkExts{}
)

func goldmarkParsers() parser.Option {
	return parser.WithASTTransformers(
		util.PrioritizedValue{
			Value:    gmExts,
			Priority: 9999,
		},
	)
}

func goldmarkRenderers() renderer.Option {
	return renderer.WithNodeRenderers(
		util.PrioritizedValue{
			Value:    gmExts,
			Priority: 1000,
		},
	)
}

// goldmarkExts inserts a newline in the HTML output between block elements.
// This ensures the HTML output is almost identical to that of Blackfriday.
type goldmarkExts struct{}

// parser.ASTTransformers:

func (*goldmarkExts) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n.Type() == ast.TypeInline {
			return ast.WalkSkipChildren, nil
		}
		p := n.Parent()
		if !entering || p == nil {
			return ast.WalkContinue, nil
		}
		k := n.Kind()
		isRightKind := k == ast.KindParagraph || k == ast.KindHeading || k == ast.KindCodeBlock ||
			k == ast.KindHTMLBlock || k == ast.KindBlockquote || k == ast.KindList
		if isRightKind && n.NextSibling() != nil {
			p.InsertAfter(p, n, &newline{})
		}
		return ast.WalkContinue, nil
	})
}

// renderer.NodeRenderer:

func (*goldmarkExts) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(kindNewline, renderNewline)
}

func renderNewline(w util.BufWriter, src []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		w.WriteString("\n")
	}
	return ast.WalkContinue, nil
}

// ast.Node:

type newline struct {
	ast.BaseInline
}

func (*newline) Kind() ast.NodeKind {
	return kindNewline
}

func (n *newline) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}
