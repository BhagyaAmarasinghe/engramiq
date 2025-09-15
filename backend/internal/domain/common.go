package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// JSON type for JSONB fields - this allows us to work with flexible JSON data
// which is crucial for storing equipment specifications and metadata
type JSON map[string]interface{}

// Value implements driver.Valuer interface for database storage
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner interface for database retrieval
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}
	
	// Handle different types that PostgreSQL JSONB might return
	switch v := value.(type) {
	case []byte:
		// Most common case - raw bytes
		if len(v) == 0 {
			*j = make(map[string]interface{})
			return nil
		}
		return json.Unmarshal(v, j)
	
	case string:
		// Sometimes JSONB returns as string
		if v == "" {
			*j = make(map[string]interface{})
			return nil
		}
		return json.Unmarshal([]byte(v), j)
	
	case map[string]interface{}:
		// Already decoded
		*j = v
		return nil
	
	default:
		// Try to handle other cases by converting to JSON first
		b, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("cannot scan type %T into JSON: %v", value, err)
		}
		return json.Unmarshal(b, j)
	}
}

// Point type for spatial data - storing location coordinates for equipment
type Point struct {
	Lng float64 `json:"lng"`
	Lat float64 `json:"lat"`
}

// Value implements driver.Valuer for Point type
func (p Point) Value() (driver.Value, error) {
	return fmt.Sprintf("%.6f,%.6f", p.Lng, p.Lat), nil
}

// Scan implements sql.Scanner for Point type
func (p *Point) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	// Parse "lng,lat" format
	if str, ok := value.(string); ok {
		_, err := fmt.Sscanf(str, "%f,%f", &p.Lng, &p.Lat)
		return err
	}
	return fmt.Errorf("cannot scan %T into Point", value)
}

// Pagination for list requests
type Pagination struct {
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Sort       string `json:"sort"`
	TotalItems int64  `json:"total_items"`
	TotalPages int    `json:"total_pages"`
}

func (p *Pagination) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

func (p *Pagination) SetTotalPages(totalItems int64) {
	p.TotalItems = totalItems
	if p.Limit > 0 {
		p.TotalPages = int((totalItems + int64(p.Limit) - 1) / int64(p.Limit))
	}
}

// Common filter types
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Filter struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}