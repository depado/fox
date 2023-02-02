package player

import (
	"fmt"
	"sync"
	"time"

	"github.com/wcharczuk/go-chart/v2"
)

type Stats struct {
	sync.RWMutex

	PlaybackPosition time.Duration
	Duration         time.Duration
	Bitrate          float32
	Size             int
	Speed            float32
	TimeAxis         []float64
	BiteRateAxis     []float64
}

var whiteStyle = chart.Style{
	StrokeColor: chart.ColorWhite,
	FontColor:   chart.ColorWhite,
}

func MinMax(array []float64) (float64, float64) {
	var max float64 = array[0]
	var min float64 = array[0]
	for _, value := range array {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min, max
}

func (s *Stats) GenerateChart() *chart.Chart {
	s.RLock()
	defer s.RUnlock()

	if len(s.BiteRateAxis) < 2 {
		return nil
	}

	min, max := MinMax(s.BiteRateAxis)
	if min > 5 {
		min -= 5
	}
	max += 5

	cSeries := chart.ContinuousSeries{
		Style: chart.Style{
			FillColor: chart.GetDefaultColor(0).WithAlpha(50),
		},
		XValues: s.TimeAxis,
		YValues: s.BiteRateAxis,
	}

	annotation := chart.LastValueAnnotationSeries(cSeries)
	annotation.Style = chart.Style{
		FontColor:   chart.ColorWhite,
		FillColor:   chart.GetDefaultColor(0).WithAlpha(50),
		StrokeColor: chart.GetDefaultColor(0),
	}

	graph := chart.Chart{
		Title:      "Encoding Stats",
		TitleStyle: whiteStyle,
		Background: chart.Style{
			FillColor: chart.ColorTransparent,
		},
		XAxis: chart.XAxis{
			TickPosition: chart.TickPositionBetweenTicks,
			Style:        whiteStyle,
			ValueFormatter: func(v interface{}) string {
				return fmtDuration(time.Duration(v.(float64)))
			},
		},
		YAxis: chart.YAxis{
			AxisType: chart.YAxisSecondary,
			Style:    whiteStyle,
			Range: &chart.ContinuousRange{
				Min: min,
				Max: max,
			},
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("%.0f kB/s", v.(float64))
			},
		},
		Canvas: chart.Style{
			FillColor: chart.ColorTransparent,
		},
		Series: []chart.Series{
			cSeries,
			annotation,
		},
	}

	return &graph
}
