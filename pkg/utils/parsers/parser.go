package parsers

import (
	"github.com/gocolly/colly/v2"
	"hestia/pkg/models"
)

// ErrFunc is a function that handles errors.
type ErrFunc func(error)

type Collector struct {
	*colly.Collector
	errHandler ErrFunc
}

func NewCollector(collector *colly.Collector, errHandler ErrFunc) *Collector {
	return &Collector{
		Collector:  collector,
		errHandler: errHandler,
	}
}

func (c *Collector) Parse(url string) (models.Flat, error) {
	flat := models.Flat{}

	// Set up callbacks to handle scraping events
	c.OnHTML(".css-y6l269.er0e7w63", func(e *colly.HTMLElement) {
		// Extract data from HTML elements
		flat.Title = e.ChildText(".css-1wnihf5.efcnut38")
		flat.Price = e.ChildText(".css-t3wmkv.e1l1avn10")
		flat.Address = e.ChildText(".e1w8sadu0.css-1helwne.exgq9l20")

		flat.Surface = e.ChildText(".css-1wi2w6s.enb64yk5")
		flat.Rooms = e.ChildText(".css-19yhkv9.enb64yk08")
		flat.Floor = e.ChildText(".css-1wi2w6s.enb64yk5")
		flat.AvailableFrom = e.ChildText(".css-x0kl3j.e1k3ukdh0")
		flat.Rent = e.ChildText(".css-1wi2w6s.enb64yk5")
		flat.Deposit = e.ChildText(".css-1wi2w6s.enb64yk5")
		flat.Description = e.ChildText(".css-1ugtzj2.e175i4j93")

	})

	// Visit the URL and start scraping
	err := c.Visit(url)
	if err != nil {
		c.errHandler(err)
		return flat, err
	}

	return flat, nil
}
