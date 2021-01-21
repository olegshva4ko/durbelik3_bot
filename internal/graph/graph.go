package graph

import (
	"Durbelik3/pkg/models"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

//CreateImage ...
func CreateImage(afkData []*models.Violation, days int, chatID int64) {
	values, ticks := processValues(afkData, days)
	graph := chart.BarChart{
		Title:    "AFK statistics",
		Width:    840,
		Height:   512,
		BarWidth: 50,
		Background: chart.Style{
			Padding: chart.Box{
				Top:    40,
				Bottom: 20,
				Right:  40,
				Left:   40,
				IsSet:  true,
			},
		},
		Canvas: chart.Style{
			Hidden:      false,
			StrokeWidth: 1,
			StrokeColor: drawing.Color{14, 54, 74, 100},
		},
		YAxis: chart.YAxis{
			Zero: chart.GridLine{
				IsMinor: true,
				Value:   0,
			},
			Style: chart.Style{
				Hidden:      false,
				StrokeWidth: 1,
				StrokeColor: drawing.Color{14, 54, 74, 100},
			},
			TickStyle: chart.Style{
				Hidden:      false,
				StrokeWidth: 1,
				StrokeColor: drawing.Color{14, 54, 74, 100},
			},
			Ticks: ticks,
		},
		BarSpacing:   30,
		UseBaseValue: true,
		BaseValue:    0,

		Bars: values,
	}
	f, _ := os.Create(fmt.Sprintf("./internal/graph/%d_graph.png", chatID))
	defer f.Close()
	graph.Render(chart.PNG, f)
}

func processValues(data []*models.Violation, days int) ([]chart.Value, []chart.Tick) {
	allStats := make(map[string]int)

	for _, afk := range data { //getting Userid, Chatid, Date, Reason, Firstname
		date := time.Unix(afk.Date, 0).Format("2006.01.02")
		allStats[date]++
	}

	keys := make([]string, 0, len(allStats))
	for k := range allStats {
		keys = append(keys, k) //getting keys from map (01.02)
	}
	sort.Strings(keys) //sorting keys

	var ChartValues []chart.Value
	for _, date := range keys {
		label := strings.Split(date, ".")

		ChartValues = append(ChartValues, chart.Value{Value: float64(allStats[date]),
			Label: strings.Join(label[1:], ".")})

	}

	if days <= len(ChartValues) {
		ChartValues = ChartValues[len(ChartValues)-days:]
	}

	maxValue := 0 //get Max Value to generate ticks
	for _, v := range ChartValues {
		if maxValue <= int(v.Value) {
			maxValue = int(v.Value)
		}
	}
	var ticks []chart.Tick
	for i := 0; i <= maxValue; i++ {
		ticks = append(ticks, chart.Tick{Value: float64(i), Label: strconv.Itoa(i)})
	}

	return ChartValues, ticks
}
