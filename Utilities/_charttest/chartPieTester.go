package main

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"math/rand"
	"os"
)

func main() {
	createPieChart()
}

// generate random data for pie chart
func generatePieItems() []opts.PieData {
	subjects := []string{"Maths", "English", "Science", "Computers", "History", "Geography"}
	items := make([]opts.PieData, 0)
	for i := 0; i < 6; i++ {
		items = append(items, opts.PieData{
			Name:  subjects[i],
			Value: rand.Intn(500)})
	}
	return items
}

func createPieChart() {
	// create a new pie instance
	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(
			opts.Title{
				Title:    "Pie chart in Go",
				Subtitle: "This is fun to use!",
			},
		),
	)
	pie.SetSeriesOptions()
	pie.AddSeries("Monthly revenue",
		generatePieItems()).
		SetSeriesOptions(
			charts.WithPieChartOpts(
				opts.PieChart{
					Radius: 200,
				},
			),
			charts.WithLabelOpts(
				opts.Label{
					Show:      true,
					Formatter: "{b}: {c}",
				},
			),
		)
	f, _ := os.Create("pie.html")
	_ = pie.Render(f)
}
