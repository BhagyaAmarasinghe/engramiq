package service

import (
	"regexp"
	"strings"
)

type ContentFilterService interface {
	ValidateQuery(queryText string) (*QueryValidationResult, error)
	SanitizeResponse(response string) string
	IsAppropriateQuery(queryText string) bool
	EnforceProfessionalTone(response string) string
}

type QueryValidationResult struct {
	IsValid      bool     `json:"is_valid"`
	IsAppropriate bool    `json:"is_appropriate"`
	Issues       []string `json:"issues"`
	Reason       string   `json:"reason,omitempty"`
}

type contentFilterService struct {
	inappropriatePatterns []*regexp.Regexp
	personalPatterns     []*regexp.Regexp
	offtopicPatterns     []*regexp.Regexp
}

func NewContentFilterService() ContentFilterService {
	return &contentFilterService{
		inappropriatePatterns: compileInappropriatePatterns(),
		personalPatterns:     compilePersonalPatterns(),
		offtopicPatterns:     compileOffTopicPatterns(),
	}
}

func (s *contentFilterService) ValidateQuery(queryText string) (*QueryValidationResult, error) {
	result := &QueryValidationResult{
		IsValid:       true,
		IsAppropriate: true,
		Issues:        []string{},
	}

	queryLower := strings.ToLower(strings.TrimSpace(queryText))

	// Check for inappropriate content
	if !s.IsAppropriateQuery(queryText) {
		result.IsValid = false
		result.IsAppropriate = false
		result.Issues = append(result.Issues, "inappropriate_content")
		result.Reason = "Query contains inappropriate content"
		return result, nil
	}

	// Check for personal/flirtatious content
	for _, pattern := range s.personalPatterns {
		if pattern.MatchString(queryLower) {
			result.IsValid = false
			result.IsAppropriate = false
			result.Issues = append(result.Issues, "personal_content")
			result.Reason = "Query contains personal or inappropriate personal interaction"
			return result, nil
		}
	}

	// Check if query is off-topic (not related to solar asset management)
	if s.isOffTopic(queryLower) {
		result.IsValid = false
		result.Issues = append(result.Issues, "off_topic")
		result.Reason = "Query is not related to solar asset management"
		return result, nil
	}

	// Check query length and complexity
	if len(queryText) < 3 {
		result.IsValid = false
		result.Issues = append(result.Issues, "too_short")
		result.Reason = "Query is too short to process meaningfully"
		return result, nil
	}

	if len(queryText) > 1000 {
		result.IsValid = false
		result.Issues = append(result.Issues, "too_long")
		result.Reason = "Query exceeds maximum length limit"
		return result, nil
	}

	return result, nil
}

func (s *contentFilterService) IsAppropriateQuery(queryText string) bool {
	queryLower := strings.ToLower(queryText)

	// Check against inappropriate patterns
	for _, pattern := range s.inappropriatePatterns {
		if pattern.MatchString(queryLower) {
			return false
		}
	}

	return true
}

func (s *contentFilterService) SanitizeResponse(response string) string {
	// Remove any potentially sensitive information that might have leaked through
	sanitized := response

	// Remove email addresses
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	sanitized = emailRegex.ReplaceAllString(sanitized, "[EMAIL_REDACTED]")

	// Remove phone numbers
	phoneRegex := regexp.MustCompile(`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`)
	sanitized = phoneRegex.ReplaceAllString(sanitized, "[PHONE_REDACTED]")

	// Remove social security numbers or similar patterns
	ssnRegex := regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	sanitized = ssnRegex.ReplaceAllString(sanitized, "[SSN_REDACTED]")

	return sanitized
}

func (s *contentFilterService) EnforceProfessionalTone(response string) string {
	// Ensure responses maintain professional tone per PRD requirements
	
	// Remove overly casual language
	response = strings.ReplaceAll(response, " awesome ", " excellent ")
	response = strings.ReplaceAll(response, " cool ", " good ")
	response = strings.ReplaceAll(response, " nice ", " appropriate ")
	
	// Avoid sycophantic language
	response = strings.ReplaceAll(response, "You're amazing", "I can help you with")
	response = strings.ReplaceAll(response, "Great question", "Regarding your query")
	
	// Ensure professional closing
	if !strings.Contains(response, "additional information") && 
	   !strings.Contains(response, "further assistance") {
		response += " Please let me know if you need additional information about your solar assets."
	}

	return response
}

func (s *contentFilterService) isOffTopic(queryLower string) bool {
	// Check if query is related to solar asset management
	solarKeywords := []string{
		"solar", "inverter", "panel", "module", "combiner", "site", "maintenance",
		"repair", "performance", "power", "energy", "electrical", "component",
		"asset", "facility", "installation", "inspection", "troubleshoot",
		"warranty", "o&m", "operations", "pv", "photovoltaic", "string",
		"transformer", "monitoring", "generation", "output", "failure",
	}

	// If query contains solar-related keywords, it's likely on-topic
	for _, keyword := range solarKeywords {
		if strings.Contains(queryLower, keyword) {
			return false
		}
	}

	// Check against known off-topic patterns
	for _, pattern := range s.offtopicPatterns {
		if pattern.MatchString(queryLower) {
			return true
		}
	}

	// If no solar keywords found and query is substantial, might be off-topic
	words := strings.Fields(queryLower)
	if len(words) > 5 {
		// For longer queries without solar keywords, flag as potentially off-topic
		// This is conservative - in production you'd use more sophisticated NLP
		return true
	}

	return false
}

// Pattern compilation functions
func compileInappropriatePatterns() []*regexp.Regexp {
	patterns := []string{
		`\b(sexy|hot|beautiful|gorgeous|handsome)\b`,
		`\b(love|romance|dating|marry|kiss)\b`,
		`\b(personal|private|intimate)\b.*\b(life|details|information)\b`,
	}
	
	var compiledPatterns []*regexp.Regexp
	for _, pattern := range patterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			compiledPatterns = append(compiledPatterns, compiled)
		}
	}
	return compiledPatterns
}

func compilePersonalPatterns() []*regexp.Regexp {
	patterns := []string{
		`\b(are you single|do you date|want to meet|personal life)\b`,
		`\b(you're (so|very) (smart|helpful|amazing|wonderful))\b`,
		`\b(i love you|you're perfect|marry me)\b`,
		`\b(what do you look like|send me a photo)\b`,
	}
	
	var compiledPatterns []*regexp.Regexp
	for _, pattern := range patterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			compiledPatterns = append(compiledPatterns, compiled)
		}
	}
	return compiledPatterns
}

func compileOffTopicPatterns() []*regexp.Regexp {
	patterns := []string{
		`\b(weather|sports|politics|entertainment|celebrity)\b`,
		`\b(recipe|cooking|food|restaurant)\b`,
		`\b(movie|music|game|television|tv show)\b`,
		`\b(vacation|travel|holiday|tourism)\b`,
		`\b(stock market|cryptocurrency|bitcoin|trading)\b.*(?!solar|energy|renewable)`,
	}
	
	var compiledPatterns []*regexp.Regexp
	for _, pattern := range patterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			compiledPatterns = append(compiledPatterns, compiled)
		}
	}
	return compiledPatterns
}