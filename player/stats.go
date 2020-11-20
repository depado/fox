package player

import (
	"fmt"
	"sync"
	"time"

	"github.com/wcharczuk/go-chart"
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
	Show:        true,
	StrokeColor: chart.ColorWhite,
	FontColor:   chart.ColorWhite,
}

func (s *Stats) GenerateChart() *chart.Chart {
	s.RLock()
	defer s.RUnlock()

	if len(s.BiteRateAxis) < 2 {
		return nil
	}

	graph := chart.Chart{
		Title:      "Encoding Stats",
		TitleStyle: whiteStyle,
		Background: chart.Style{
			Show:      true,
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
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("%.0f kB/s", v.(float64))
			},
		},
		Canvas: chart.Style{
			Show:      true,
			FillColor: chart.ColorTransparent,
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Style: chart.Style{
					Show:      true,
					FillColor: chart.GetDefaultColor(0).WithAlpha(50),
				},
				XValues: s.TimeAxis,
				YValues: s.BiteRateAxis,
			},
		},
	}

	return &graph
}
