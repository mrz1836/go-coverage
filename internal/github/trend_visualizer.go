// Package github provides ASCII trend visualization for coverage data
package github

import (
	"fmt"
	"math"
	"strings"

	"github.com/mrz1836/go-coverage/internal/history"
)

// TrendVisualizer generates ASCII charts for coverage trends
type TrendVisualizer struct {
	config *TrendConfig
}

// TrendConfig configures trend visualization
type TrendConfig struct {
	Width      int    // Chart width in characters
	Height     int    // Chart height in characters
	ShowValues bool   // Show values on data points
	ShowGrid   bool   // Show grid lines
	ShowLegend bool   // Show legend
	TimeFormat string // Time format for X-axis labels
	ChartStyle string // Chart style: "line", "bar", "sparkline"
}

// DefaultTrendConfig returns default configuration for trend visualization
func DefaultTrendConfig() *TrendConfig {
	return &TrendConfig{
		Width:      50,
		Height:     8,
		ShowValues: true,
		ShowGrid:   false,
		ShowLegend: true,
		TimeFormat: "Jan 2",
		ChartStyle: "line",
	}
}

// NewTrendVisualizer creates a new trend visualizer
func NewTrendVisualizer(config *TrendConfig) *TrendVisualizer {
	if config == nil {
		config = DefaultTrendConfig()
	}
	return &TrendVisualizer{config: config}
}

// GenerateASCIIChart creates an ASCII chart from coverage history records
func (tv *TrendVisualizer) GenerateASCIIChart(records []history.CoverageRecord) string {
	if len(records) == 0 {
		return "No coverage history data available"
	}

	switch tv.config.ChartStyle {
	case "sparkline":
		return tv.generateSparkline(records)
	case "bar":
		return tv.generateBarChart(records)
	default:
		return tv.generateLineChart(records)
	}
}

// generateLineChart creates a line chart visualization
func (tv *TrendVisualizer) generateLineChart(records []history.CoverageRecord) string {
	if len(records) < 2 {
		return tv.generateSinglePoint(records[0])
	}

	// Find min/max values for scaling
	minVal, maxVal := tv.findMinMax(records)
	if maxVal == minVal {
		maxVal = minVal + 1 // Avoid division by zero
	}

	var lines []string
	width := tv.config.Width
	height := tv.config.Height

	// Create the chart grid
	for y := height - 1; y >= 0; y-- {
		line := make([]rune, width)

		// Fill with spaces initially
		for x := range line {
			line[x] = ' '
		}

		// Calculate Y value for this row
		yVal := minVal + (maxVal-minVal)*float64(y)/float64(height-1)

		// Plot data points
		for i, record := range records {
			if len(records) == 1 {
				continue
			}

			// Calculate X position
			x := int(float64(i) * float64(width-1) / float64(len(records)-1))
			if x >= width {
				x = width - 1
			}

			// Calculate if this point should be plotted at this Y level
			pointY := int((record.Percentage - minVal) / (maxVal - minVal) * float64(height-1))

			if pointY == y {
				if tv.config.ShowValues && i == len(records)-1 {
					// Show value for the last point
					valueStr := fmt.Sprintf("%.1f%%", record.Percentage)
					if x+len(valueStr) <= width {
						for j, ch := range valueStr {
							if x+j < width {
								line[x+j] = ch
							}
						}
					} else {
						line[x] = '●'
					}
				} else {
					line[x] = '●'
				}
			}
		}

		// Add Y-axis label
		yLabel := fmt.Sprintf("%4.1f%%", yVal)
		lineStr := yLabel + " ┤" + string(line)
		lines = append(lines, lineStr)
	}

	// Add X-axis
	xAxis := strings.Repeat(" ", 6) + "└" + strings.Repeat("─", width) + "─"
	lines = append(lines, xAxis)

	// Add time labels
	if len(records) >= 2 {
		timeLabels := strings.Repeat(" ", 7)
		firstTime := records[0].Timestamp.Format(tv.config.TimeFormat)
		lastTime := records[len(records)-1].Timestamp.Format(tv.config.TimeFormat)

		timeLabels += firstTime
		padding := width - len(firstTime) - len(lastTime) + 1
		if padding > 0 {
			timeLabels += strings.Repeat(" ", padding) + lastTime
		}
		lines = append(lines, timeLabels)
	}

	return strings.Join(lines, "\n")
}

// generateSparkline creates a compact sparkline visualization
func (tv *TrendVisualizer) generateSparkline(records []history.CoverageRecord) string {
	if len(records) == 0 {
		return ""
	}

	sparkChars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	// Find min/max for scaling
	minVal, maxVal := tv.findMinMax(records)
	if maxVal == minVal {
		return string(sparkChars[4]) + fmt.Sprintf(" %.1f%%", records[0].Percentage)
	}

	var sparkline strings.Builder
	for _, record := range records {
		// Normalize value to 0-7 range for spark characters
		normalized := (record.Percentage - minVal) / (maxVal - minVal)
		index := int(normalized * 7)
		if index > 7 {
			index = 7
		}
		if index < 0 {
			index = 0
		}
		sparkline.WriteRune(sparkChars[index])
	}

	// Add current value
	current := records[len(records)-1].Percentage
	return sparkline.String() + fmt.Sprintf(" %.1f%%", current)
}

// generateBarChart creates a horizontal bar chart
func (tv *TrendVisualizer) generateBarChart(records []history.CoverageRecord) string {
	if len(records) == 0 {
		return "No data"
	}

	var lines []string
	maxLabelWidth := 0

	// Find max label width for alignment
	for _, record := range records {
		label := record.Timestamp.Format(tv.config.TimeFormat)
		if len(label) > maxLabelWidth {
			maxLabelWidth = len(label)
		}
	}

	// Find min/max for scaling
	minVal, maxVal := tv.findMinMax(records)
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	// Show only the last few records to avoid overwhelming the comment
	startIdx := 0
	if len(records) > 10 {
		startIdx = len(records) - 10
	}

	for i := startIdx; i < len(records); i++ {
		record := records[i]
		label := fmt.Sprintf("%-*s", maxLabelWidth, record.Timestamp.Format(tv.config.TimeFormat))

		// Calculate bar length
		barLength := int((record.Percentage - minVal) / (maxVal - minVal) * float64(tv.config.Width-maxLabelWidth-10))
		if barLength < 0 {
			barLength = 0
		}

		bar := strings.Repeat("█", barLength)
		value := fmt.Sprintf("%.1f%%", record.Percentage)

		line := fmt.Sprintf("%s │%s %s", label, bar, value)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// generateSinglePoint creates a simple representation for a single data point
func (tv *TrendVisualizer) generateSinglePoint(record history.CoverageRecord) string {
	return fmt.Sprintf("● %.1f%% (%s)", record.Percentage, record.Timestamp.Format(tv.config.TimeFormat))
}

// findMinMax finds the minimum and maximum coverage values in the records
func (tv *TrendVisualizer) findMinMax(records []history.CoverageRecord) (float64, float64) {
	if len(records) == 0 {
		return 0, 100
	}

	minVal := records[0].Percentage
	maxVal := records[0].Percentage

	for _, record := range records {
		if record.Percentage < minVal {
			minVal = record.Percentage
		}
		if record.Percentage > maxVal {
			maxVal = record.Percentage
		}
	}

	// Add some padding to make the chart more readable
	padding := (maxVal - minVal) * 0.1
	if padding < 1.0 {
		padding = 1.0
	}

	minVal = math.Max(0, minVal-padding)
	maxVal = math.Min(100, maxVal+padding)

	return minVal, maxVal
}

// GenerateTrendSummary creates a text summary of the trend
func (tv *TrendVisualizer) GenerateTrendSummary(records []history.CoverageRecord) string {
	if len(records) < 2 {
		if len(records) == 1 {
			return fmt.Sprintf("Current coverage: %.1f%%", records[0].Percentage)
		}
		return "No trend data available"
	}

	current := records[len(records)-1].Percentage
	previous := records[len(records)-2].Percentage
	diff := current - previous

	var trendEmoji string
	var trendText string

	if diff > 0.5 {
		trendEmoji = "📈"
		trendText = "improving"
	} else if diff < -0.5 {
		trendEmoji = "📉"
		trendText = "declining"
	} else {
		trendEmoji = "📊"
		trendText = "stable"
	}

	// Calculate overall trend over all records
	if len(records) >= 3 {
		oldest := records[0].Percentage
		overallDiff := current - oldest
		days := int(records[len(records)-1].Timestamp.Sub(records[0].Timestamp).Hours() / 24)

		if days > 0 {
			return fmt.Sprintf("%s Coverage %s (%.1f%% → %.1f%%, %+.1f%% over %d days)",
				trendEmoji, trendText, oldest, current, overallDiff, days)
		}
	}

	return fmt.Sprintf("%s Coverage %s (%.1f%% → %.1f%%, %+.1f%%)",
		trendEmoji, trendText, previous, current, diff)
}

// GenerateCompactTrend creates a compact one-line trend representation
func (tv *TrendVisualizer) GenerateCompactTrend(records []history.CoverageRecord) string {
	if len(records) == 0 {
		return "No data"
	}

	sparkline := tv.generateSparkline(records)
	summary := ""

	if len(records) >= 2 {
		current := records[len(records)-1].Percentage
		previous := records[0].Percentage
		diff := current - previous

		if diff > 0.1 {
			summary = fmt.Sprintf(" (+%.1f%%)", diff)
		} else if diff < -0.1 {
			summary = fmt.Sprintf(" (%.1f%%)", diff)
		}
	}

	return sparkline + summary
}
