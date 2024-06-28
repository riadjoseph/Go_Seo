package main

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"os"
)

func main() {
	createWordCloud()
}

var wordCloudData = map[string]interface{}{
	"Bitcoin":      10000,
	"Ethereum":     8000,
	"Cardano":      5000,
	"Polygon":      4000,
	"Polkadot":     3000,
	"Chainlink":    2500,
	"Solana":       2000,
	"Ripple":       1500,
	"Decentraland": 1000,
	"Tron":         800,
	"Sandbox":      500,
	"Litecoin":     200,
}

// generate random data for word cloud
func generateWordCloudData(data map[string]interface{}) (items []opts.WordCloudData) {
	items = make([]opts.WordCloudData, 0)
	for k, v := range data {
		items = append(items, opts.WordCloudData{Name: k, Value: v})
	}
	return
}

func createWordCloud() {
	wc := charts.NewWordCloud()
	wc.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    "Popular Cryptocurrencies",
			Subtitle: "Spot your favourite coins",
		}))
	wc.AddSeries("wordcloud", generateWordCloudData(wordCloudData)).
		SetSeriesOptions(
			charts.WithWorldCloudChartOpts(
				opts.WordCloudChart{
					SizeRange: []float32{40, 80},
					Shape:     "cardioid",
				}),
		)
	f, _ := os.Create("word_cloud.html")
	_ = wc.Render(f)
}
