// Copyright 2019 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hugolib

import (
	"context"
	"html/template"

	"github.com/gohugoio/hugo/resources/page"
)

// A placeholder for the TableOfContents markup. This is what we pass to the Goldmark etc. renderers.
var tocShortcodePlaceholder = createShortcodePlaceholder("TOC", 0)

// shortcodeRenderer is typically used to delay rendering of inner shortcodes
// marked with placeholders in the content.
type shortcodeRenderer interface {
	renderShortcode(context.Context) ([]byte, bool, error)
	renderShortcodeString(context.Context) (string, bool, error)
}

type shortcodeRenderFunc func(context.Context) ([]byte, bool, error)

func (f shortcodeRenderFunc) renderShortcode(ctx context.Context) ([]byte, bool, error) {
	return f(ctx)
}

func (f shortcodeRenderFunc) renderShortcodeString(ctx context.Context) (string, bool, error) {
	b, has, err := f(ctx)
	return string(b), has, err
}

type prerenderedShortcode struct {
	s           string
	hasVariants bool
}

func (p prerenderedShortcode) renderShortcode(context.Context) ([]byte, bool, error) {
	return []byte(p.s), p.hasVariants, nil
}

func (p prerenderedShortcode) renderShortcodeString(context.Context) (string, bool, error) {
	return p.s, p.hasVariants, nil
}

var zeroShortcode = prerenderedShortcode{}

// This is sent to the shortcodes. They cannot access the content
// they're a part of. It would cause an infinite regress.
//
// Go doesn't support virtual methods, so this careful dance is currently (I think)
// the best we can do.
type pageForShortcode struct {
	page.PageWithoutContent
	page.ContentProvider

	// We need to replace it after we have rendered it, so provide a
	// temporary placeholder.
	toc template.HTML

	p *pageState
}

func newPageForShortcode(p *pageState) page.Page {
	return &pageForShortcode{
		PageWithoutContent: p,
		ContentProvider:    page.NopPage,
		toc:                template.HTML(tocShortcodePlaceholder),
		p:                  p,
	}
}

func (p *pageForShortcode) page() page.Page {
	return p.PageWithoutContent.(page.Page)
}

func (p *pageForShortcode) String() string {
	return p.p.String()
}

func (p *pageForShortcode) TableOfContents(context.Context) template.HTML {
	p.p.enablePlaceholders()
	return p.toc
}

// This is what is sent into the content render hooks (link, image).
type pageForRenderHooks struct {
	page.PageWithoutContent
	page.TableOfContentsProvider
	page.ContentProvider
}

func newPageForRenderHook(p *pageState) page.Page {
	return &pageForRenderHooks{
		PageWithoutContent:      p,
		ContentProvider:         page.NopPage,
		TableOfContentsProvider: page.NopPage,
	}
}

func (p *pageForRenderHooks) page() page.Page {
	return p.PageWithoutContent.(page.Page)
}
