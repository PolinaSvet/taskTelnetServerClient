package proverbs

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	goquery "github.com/PuerkitoBio/goquery"
)

var (
	proverbsList  []string
	localProverbs = []string{
		"Local: Don't communicate by sharing memory, share memory by communicating.",
		"Local: Concurrency is not parallelism.",
		"Local: Channels orchestrate; mutexes serialize.",
		"Local: The bigger the interface, the weaker the abstraction.",
		"Local: Make the zero value useful.",
		"Local: interface{} says nothing.",
		"Local: Gofmt's style is no one's favorite, yet gofmt is everyone's favorite.",
		"Local: A little copying is better than a little dependency.",
		"Local: Syscall must always be guarded with build tags.",
		"Local: Cgo must always be guarded with build tags.",
		"Local: Cgo is not Go.",
		"Local: With the unsafe package there are no guarantees.",
		"Local: Clear is better than clever.",
		"Local: Reflection is never clear.",
		"Local: Errors are values.",
		"Local: Don't just check errors, handle them gracefully.",
		"Local: Design the architecture, name the components, document the details.",
		"Local: Documentation is for users.",
		"Local: Don't panic.",
	}
	loadProverbsURL = "https://go-proverbs.github.io/"
)

// Инициализация: загрузка поговорок с сайта
func init() {
	rand.Seed(time.Now().UnixNano())
	if err := loadProverbs(loadProverbsURL); err != nil {
		proverbsList = localProverbs
	}
}

// Загрузка поговорок с сайта
func loadProverbs(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch data: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Используем goquery для парсинга HTML
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return err
	}

	doc.Find("h3").Each(func(index int, item *goquery.Selection) {
		proverbsList = append(proverbsList, item.Text())
	})

	return nil
}

// Получение случайной поговорки
func GetRandomProverb() string {
	if len(proverbsList) == 0 {
		return "No proverbs available."
	}
	return proverbsList[rand.Intn(len(proverbsList))]
}
