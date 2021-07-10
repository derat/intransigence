// Copyright 2021 Daniel Erat <dan@erat.org>.
// All rights reserved.

package build

import (
	"os"
	"time"

	"github.com/derat/intransigence/render"

	"github.com/gorilla/feeds"
)

const maxFeedItems = 10

// writeFeed writes an Atom feed listing pages from feedInfos to path p.
// feedInfos should be sorted newest-to-oldest.
func writeFeed(p string, si *render.SiteInfo, feedInfos []render.PageFeedInfo) error {
	feed := &feeds.Feed{
		Title:       si.FeedTitle,
		Link:        &feeds.Link{Href: si.BaseURL},
		Description: si.FeedDesc,
		Author:      &feeds.Author{Name: si.AuthorName, Email: si.AuthorEmail},
	}

	for i := 0; i < len(feedInfos) && i < maxFeedItems; i++ {
		fi := &feedInfos[i]
		feed.Add(&feeds.Item{
			Title:       fi.Title,
			Link:        &feeds.Link{Href: fi.AbsURL},
			Description: fi.Desc,
			Author:      &feeds.Author{Name: si.AuthorName, Email: si.AuthorEmail},
			Created:     fi.Created.UTC(),
		})
	}

	// RFC 4287 says that the atom:updated field should indicate "the most recent instant in
	// time when an entry or feed was modified in a way the publisher considers significant",
	// but I don't have any good way of diffing against the currently-published feed to see if
	// anything has changed.
	if len(feed.Items) > 0 {
		feed.Created = feed.Items[0].Created
	} else {
		feed.Created = time.Now().UTC()
	}

	atom, err := feed.ToAtom()
	if err != nil {
		return err
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(atom); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
