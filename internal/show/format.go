package show

import (
	"io"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/olekukonko/tablewriter"

	"github.com/Chyroc/phpmyadmin-cli/internal/common"
)

var out io.Writer = os.Stdout

func TestSetOut(w io.Writer) {
	out = w
}

func parseFromHTML(html string) ([]string, [][]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, nil, err
	}

	var header []string
	var datas [][]string
	var columnLine = -1
	var rowLine = -1

	doc.Find("table").Each(func(i int, selection *goquery.Selection) {
		if selection.HasClass("print_ignore") {
			selection.Remove()
		}
	})

	// header
	doc.Find("tr").Each(func(j int, tr *goquery.Selection) {
		if columnLine == -1 {
			if strings.Contains(tr.Find("td").Text(), "Edit Copy") {
				columnLine = 3
			}
		}
		if rowLine == -1 {
			if len(header) == 1 && (header[0] == "Database" || strings.HasPrefix(header[0], "Tables_in_")) {
				rowLine = 0
			}
		}

		tr.Find("th").Each(func(_ int, th *goquery.Selection) {
			if th.Find("span").HasClass("tblcomment") {
				th.Find("span").Remove()
			}

			thText := th.Text()
			if thText != "" {
				header = append(header, strings.TrimSpace(thText))
			}
		})
	})

	// datas
	doc.Find("tr").Each(func(j int, tr *goquery.Selection) {
		if j <= rowLine {
			return
		}

		var data []string
		tr.Find("td").Each(func(i int, td *goquery.Selection) {
			if i <= columnLine {
				return
			}
			data = append(data, td.Text())
		})

		if len(data) != 0 && (len(header) == 0 || (len(header) > 0 && len(header) == len(data))) {
			datas = append(datas, data)
		}
	})

	return header, datas, nil
}

// FromHTML Parse table from HTML
func ParseFromHTML(html string) error {
	header, datas, err := parseFromHTML(html)
	if err != nil {
		return err
	}

	for _, v := range header {
		common.Debug3("header [%s]\n", v)
	}

	for _, vv := range datas {
		for _, v := range vv {
			common.Debug3("datas [%s]\n", v)
		}

	}

	t := tablewriter.NewWriter(out)
	t.SetHeader(header)
	t.SetAutoFormatHeaders(false)
	t.SetAutoWrapText(false)
	for _, v := range datas {
		t.Append(v)
	}
	t.Render()

	return nil
}
