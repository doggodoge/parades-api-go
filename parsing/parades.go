package parsing

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ParadeSummary struct {
	Date          string `json:"date"`
	Parade        string `json:"parade"`
	Town          string `json:"town"`
	StartTime     string `json:"startTime"`
	Determination string `json:"endTime"`
	DetailsURL    string `json:"detailsUrl"`
}

type ParadeDetails struct {
	DateOfParade            string `json:"dateOfParade"`
	StartTimeOfOutwardRoute string `json:"startTimeOfOutwardRoute"`
	ProposedOutwardRoute    string `json:"proposedOutwardRoute"`
	EndTimeOfOutwardRoute   string `json:"endTimeOfOutwardRoute"`
	StartTimeOfReturnRoute  string `json:"startTimeOfReturnRoute"`
	ProposedReturnRoute     string `json:"proposedReturnRoute"`
	EndTimeOfReturnRoute    string `json:"endTimeOfReturnRoute"`
	NumberOfBands           string `json:"numberOfBands"`
	Bands                   string `json:"bands"`
	NumberOfParticipants    string `json:"numberOfParticipants"`
	NumberOfSupporters      string `json:"numberOfSupporters"`
}

func parseParadesHTML(html string) []*ParadeSummary {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}

	table := doc.Find("table.HomePageTable")
	rows := table.Find("tr")

	parades := make([]*ParadeSummary, 0)

	rows.Each(func(_ int, row *goquery.Selection) {
		cells := row.Find("td")

		if cells.Length() == 0 {
			return
		}

		parade := &ParadeSummary{
			Date:          cells.Eq(0).Text(),
			Parade:        cells.Eq(1).Text(),
			Town:          cells.Eq(2).Text(),
			StartTime:     cells.Eq(3).Text(),
			Determination: cells.Eq(4).Text(),
			DetailsURL:    "https://www.paradescommission.org" + cells.Eq(1).Find("a").AttrOr("href", ""),
		}

		parades = append(parades, parade)
	})

	return parades
}

func ParseParadesDetailsHTML(html string) ParadeDetails {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		log.Fatal(err)
	}

	table := doc.Find("table.HomePageTable")
	secondColumn := table.Find("td:nth-child(2)")

	parade := ParadeDetails{
		DateOfParade:            secondColumn.Eq(1).Text(),
		StartTimeOfOutwardRoute: secondColumn.Eq(2).Text(),
		ProposedOutwardRoute:    secondColumn.Eq(3).Text(),
		EndTimeOfOutwardRoute:   secondColumn.Eq(4).Text(),
		StartTimeOfReturnRoute:  secondColumn.Eq(5).Text(),
		ProposedReturnRoute:     secondColumn.Eq(6).Text(),
		EndTimeOfReturnRoute:    secondColumn.Eq(7).Text(),
		NumberOfBands:           secondColumn.Eq(8).Text(),
		Bands:                   secondColumn.Eq(9).Text(),
		NumberOfParticipants:    secondColumn.Eq(10).Text(),
		NumberOfSupporters:      secondColumn.Eq(11).Text(),
	}

	return parade
}

func AllParades() []*ParadeSummary {
	resp, err := http.Get("https://www.paradescommission.org/home.aspx")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return parseParadesHTML(string(body))
}
