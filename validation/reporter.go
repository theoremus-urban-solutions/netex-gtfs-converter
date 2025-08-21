package validation

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"sort"
	"strings"
	"time"
)

// ReportFormat represents different output formats for validation reports
type ReportFormat int

const (
	FormatJSON ReportFormat = iota
	FormatHTML
	FormatText
	FormatCSV
	FormatMarkdown
)

// Reporter handles the generation and formatting of validation reports
type Reporter struct {
	config ReporterConfig
}

// ReporterConfig controls report generation behavior
type ReporterConfig struct {
	IncludeProcessingStats bool                 `json:"include_processing_stats"`
	IncludeDetailedIssues  bool                 `json:"include_detailed_issues"`
	GroupBySeverity        bool                 `json:"group_by_severity"`
	GroupByEntityType      bool                 `json:"group_by_entity_type"`
	MaxIssuesPerGroup      int                  `json:"max_issues_per_group"`
	SeverityFilter         []ValidationSeverity `json:"severity_filter"`
	EntityTypeFilter       []string             `json:"entity_type_filter"`
}

// NewReporter creates a new reporter with default configuration
func NewReporter() *Reporter {
	config := ReporterConfig{
		IncludeProcessingStats: true,
		IncludeDetailedIssues:  true,
		GroupBySeverity:        true,
		GroupByEntityType:      false,
		MaxIssuesPerGroup:      50,
		SeverityFilter:         []ValidationSeverity{SeverityInfo, SeverityWarning, SeverityError, SeverityCritical},
		EntityTypeFilter:       []string{}, // Empty means include all
	}

	return &Reporter{config: config}
}

// SetConfig updates the reporter configuration
func (r *Reporter) SetConfig(config ReporterConfig) {
	r.config = config
}

// GenerateReport generates a validation report in the specified format
func (r *Reporter) GenerateReport(report ValidationReport, format ReportFormat, writer io.Writer) error {
	// Filter report based on configuration
	filteredReport := r.filterReport(report)

	switch format {
	case FormatJSON:
		return r.generateJSONReport(filteredReport, writer)
	case FormatHTML:
		return r.generateHTMLReport(filteredReport, writer)
	case FormatText:
		return r.generateTextReport(filteredReport, writer)
	case FormatCSV:
		return r.generateCSVReport(filteredReport, writer)
	case FormatMarkdown:
		return r.generateMarkdownReport(filteredReport, writer)
	default:
		return fmt.Errorf("unsupported report format: %d", format)
	}
}

// filterReport applies configuration filters to the report
func (r *Reporter) filterReport(report ValidationReport) ValidationReport {
	filteredIssues := make([]ValidationIssue, 0)

	for _, issue := range report.Issues {
		// Filter by severity
		if !r.containsSeverity(issue.Severity) {
			continue
		}

		// Filter by entity type
		if len(r.config.EntityTypeFilter) > 0 && !r.containsEntityType(issue.EntityType) {
			continue
		}

		filteredIssues = append(filteredIssues, issue)
	}

	// Update summary with filtered data
	report.Issues = filteredIssues
	report.Summary = r.generateFilteredSummary(filteredIssues)

	return report
}

// containsSeverity checks if severity is in the filter
func (r *Reporter) containsSeverity(severity ValidationSeverity) bool {
	for _, s := range r.config.SeverityFilter {
		if s == severity {
			return true
		}
	}
	return false
}

// containsEntityType checks if entity type is in the filter
func (r *Reporter) containsEntityType(entityType string) bool {
	for _, et := range r.config.EntityTypeFilter {
		if et == entityType {
			return true
		}
	}
	return false
}

// generateFilteredSummary recalculates summary for filtered issues
func (r *Reporter) generateFilteredSummary(issues []ValidationIssue) ValidationSummary {
	summary := ValidationSummary{
		TotalIssues:  len(issues),
		BySeverity:   make(map[ValidationSeverity]int),
		ByCategory:   make(map[string]int),
		ByEntityType: make(map[string]int),
		IsValid:      true,
		HasCritical:  false,
		HasErrors:    false,
	}

	for _, issue := range issues {
		summary.BySeverity[issue.Severity]++
		summary.ByEntityType[issue.EntityType]++

		// Extract category from issue code
		parts := strings.Split(issue.Code, "_")
		if len(parts) > 0 {
			summary.ByCategory[parts[0]]++
		}

		// Update validation status
		if issue.Severity == SeverityCritical {
			summary.HasCritical = true
			summary.IsValid = false
		}
		if issue.Severity == SeverityError {
			summary.HasErrors = true
			summary.IsValid = false
		}
	}

	return summary
}

// generateJSONReport generates a JSON format report
func (r *Reporter) generateJSONReport(report ValidationReport, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// writef is a helper function to handle fmt.Fprintf errors
func (r *Reporter) writef(writer io.Writer, format string, args ...interface{}) error {
	_, err := fmt.Fprintf(writer, format, args...)
	return err
}

// generateTextReport generates a plain text format report
func (r *Reporter) generateTextReport(report ValidationReport, writer io.Writer) error {
	// Write header
	if err := r.writef(writer, "=== NeTEx to GTFS Validation Report ===\n"); err != nil {
		return err
	}
	if err := r.writef(writer, "Generated: %s\n", report.Timestamp.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := r.writef(writer, "\n"); err != nil {
		return err
	}

	// Write summary
	if err := r.writef(writer, "=== SUMMARY ===\n"); err != nil {
		return err
	}
	if err := r.writef(writer, "Total Issues: %d\n", report.Summary.TotalIssues); err != nil {
		return err
	}
	if err := r.writef(writer, "Validation Status: %s\n", r.getValidationStatusText(report.Summary)); err != nil {
		return err
	}
	if err := r.writef(writer, "\n"); err != nil {
		return err
	}

	// Write severity breakdown
	if err := r.writef(writer, "Issues by Severity:\n"); err != nil {
		return err
	}
	severities := []ValidationSeverity{SeverityCritical, SeverityError, SeverityWarning, SeverityInfo}
	for _, severity := range severities {
		if count, exists := report.Summary.BySeverity[severity]; exists && count > 0 {
			if err := r.writef(writer, "  %s: %d\n", severity.String(), count); err != nil {
				return err
			}
		}
	}
	if err := r.writef(writer, "\n"); err != nil {
		return err
	}

	// Write entity type breakdown
	if len(report.Summary.ByEntityType) > 0 {
		if err := r.writef(writer, "Issues by Entity Type:\n"); err != nil {
			return err
		}
		for entityType, count := range report.Summary.ByEntityType {
			if err := r.writef(writer, "  %s: %d\n", entityType, count); err != nil {
				return err
			}
		}
		if err := r.writef(writer, "\n"); err != nil {
			return err
		}
	}

	// Write processing statistics
	if r.config.IncludeProcessingStats {
		if err := r.writef(writer, "=== PROCESSING STATISTICS ===\n"); err != nil {
			return err
		}
		if err := r.writef(writer, "Overall Conversion Rate: %.2f%%\n", report.ProcessingStats.ConversionRate); err != nil {
			return err
		}
		if err := r.writef(writer, "Processing Duration: %v\n", report.ProcessingStats.ProcessingDuration); err != nil {
			return err
		}
		if err := r.writef(writer, "Memory Usage: %.2f MB\n", report.ProcessingStats.MemoryUsageMB); err != nil {
			return err
		}
		if err := r.writef(writer, "\n"); err != nil {
			return err
		}

		if len(report.ProcessingStats.EntitiesProcessed) > 0 {
			if err := r.writef(writer, "Entity Processing Details:\n"); err != nil {
				return err
			}
			for entityType, processed := range report.ProcessingStats.EntitiesProcessed {
				converted := report.ProcessingStats.EntitiesConverted[entityType]
				skipped := report.ProcessingStats.EntitiesSkipped[entityType]
				rate := float64(0)
				if processed > 0 {
					rate = float64(converted) / float64(processed) * 100
				}
				if err := r.writef(writer, "  %s: %d processed, %d converted, %d skipped (%.1f%% success)\n",
					entityType, processed, converted, skipped, rate); err != nil {
					return err
				}
			}
			if err := r.writef(writer, "\n"); err != nil {
				return err
			}
		}
	}

	// Write detailed issues
	if r.config.IncludeDetailedIssues && len(report.Issues) > 0 {
		if err := r.writef(writer, "=== DETAILED ISSUES ===\n"); err != nil {
			return err
		}

		switch {
		case r.config.GroupBySeverity:
			r.writeGroupedIssues(writer, report.Issues, func(issue ValidationIssue) string {
				return issue.Severity.String()
			})
		case r.config.GroupByEntityType:
			r.writeGroupedIssues(writer, report.Issues, func(issue ValidationIssue) string {
				return issue.EntityType
			})
		default:
			r.writeIssuesList(writer, report.Issues, "All Issues")
		}
	}

	return nil
}

// writeGroupedIssues writes issues grouped by a key function
func (r *Reporter) writeGroupedIssues(writer io.Writer, issues []ValidationIssue, keyFunc func(ValidationIssue) string) {
	groups := make(map[string][]ValidationIssue)

	// Group issues
	for _, issue := range issues {
		key := keyFunc(issue)
		groups[key] = append(groups[key], issue)
	}

	// Sort group keys
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Write each group
	for _, key := range keys {
		groupIssues := groups[key]
		r.writeIssuesList(writer, groupIssues, key)
	}
}

// writeIssuesList writes a list of issues under a heading
func (r *Reporter) writeIssuesList(writer io.Writer, issues []ValidationIssue, heading string) {
	if err := r.writef(writer, "\n--- %s (%d issues) ---\n", heading, len(issues)); err != nil {
		return
	}

	// Limit issues per group if configured
	displayIssues := issues
	if r.config.MaxIssuesPerGroup > 0 && len(issues) > r.config.MaxIssuesPerGroup {
		displayIssues = issues[:r.config.MaxIssuesPerGroup]
		if err := r.writef(writer, "(Showing first %d of %d issues)\n", r.config.MaxIssuesPerGroup, len(issues)); err != nil {
			return
		}
	}

	for i, issue := range displayIssues {
		if err := r.writef(writer, "\n%d. [%s] %s\n", i+1, issue.Severity.String(), issue.Message); err != nil {
			return
		}

		if issue.EntityID != "" {
			if err := r.writef(writer, "   Entity: %s (%s)\n", issue.EntityID, issue.EntityType); err != nil {
				return
			}
		}

		if issue.Field != "" {
			if err := r.writef(writer, "   Field: %s", issue.Field); err != nil {
				return
			}
			if issue.Value != "" {
				if err := r.writef(writer, " = '%s'", issue.Value); err != nil {
					return
				}
			}
			if err := r.writef(writer, "\n"); err != nil {
				return
			}
		}

		if issue.Suggestion != "" {
			if err := r.writef(writer, "   Suggestion: %s\n", issue.Suggestion); err != nil {
				return
			}
		}

		if issue.Location != "" {
			if err := r.writef(writer, "   Location: %s\n", issue.Location); err != nil {
				return
			}
		}

		if len(issue.Context) > 0 {
			if err := r.writef(writer, "   Context: "); err != nil {
				return
			}
			contextPairs := make([]string, 0, len(issue.Context))
			for k, v := range issue.Context {
				contextPairs = append(contextPairs, fmt.Sprintf("%s=%s", k, v))
			}
			if err := r.writef(writer, "%s\n", strings.Join(contextPairs, ", ")); err != nil {
				return
			}
		}
	}
}

// generateHTMLReport generates an HTML format report
func (r *Reporter) generateHTMLReport(report ValidationReport, writer io.Writer) error {
	tmpl := template.Must(template.New("report").Parse(htmlTemplate))

	data := struct {
		Report      ValidationReport
		Config      ReporterConfig
		GeneratedAt string
	}{
		Report:      report,
		Config:      r.config,
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
	}

	return tmpl.Execute(writer, data)
}

// generateMarkdownReport generates a Markdown format report
func (r *Reporter) generateMarkdownReport(report ValidationReport, writer io.Writer) error {
	// Write header
	if err := r.writef(writer, "# NeTEx to GTFS Validation Report\n\n"); err != nil {
		return err
	}
	if err := r.writef(writer, "**Generated:** %s\n\n", report.Timestamp.Format(time.RFC3339)); err != nil {
		return err
	}

	// Write summary
	if err := r.writef(writer, "## Summary\n\n"); err != nil {
		return err
	}
	if err := r.writef(writer, "- **Total Issues:** %d\n", report.Summary.TotalIssues); err != nil {
		return err
	}
	if err := r.writef(writer, "- **Validation Status:** %s\n", r.getValidationStatusText(report.Summary)); err != nil {
		return err
	}
	if err := r.writef(writer, "- **Conversion Rate:** %.2f%%\n\n", report.ProcessingStats.ConversionRate); err != nil {
		return err
	}

	// Write severity breakdown table
	if err := r.writef(writer, "### Issues by Severity\n\n"); err != nil {
		return err
	}
	if err := r.writef(writer, "| Severity | Count |\n"); err != nil {
		return err
	}
	if err := r.writef(writer, "|----------|-------|\n"); err != nil {
		return err
	}
	severities := []ValidationSeverity{SeverityCritical, SeverityError, SeverityWarning, SeverityInfo}
	for _, severity := range severities {
		if count, exists := report.Summary.BySeverity[severity]; exists {
			if err := r.writef(writer, "| %s | %d |\n", severity.String(), count); err != nil {
				return err
			}
		}
	}
	if err := r.writef(writer, "\n"); err != nil {
		return err
	}

	// Write detailed issues
	if r.config.IncludeDetailedIssues && len(report.Issues) > 0 {
		if err := r.writef(writer, "## Detailed Issues\n\n"); err != nil {
			return err
		}

		for i, issue := range report.Issues {
			if r.config.MaxIssuesPerGroup > 0 && i >= r.config.MaxIssuesPerGroup {
				if err := r.writef(writer, "*...and %d more issues*\n", len(report.Issues)-i); err != nil {
					return err
				}
				break
			}

			if err := r.writef(writer, "### %d. %s\n\n", i+1, issue.Message); err != nil {
				return err
			}
			if err := r.writef(writer, "- **Severity:** %s\n", issue.Severity.String()); err != nil {
				return err
			}
			if err := r.writef(writer, "- **Code:** %s\n", issue.Code); err != nil {
				return err
			}

			if issue.EntityID != "" {
				if err := r.writef(writer, "- **Entity:** %s (%s)\n", issue.EntityID, issue.EntityType); err != nil {
					return err
				}
			}

			if issue.Field != "" {
				if err := r.writef(writer, "- **Field:** %s", issue.Field); err != nil {
					return err
				}
				if issue.Value != "" {
					if err := r.writef(writer, " = `%s`", issue.Value); err != nil {
						return err
					}
				}
				if err := r.writef(writer, "\n"); err != nil {
					return err
				}
			}

			if issue.Suggestion != "" {
				if err := r.writef(writer, "- **Suggestion:** %s\n", issue.Suggestion); err != nil {
					return err
				}
			}

			if err := r.writef(writer, "\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

// generateCSVReport generates a CSV format report
func (r *Reporter) generateCSVReport(report ValidationReport, writer io.Writer) error {
	// Write CSV header
	if err := r.writef(writer, "Severity,Code,Message,EntityType,EntityID,Field,Value,Suggestion,Location\n"); err != nil {
		return err
	}

	// Write issues
	for _, issue := range report.Issues {
		if err := r.writef(writer, "%s,%s,\"%s\",%s,%s,%s,\"%s\",\"%s\",%s\n",
			issue.Severity.String(),
			issue.Code,
			strings.ReplaceAll(issue.Message, "\"", "\"\""), // Escape quotes
			issue.EntityType,
			issue.EntityID,
			issue.Field,
			strings.ReplaceAll(issue.Value, "\"", "\"\""),
			strings.ReplaceAll(issue.Suggestion, "\"", "\"\""),
			issue.Location,
		); err != nil {
			return err
		}
	}

	return nil
}

// getValidationStatusText returns a human-readable validation status
func (r *Reporter) getValidationStatusText(summary ValidationSummary) string {
	if summary.HasCritical {
		return "❌ CRITICAL ERRORS FOUND"
	}
	if summary.HasErrors {
		return "⚠️ ERRORS FOUND"
	}
	if summary.TotalIssues > 0 {
		return "⚠️ WARNINGS FOUND"
	}
	return "✅ VALID"
}

// HTML template for report generation
const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>NeTEx to GTFS Validation Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { border-bottom: 2px solid #333; padding-bottom: 20px; margin-bottom: 30px; }
        .summary { background: #f5f5f5; padding: 20px; border-radius: 5px; margin-bottom: 30px; }
        .issue { border: 1px solid #ddd; padding: 15px; margin: 10px 0; border-radius: 3px; }
        .critical { border-left: 5px solid #d32f2f; }
        .error { border-left: 5px solid #f57c00; }
        .warning { border-left: 5px solid #fbc02d; }
        .info { border-left: 5px solid #1976d2; }
        .severity { font-weight: bold; text-transform: uppercase; }
        .entity { color: #666; }
        .field { font-family: monospace; background: #f0f0f0; padding: 2px 4px; }
        .value { font-family: monospace; background: #e8f5e8; padding: 2px 4px; }
        .suggestion { color: #2e7d32; font-style: italic; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>NeTEx to GTFS Validation Report</h1>
        <p><strong>Generated:</strong> {{.GeneratedAt}}</p>
    </div>
    
    <div class="summary">
        <h2>Summary</h2>
        <table>
            <tr><td><strong>Total Issues:</strong></td><td>{{.Report.Summary.TotalIssues}}</td></tr>
            <tr><td><strong>Validation Status:</strong></td><td>{{if .Report.Summary.IsValid}}✅ VALID{{else}}❌ ISSUES FOUND{{end}}</td></tr>
            <tr><td><strong>Conversion Rate:</strong></td><td>{{printf "%.2f" .Report.ProcessingStats.ConversionRate}}%</td></tr>
        </table>
        
        {{if .Report.Summary.BySeverity}}
        <h3>Issues by Severity</h3>
        <table>
            <tr><th>Severity</th><th>Count</th></tr>
            {{range $severity, $count := .Report.Summary.BySeverity}}
            <tr><td>{{$severity}}</td><td>{{$count}}</td></tr>
            {{end}}
        </table>
        {{end}}
    </div>
    
    {{if and .Config.IncludeDetailedIssues .Report.Issues}}
    <h2>Detailed Issues</h2>
    {{range $index, $issue := .Report.Issues}}
    {{if lt $index $.Config.MaxIssuesPerGroup}}
    <div class="issue {{if eq $issue.Severity 3}}critical{{else if eq $issue.Severity 2}}error{{else if eq $issue.Severity 1}}warning{{else}}info{{end}}">
        <div class="severity">{{$issue.Severity}}</div>
        <h3>{{$issue.Message}}</h3>
        <p><strong>Code:</strong> {{$issue.Code}}</p>
        {{if $issue.EntityID}}<p class="entity"><strong>Entity:</strong> {{$issue.EntityID}} ({{$issue.EntityType}})</p>{{end}}
        {{if $issue.Field}}<p><strong>Field:</strong> <span class="field">{{$issue.Field}}</span>{{if $issue.Value}} = <span class="value">{{$issue.Value}}</span>{{end}}</p>{{end}}
        {{if $issue.Suggestion}}<p class="suggestion"><strong>Suggestion:</strong> {{$issue.Suggestion}}</p>{{end}}
    </div>
    {{end}}
    {{end}}
    {{end}}
</body>
</html>
`
