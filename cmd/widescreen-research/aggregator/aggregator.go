// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package aggregator

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/glassBead-tc/widescreen-research/cmd/widescreen-research-mcp/schemas"
	"github.com/google/uuid"
)

// ReportAggregator handles aggregation of drone results into master reports
// This is where report generation happens in the new architecture (moved from orchestrator)
type ReportAggregator struct {
	// Future: could add AI report generation capabilities here
}

// NewReportAggregator creates a new report aggregator
func NewReportAggregator() *ReportAggregator {
	return &ReportAggregator{}
}

// GenerateReport aggregates drone results into a comprehensive research report
// This is the KEY architectural change - report generation moved to host layer
func (a *ReportAggregator) GenerateReport(ctx context.Context, config *schemas.ResearchConfig, results []schemas.DroneResult) (*schemas.ResearchReport, error) {
	log.Printf("Aggregating %d drone results into master report for session %s", len(results), config.SessionID)

	// 1. Analyze results
	analysis := a.analyzeResults(results)

	// 2. Generate report structure
	report := &schemas.ResearchReport{
		ID:          uuid.New().String(),
		SessionID:   config.SessionID,
		Title:       fmt.Sprintf("Research Report: %s", config.Topic),
		Executive:   a.generateExecutiveSummary(config, results, analysis),
		Sections:    a.generateReportSections(config, results, analysis),
		Methodology: fmt.Sprintf("Distributed research using %d parallel drones with %s depth.", config.ResearcherCount, config.ResearchDepth),
		Data:        a.aggregateDroneData(results),
		CreatedAt:   time.Now(),
		Metadata: schemas.ReportMetadata{
			ResearchTopic:   config.Topic,
			ResearcherCount: config.ResearcherCount,
			Duration:        analysis.Duration,
			DataPoints:      len(results),
			Sources:         a.extractSources(results),
			Metrics:         analysis.Metrics,
		},
	}

	// 3. Save report artifacts
	if err := a.saveReportArtifacts(report, results); err != nil {
		log.Printf("Warning: failed to save report artifacts: %v", err)
	}

	log.Printf("Successfully generated report %s for session %s", report.ID, config.SessionID)
	return report, nil
}

// DataAnalysis contains analysis results from research data
type DataAnalysis struct {
	Patterns          []schemas.Pattern
	TopInsights       []string
	Statistics        map[string]interface{}
	Duration          time.Duration
	AverageConfidence float64
	Metrics           schemas.ResearchMetrics
}

// analyzeResults performs analysis on collected drone results
func (a *ReportAggregator) analyzeResults(results []schemas.DroneResult) *DataAnalysis {
	startTime := time.Now()
	if len(results) > 0 {
		// Find earliest completion time
		for _, result := range results {
			if result.CompletedAt.Before(startTime) && !result.CompletedAt.IsZero() {
				startTime = result.CompletedAt.Add(-result.ProcessingTime)
			}
		}
	}

	successCount := 0
	failureCount := 0
	totalConfidence := 0.0
	confidenceCount := 0

	for _, result := range results {
		if result.Status == "completed" {
			successCount++
		} else {
			failureCount++
		}

		// Extract confidence if available
		if confidence, ok := result.Data["confidence"].(float64); ok {
			totalConfidence += confidence
			confidenceCount++
		}
	}

	avgConfidence := 0.0
	if confidenceCount > 0 {
		avgConfidence = totalConfidence / float64(confidenceCount)
	}

	// Extract patterns (simplified - could use AI for better pattern detection)
	patterns := a.extractPatterns(results)

	// Generate insights
	insights := a.generateInsights(results, patterns)

	return &DataAnalysis{
		Patterns:          patterns,
		TopInsights:       insights,
		Statistics:        a.calculateStatistics(results),
		Duration:          time.Since(startTime),
		AverageConfidence: avgConfidence,
		Metrics: schemas.ResearchMetrics{
			DronesProvisioned:   len(results),
			DronesCompleted:     successCount,
			DronesFailed:        failureCount,
			TotalDuration:       time.Since(startTime),
			DataPointsCollected: len(results),
			CostEstimate:        float64(len(results)) * 0.01, // Simplified cost estimate
		},
	}
}

// extractPatterns identifies patterns in the collected data
func (a *ReportAggregator) extractPatterns(results []schemas.DroneResult) []schemas.Pattern {
	// Simplified pattern extraction - could be enhanced with AI
	patterns := []schemas.Pattern{}

	// Count term frequencies across all results
	termFreq := make(map[string]int)
	for _, result := range results {
		if result.Status == "completed" {
			// Extract terms from data (simplified)
			if terms, ok := result.Data["keywords"].([]interface{}); ok {
				for _, term := range terms {
					if termStr, ok := term.(string); ok {
						termFreq[termStr]++
					}
				}
			}
		}
	}

	// Convert frequent terms to patterns
	for term, freq := range termFreq {
		if freq >= 2 { // Minimum frequency threshold
			patterns = append(patterns, schemas.Pattern{
				Name:        term,
				Description: fmt.Sprintf("Term appeared in %d drone results", freq),
				Frequency:   freq,
				Confidence:  float64(freq) / float64(len(results)),
			})
		}
	}

	// Sort by frequency
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Frequency > patterns[j].Frequency
	})

	return patterns
}

// generateInsights generates key insights from the data
func (a *ReportAggregator) generateInsights(results []schemas.DroneResult, patterns []schemas.Pattern) []string {
	insights := []string{}

	// Add pattern-based insights
	if len(patterns) > 0 {
		insights = append(insights, fmt.Sprintf("Most common pattern: '%s' (appeared %d times)", patterns[0].Name, patterns[0].Frequency))
	}

	// Add coverage insights
	successCount := 0
	for _, result := range results {
		if result.Status == "completed" {
			successCount++
		}
	}
	coverage := float64(successCount) / float64(len(results)) * 100
	insights = append(insights, fmt.Sprintf("Research coverage: %.1f%% (%d/%d drones completed successfully)", coverage, successCount, len(results)))

	// Add data diversity insights
	uniqueSources := len(a.extractSources(results))
	insights = append(insights, fmt.Sprintf("Data collected from %d unique sources", uniqueSources))

	return insights
}

// calculateStatistics calculates statistics from the results
func (a *ReportAggregator) calculateStatistics(results []schemas.DroneResult) map[string]interface{} {
	stats := make(map[string]interface{})

	successCount := 0
	totalProcessingTime := time.Duration(0)

	for _, result := range results {
		if result.Status == "completed" {
			successCount++
			totalProcessingTime += result.ProcessingTime
		}
	}

	stats["total_drones"] = len(results)
	stats["successful_drones"] = successCount
	stats["failed_drones"] = len(results) - successCount
	stats["success_rate"] = float64(successCount) / float64(len(results))

	if successCount > 0 {
		stats["avg_processing_time"] = totalProcessingTime / time.Duration(successCount)
	}

	return stats
}

// generateExecutiveSummary creates an executive summary
func (a *ReportAggregator) generateExecutiveSummary(config *schemas.ResearchConfig, results []schemas.DroneResult, analysis *DataAnalysis) string {
	successCount := 0
	for _, result := range results {
		if result.Status == "completed" {
			successCount++
		}
	}

	summary := fmt.Sprintf("# Research Summary: %s\n\n", config.Topic)
	summary += fmt.Sprintf("Completed using %d research drones over %v.\n\n", config.ResearcherCount, analysis.Duration.Round(time.Second))
	summary += fmt.Sprintf("Successfully collected data from %d out of %d drones (%.1f%% success rate).\n\n",
		successCount, len(results), float64(successCount)/float64(len(results))*100)

	if len(analysis.TopInsights) > 0 {
		summary += "## Key Findings:\n\n"
		for i, insight := range analysis.TopInsights {
			if i >= 5 {
				break
			}
			summary += fmt.Sprintf("- %s\n", insight)
		}
	}

	return summary
}

// generateReportSections creates detailed report sections
func (a *ReportAggregator) generateReportSections(config *schemas.ResearchConfig, results []schemas.DroneResult, analysis *DataAnalysis) []schemas.ReportSection {
	sections := []schemas.ReportSection{
		{
			Title:    "Research Findings",
			Content:  fmt.Sprintf("Collected data from %d research drones. Identified %d patterns with average confidence of %.2f.", len(results), len(analysis.Patterns), analysis.AverageConfidence),
			Insights: analysis.TopInsights,
			Data:     analysis.Statistics,
		},
	}

	// Add pattern section if patterns were found
	if len(analysis.Patterns) > 0 {
		patternData := make(map[string]interface{})
		patternData["patterns"] = analysis.Patterns
		sections = append(sections, schemas.ReportSection{
			Title:   "Identified Patterns",
			Content: fmt.Sprintf("Discovered %d patterns across the research data.", len(analysis.Patterns)),
			Data:    patternData,
		})
	}

	return sections
}

// aggregateDroneData aggregates raw data from all drone results
func (a *ReportAggregator) aggregateDroneData(results []schemas.DroneResult) map[string]interface{} {
	aggregated := make(map[string]interface{})

	var allData []map[string]interface{}
	for _, result := range results {
		if result.Status == "completed" && result.Data != nil {
			allData = append(allData, result.Data)
		}
	}

	aggregated["drone_results"] = allData
	aggregated["total_drones"] = len(results)
	aggregated["successful_drones"] = len(allData)

	return aggregated
}

// extractSources extracts unique sources from drone results
func (a *ReportAggregator) extractSources(results []schemas.DroneResult) []string {
	sourceMap := make(map[string]bool)

	for _, result := range results {
		if sources, ok := result.Data["sources"].([]interface{}); ok {
			for _, source := range sources {
				if s, ok := source.(string); ok {
					sourceMap[s] = true
				}
			}
		}
	}

	sources := make([]string, 0, len(sourceMap))
	for source := range sourceMap {
		sources = append(sources, source)
	}

	return sources
}

// saveReportArtifacts saves report and individual results to disk
func (a *ReportAggregator) saveReportArtifacts(report *schemas.ResearchReport, results []schemas.DroneResult) error {
	// Create reports directory
	reportDir := fmt.Sprintf("reports/session_%s", report.SessionID)
	if err := os.MkdirAll(reportDir, 0755); err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	// Save master report as markdown
	markdownPath := fmt.Sprintf("%s/report.md", reportDir)
	markdownContent := a.renderReportAsMarkdown(report)
	if err := os.WriteFile(markdownPath, []byte(markdownContent), 0644); err != nil {
		return fmt.Errorf("failed to save markdown report: %w", err)
	}

	log.Printf("Saved report artifacts to %s", reportDir)
	return nil
}

// renderReportAsMarkdown converts report to markdown format
func (a *ReportAggregator) renderReportAsMarkdown(report *schemas.ResearchReport) string {
	md := fmt.Sprintf("# %s\n\n", report.Title)
	md += fmt.Sprintf("**Report ID:** %s  \n", report.ID)
	md += fmt.Sprintf("**Session ID:** %s  \n", report.SessionID)
	md += fmt.Sprintf("**Generated:** %s  \n\n", report.CreatedAt.Format(time.RFC3339))

	md += "---\n\n"
	md += report.Executive + "\n\n"

	md += "---\n\n"
	md += "## Methodology\n\n"
	md += report.Methodology + "\n\n"

	for _, section := range report.Sections {
		md += fmt.Sprintf("## %s\n\n", section.Title)
		md += section.Content + "\n\n"

		if len(section.Insights) > 0 {
			md += "### Insights\n\n"
			for _, insight := range section.Insights {
				md += fmt.Sprintf("- %s\n", insight)
			}
			md += "\n"
		}
	}

	md += "---\n\n"
	md += "## Metadata\n\n"
	md += fmt.Sprintf("- **Research Topic:** %s\n", report.Metadata.ResearchTopic)
	md += fmt.Sprintf("- **Researcher Count:** %d\n", report.Metadata.ResearcherCount)
	md += fmt.Sprintf("- **Duration:** %v\n", report.Metadata.Duration.Round(time.Second))
	md += fmt.Sprintf("- **Data Points:** %d\n", report.Metadata.DataPoints)
	md += fmt.Sprintf("- **Sources:** %d unique sources\n", len(report.Metadata.Sources))

	return md
}
