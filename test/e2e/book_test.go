package books_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type Book struct {
	Title  string
	Author string
	Pages  int
}

func (book Book) Category() category {
	if book.Pages > 300 {
		return CategoryNovel
	}
	return CategoryShortStory
}

type category string

const (
	CategoryNovel      category = "NOVEL"
	CategoryShortStory category = "SHORT STORY"
)

var _ = Describe("Books", func() {
	var foxInSocks, lesMis *Book

	BeforeEach(func() {
		lesMis = &Book{
			Title:  "Les Miserables",
			Author: "Victor Hugo",
			Pages:  2783,
		}

		foxInSocks = &Book{
			Title:  "Fox In Socks",
			Author: "Dr. Seuss",
			Pages:  24,
		}
	})

	Describe("Categorizing books", func() {
		Context("with more than 300 pages", func() {
			It("should be a novel", func() {
				Expect(lesMis.Category()).To(Equal(CategoryNovel))
			})
		})

		Context("with fewer than 300 pages", func() {
			It("should be a short story", func() {
				Expect(foxInSocks.Category()).To(Equal(CategoryShortStory))
			})
		})
	})
})
