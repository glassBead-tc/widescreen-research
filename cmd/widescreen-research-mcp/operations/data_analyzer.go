package operations

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spawn-mcp/coordinator/cmd/widescreen-research-mcp/schemas"
)

// DataAnalyzer performs analysis on research findings
type DataAnalyzer struct{}

// NewDataAnalyzer creates a new data analyzer
func NewDataAnalyzer() *DataAnalyzer {
	return &DataAnalyzer{}
}

// Execute analyzes research data
func (da *DataAnalyzer) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Extract drone results
	var droneResults []schemas.DroneResult
	
	if data, ok := params["data"].([]interface{}); ok {
		for _, d := range data {
			if result, ok := d.(schemas.DroneResult); ok {
				droneResults = append(droneResults, result)
			}
		}
	}

	if len(droneResults) == 0 {
		return nil, fmt.Errorf("no data provided for analysis")
	}

	// Get analysis type
	analysisType := "comprehensive"
	if at, ok := params["analysis_type"].(string); ok {
		analysisType = at
	}

	// Additional parameters
	additionalParams := make(map[string]interface{})
	if ap, ok := params["parameters"].(map[string]interface{}); ok {
		additionalParams = ap
	}

	// Perform analysis based on type
	switch analysisType {
	case "comprehensive":
		return da.comprehensiveAnalysis(ctx, droneResults, additionalParams)
	case "statistical":
		return da.statisticalAnalysis(ctx, droneResults, additionalParams)
	case "pattern":
		return da.patternAnalysis(ctx, droneResults, additionalParams)
	case "summary":
		return da.summaryAnalysis(ctx, droneResults, additionalParams)
	default:
		return da.comprehensiveAnalysis(ctx, droneResults, additionalParams)
	}
}

// comprehensiveAnalysis performs comprehensive data analysis
func (da *DataAnalyzer) comprehensiveAnalysis(ctx context.Context, results []schemas.DroneResult, params map[string]interface{}) (*schemas.DataAnalysisResponse, error) {
	// Initialize response
	response := &schemas.DataAnalysisResponse{
		Summary:        da.generateSummary(results),
		Insights:       da.extractInsights(results),
		Patterns:       da.identifyPatterns(results),
		Statistics:     da.calculateStatistics(results),
		Visualizations: da.generateVisualizations(results),
	}

	return response, nil
}

// statisticalAnalysis performs statistical analysis
func (da *DataAnalyzer) statisticalAnalysis(ctx context.Context, results []schemas.DroneResult, params map[string]interface{}) (*schemas.DataAnalysisResponse, error) {
	stats := da.calculateDetailedStatistics(results)
	
	return &schemas.DataAnalysisResponse{
		Summary:    "Statistical analysis of research data",
		Statistics: stats,
		Insights: []string{
			fmt.Sprintf("Total data points analyzed: %d", len(results)),
			fmt.Sprintf("Success rate: %.2f%%", stats["success_rate"].(float64)*100),
			fmt.Sprintf("Average processing time: %.2f seconds", stats["avg_processing_time"].(float64)),
		},
	}, nil
}

// patternAnalysis performs pattern analysis
func (da *DataAnalyzer) patternAnalysis(ctx context.Context, results []schemas.DroneResult, params map[string]interface{}) (*schemas.DataAnalysisResponse, error) {
	patterns := da.identifyDetailedPatterns(results)
	
	return &schemas.DataAnalysisResponse{
		Summary:  "Pattern analysis of research data",
		Patterns: patterns,
		Insights: da.generatePatternInsights(patterns),
	}, nil
}

// summaryAnalysis performs summary analysis
func (da *DataAnalyzer) summaryAnalysis(ctx context.Context, results []schemas.DroneResult, params map[string]interface{}) (*schemas.DataAnalysisResponse, error) {
	return &schemas.DataAnalysisResponse{
		Summary:  da.generateDetailedSummary(results),
		Insights: da.extractTopInsights(results, 5),
	}, nil
}

// Helper methods

func (da *DataAnalyzer) generateSummary(results []schemas.DroneResult) string {
	successCount := 0
	totalDataPoints := 0
	
	for _, result := range results {
		if result.Status == "completed" {
			successCount++
			totalDataPoints += len(result.Data)
		}
	}
	
	return fmt.Sprintf("Analysis of %d research results: %d successful completions with %d total data points collected",
		len(results), successCount, totalDataPoints)
}

func (da *DataAnalyzer) extractInsights(results []schemas.DroneResult) []string {
	insights := []string{}
	
	// Analyze completion rates
	completionRate := da.calculateCompletionRate(results)
	insights = append(insights, fmt.Sprintf("Research completion rate: %.2f%%", completionRate*100))
	
	// Analyze data quality
	dataQuality := da.assessDataQuality(results)
	insights = append(insights, fmt.Sprintf("Data quality score: %.2f/10", dataQuality))
	
	// Identify top sources
	topSources := da.identifyTopSources(results)
	if len(topSources) > 0 {
		insights = append(insights, fmt.Sprintf("Top data sources: %s", strings.Join(topSources[:3], ", ")))
	}
	
	// Analyze processing times
	avgTime, minTime, maxTime := da.analyzeProcessingTimes(results)
	insights = append(insights, fmt.Sprintf("Processing times - Avg: %.2fs, Min: %.2fs, Max: %.2fs", 
		avgTime.Seconds(), minTime.Seconds(), maxTime.Seconds()))
	
	return insights
}

func (da *DataAnalyzer) identifyPatterns(results []schemas.DroneResult) []schemas.Pattern {
	patterns := []schemas.Pattern{}
	
	// Pattern: Successful completion clustering
	if pattern := da.identifyCompletionPattern(results); pattern != nil {
		patterns = append(patterns, *pattern)
	}
	
	// Pattern: Data volume distribution
	if pattern := da.identifyDataVolumePattern(results); pattern != nil {
		patterns = append(patterns, *pattern)
	}
	
	// Pattern: Error patterns
	if pattern := da.identifyErrorPattern(results); pattern != nil {
		patterns = append(patterns, *pattern)
	}
	
	// Pattern: Source diversity
	if pattern := da.identifySourceDiversityPattern(results); pattern != nil {
		patterns = append(patterns, *pattern)
	}
	
	return patterns
}

func (da *DataAnalyzer) calculateStatistics(results []schemas.DroneResult) map[string]interface{} {
	stats := make(map[string]interface{})
	
	// Basic counts
	stats["total_results"] = len(results)
	stats["successful_results"] = da.countSuccessful(results)
	stats["failed_results"] = len(results) - stats["successful_results"].(int)
	
	// Success rate
	if len(results) > 0 {
		stats["success_rate"] = float64(stats["successful_results"].(int)) / float64(len(results))
	} else {
		stats["success_rate"] = 0.0
	}
	
	// Data points
	totalDataPoints := 0
	dataPointsPerDrone := make([]int, 0)
	
	for _, result := range results {
		if result.Status == "completed" {
			points := len(result.Data)
			totalDataPoints += points
			dataPointsPerDrone = append(dataPointsPerDrone, points)
		}
	}
	
	stats["total_data_points"] = totalDataPoints
	stats["avg_data_points_per_drone"] = 0.0
	if len(dataPointsPerDrone) > 0 {
		stats["avg_data_points_per_drone"] = float64(totalDataPoints) / float64(len(dataPointsPerDrone))
	}
	
	// Processing times
	avgTime, _, _ := da.analyzeProcessingTimes(results)
	stats["avg_processing_time"] = avgTime.Seconds()
	
	return stats
}

func (da *DataAnalyzer) generateVisualizations(results []schemas.DroneResult) []schemas.Visualization {
	visualizations := []schemas.Visualization{
		{
			Type:  "bar_chart",
			Title: "Research Completion Status",
			Data: map[string]interface{}{
				"labels": []string{"Completed", "Failed"},
				"values": []int{da.countSuccessful(results), len(results) - da.countSuccessful(results)},
			},
		},
		{
			Type:  "time_series",
			Title: "Research Progress Over Time",
			Data:  da.generateTimeSeriesData(results),
		},
	}
	
	return visualizations
}

// Utility methods

func (da *DataAnalyzer) calculateCompletionRate(results []schemas.DroneResult) float64 {
	if len(results) == 0 {
		return 0.0
	}
	return float64(da.countSuccessful(results)) / float64(len(results))
}

func (da *DataAnalyzer) countSuccessful(results []schemas.DroneResult) int {
	count := 0
	for _, result := range results {
		if result.Status == "completed" {
			count++
		}
	}
	return count
}

func (da *DataAnalyzer) assessDataQuality(results []schemas.DroneResult) float64 {
	// Simple quality assessment based on completeness and data volume
	totalScore := 0.0
	validResults := 0
	
	for _, result := range results {
		if result.Status == "completed" && len(result.Data) > 0 {
			score := 10.0
			
			// Deduct points for missing data
			if len(result.Data) < 5 {
				score -= 2.0
			}
			
			// Deduct points for errors
			if result.Error != "" {
				score -= 3.0
			}
			
			totalScore += score
			validResults++
		}
	}
	
	if validResults == 0 {
		return 0.0
	}
	
	return totalScore / float64(validResults)
}

func (da *DataAnalyzer) identifyTopSources(results []schemas.DroneResult) []string {
	sourceCount := make(map[string]int)
	
	for _, result := range results {
		if sources, ok := result.Data["sources"].([]interface{}); ok {
			for _, source := range sources {
				if s, ok := source.(string); ok {
					sourceCount[s]++
				}
			}
		}
	}
	
	// Sort sources by count
	type sourceFreq struct {
		source string
		count  int
	}
	
	var sources []sourceFreq
	for source, count := range sourceCount {
		sources = append(sources, sourceFreq{source, count})
	}
	
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].count > sources[j].count
	})
	
	topSources := []string{}
	for i, sf := range sources {
		if i >= 5 {
			break
		}
		topSources = append(topSources, sf.source)
	}
	
	return topSources
}

func (da *DataAnalyzer) analyzeProcessingTimes(results []schemas.DroneResult) (avg, min, max time.Duration) {
	if len(results) == 0 {
		return
	}
	
	var times []time.Duration
	for _, result := range results {
		if result.ProcessingTime > 0 {
			times = append(times, result.ProcessingTime)
		}
	}
	
	if len(times) == 0 {
		return
	}
	
	// Calculate min and max
	min = times[0]
	max = times[0]
	total := time.Duration(0)
	
	for _, t := range times {
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
		total += t
	}
	
	avg = total / time.Duration(len(times))
	return avg, min, max
}

// Pattern identification methods

func (da *DataAnalyzer) identifyCompletionPattern(results []schemas.DroneResult) *schemas.Pattern {
	successRate := da.calculateCompletionRate(results)
	
	if successRate > 0.9 {
		return &schemas.Pattern{
			Name:        "High Success Rate",
			Description: "Research drones achieved exceptional completion rate",
			Frequency:   da.countSuccessful(results),
			Confidence:  successRate,
		}
	} else if successRate < 0.5 {
		return &schemas.Pattern{
			Name:        "Low Success Rate",
			Description: "Research drones experienced significant failure rate",
			Frequency:   len(results) - da.countSuccessful(results),
			Confidence:  1.0 - successRate,
		}
	}
	
	return nil
}

func (da *DataAnalyzer) identifyDataVolumePattern(results []schemas.DroneResult) *schemas.Pattern {
	var volumes []int
	for _, result := range results {
		if result.Status == "completed" {
			volumes = append(volumes, len(result.Data))
		}
	}
	
	if len(volumes) == 0 {
		return nil
	}
	
	// Calculate variance
	avg := 0
	for _, v := range volumes {
		avg += v
	}
	avg /= len(volumes)
	
	variance := 0.0
	for _, v := range volumes {
		diff := float64(v - avg)
		variance += diff * diff
	}
	variance /= float64(len(volumes))
	
	if variance < float64(avg)*0.1 {
		return &schemas.Pattern{
			Name:        "Consistent Data Volume",
			Description: "Research drones collected similar amounts of data",
			Frequency:   len(volumes),
			Confidence:  0.85,
		}
	}
	
	return nil
}

func (da *DataAnalyzer) identifyErrorPattern(results []schemas.DroneResult) *schemas.Pattern {
	errorTypes := make(map[string]int)
	
	for _, result := range results {
		if result.Error != "" {
			// Simple error categorization
			if strings.Contains(strings.ToLower(result.Error), "timeout") {
				errorTypes["timeout"]++
			} else if strings.Contains(strings.ToLower(result.Error), "connection") {
				errorTypes["connection"]++
			} else {
				errorTypes["other"]++
			}
		}
	}
	
	// Find most common error
	maxCount := 0
	maxType := ""
	for errType, count := range errorTypes {
		if count > maxCount {
			maxCount = count
			maxType = errType
		}
	}
	
	if maxCount > len(results)/10 { // More than 10% errors of same type
		return &schemas.Pattern{
			Name:        fmt.Sprintf("Recurring %s Errors", strings.Title(maxType)),
			Description: fmt.Sprintf("Multiple drones experienced %s errors", maxType),
			Frequency:   maxCount,
			Confidence:  float64(maxCount) / float64(len(results)),
		}
	}
	
	return nil
}

func (da *DataAnalyzer) identifySourceDiversityPattern(results []schemas.DroneResult) *schemas.Pattern {
	uniqueSources := make(map[string]bool)
	totalSources := 0
	
	for _, result := range results {
		if sources, ok := result.Data["sources"].([]interface{}); ok {
			for _, source := range sources {
				if s, ok := source.(string); ok {
					uniqueSources[s] = true
					totalSources++
				}
			}
		}
	}
	
	if totalSources == 0 {
		return nil
	}
	
	diversityRatio := float64(len(uniqueSources)) / float64(totalSources)
	
	if diversityRatio > 0.7 {
		return &schemas.Pattern{
			Name:        "High Source Diversity",
			Description: "Research covered a wide variety of sources",
			Frequency:   len(uniqueSources),
			Confidence:  diversityRatio,
		}
	} else if diversityRatio < 0.3 {
		return &schemas.Pattern{
			Name:        "Source Concentration",
			Description: "Research focused on a limited set of sources",
			Frequency:   totalSources,
			Confidence:  1.0 - diversityRatio,
		}
	}
	
	return nil
}

// Additional analysis methods

func (da *DataAnalyzer) calculateDetailedStatistics(results []schemas.DroneResult) map[string]interface{} {
	stats := da.calculateStatistics(results)
	
	// Add more detailed statistics
	stats["error_rate"] = 1.0 - stats["success_rate"].(float64)
	
	// Calculate percentiles for data volumes
	var volumes []int
	for _, result := range results {
		if result.Status == "completed" {
			volumes = append(volumes, len(result.Data))
		}
	}
	
	if len(volumes) > 0 {
		sort.Ints(volumes)
		stats["data_volume_p50"] = volumes[len(volumes)/2]
		stats["data_volume_p90"] = volumes[int(float64(len(volumes))*0.9)]
		stats["data_volume_min"] = volumes[0]
		stats["data_volume_max"] = volumes[len(volumes)-1]
	}
	
	return stats
}

func (da *DataAnalyzer) identifyDetailedPatterns(results []schemas.DroneResult) []schemas.Pattern {
	patterns := da.identifyPatterns(results)
	
	// Add time-based patterns
	if pattern := da.identifyTimePattern(results); pattern != nil {
		patterns = append(patterns, *pattern)
	}
	
	// Add performance patterns
	if pattern := da.identifyPerformancePattern(results); pattern != nil {
		patterns = append(patterns, *pattern)
	}
	
	return patterns
}

func (da *DataAnalyzer) identifyTimePattern(results []schemas.DroneResult) *schemas.Pattern {
	// Group results by completion time
	hourCounts := make(map[int]int)
	
	for _, result := range results {
		hour := result.CompletedAt.Hour()
		hourCounts[hour]++
	}
	
	// Find peak hours
	maxCount := 0
	peakHour := 0
	for hour, count := range hourCounts {
		if count > maxCount {
			maxCount = count
			peakHour = hour
		}
	}
	
	if maxCount > len(results)/4 { // More than 25% in same hour
		return &schemas.Pattern{
			Name:        fmt.Sprintf("Peak Activity at %02d:00", peakHour),
			Description: "Research activity concentrated during specific time period",
			Frequency:   maxCount,
			Confidence:  float64(maxCount) / float64(len(results)),
		}
	}
	
	return nil
}

func (da *DataAnalyzer) identifyPerformancePattern(results []schemas.DroneResult) *schemas.Pattern {
    avg, _, max := da.analyzeProcessingTimes(results)
	
	if max > avg*3 { // Some drones took much longer
		return &schemas.Pattern{
			Name:        "Performance Variance",
			Description: "Significant variation in drone processing times detected",
			Frequency:   len(results),
			Confidence:  0.75,
		}
	}
	
	return nil
}

func (da *DataAnalyzer) generateDetailedSummary(results []schemas.DroneResult) string {
	summary := da.generateSummary(results)
	
	// Add more details
	summary += fmt.Sprintf("\n\nDetailed Analysis:\n")
	summary += fmt.Sprintf("- Completion rate: %.2f%%\n", da.calculateCompletionRate(results)*100)
	summary += fmt.Sprintf("- Data quality score: %.2f/10\n", da.assessDataQuality(results))
	
	avg, min, max := da.analyzeProcessingTimes(results)
	summary += fmt.Sprintf("- Processing times: avg=%.2fs, min=%.2fs, max=%.2fs\n", 
		avg.Seconds(), min.Seconds(), max.Seconds())
	
	topSources := da.identifyTopSources(results)
	if len(topSources) > 0 {
		summary += fmt.Sprintf("- Top sources: %s\n", strings.Join(topSources, ", "))
	}
	
	return summary
}

func (da *DataAnalyzer) extractTopInsights(results []schemas.DroneResult, count int) []string {
	insights := da.extractInsights(results)
	
	if len(insights) > count {
		return insights[:count]
	}
	
	return insights
}

func (da *DataAnalyzer) generatePatternInsights(patterns []schemas.Pattern) []string {
	insights := []string{}
	
	for _, pattern := range patterns {
		insight := fmt.Sprintf("%s: %s (confidence: %.2f%%)", 
			pattern.Name, pattern.Description, pattern.Confidence*100)
		insights = append(insights, insight)
	}
	
	return insights
}

func (da *DataAnalyzer) generateTimeSeriesData(results []schemas.DroneResult) map[string]interface{} {
	// Group results by time intervals
	timeData := make(map[string]int)
	
	for _, result := range results {
		// Round to nearest hour
		hour := result.CompletedAt.Truncate(time.Hour)
		key := hour.Format("2006-01-02T15:04:05Z")
		timeData[key]++
	}
	
	// Convert to arrays for visualization
	var times []string
	var counts []int
	
	for time, count := range timeData {
		times = append(times, time)
		counts = append(counts, count)
	}
	
	// Sort by time
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})
	
	sortedCounts := make([]int, len(times))
	for i, t := range times {
		sortedCounts[i] = timeData[t]
	}
	
	return map[string]interface{}{
		"timestamps": times,
		"values":     sortedCounts,
	}
}

// GetDescription returns the operation description
func (da *DataAnalyzer) GetDescription() string {
	return "Analyzes research data from multiple drones to identify patterns, generate insights, and produce statistical analysis"
}