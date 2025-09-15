# Engramiq Solar Asset Reporting Agent API Specification

## API Design Principles

### 1. RESTful Design
- Resources are nouns (e.g., `/sites`, `/components`, `/documents`)
- HTTP verbs define actions (GET, POST, PUT, DELETE)
- Consistent URL patterns
- Stateless communication

### 2. Versioning
- URL-based versioning: `/api/v1/`
- Backward compatibility for minor versions
- Deprecation notices for breaking changes

### 3. Authentication
- Bearer token in Authorization header
- Refresh token in HTTP-only cookie
- API keys for enterprise integrations

### 4. Response Standards

#### Success Response
```json
{
  "success": true,
  "data": {
    // Resource data
  },
  "meta": {
    "timestamp": "2024-01-20T10:30:00Z",
    "version": "1.0.0"
  }
}
```

#### Error Response
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "title",
        "message": "Title is required"
      }
    ]
  },
  "meta": {
    "timestamp": "2024-01-20T10:30:00Z",
    "request_id": "req_123456"
  }
}
```

### 5. HTTP Status Codes
- `200 OK`: Successful GET, PUT
- `201 Created`: Successful POST
- `204 No Content`: Successful DELETE
- `400 Bad Request`: Invalid request
- `401 Unauthorized`: Missing/invalid auth
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict
- `422 Unprocessable Entity`: Validation error
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

## Authentication Endpoints

### Register User
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "full_name": "John Doe"
}

Response: 201 Created
{
  "success": true,
  "data": {
    "user": {
      "id": "usr_123456",
      "email": "user@example.com",
      "full_name": "John Doe",
      "created_at": "2024-01-20T10:30:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

### Login
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}

Response: 200 OK
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": "usr_123456",
      "email": "user@example.com",
      "full_name": "John Doe"
    }
  }
}

Set-Cookie: refresh_token=...; HttpOnly; Secure; SameSite=Strict; Max-Age=604800
```

### Refresh Token
```http
POST /api/v1/auth/refresh
Cookie: refresh_token=...

Response: 200 OK
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 900
  }
}
```

## Sites Endpoints

### List Sites
```http
GET /api/v1/sites?page=1&limit=20
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "sites": [
      {
        "id": "site_123456",
        "site_code": "S2367",
        "name": "University Solar Installation",
        "address": "United States",
        "total_capacity_kw": 756.0,
        "number_of_inverters": 46,
        "installation_date": "2017-05-01",
        "created_at": "2024-01-20T10:30:00Z",
        "updated_at": "2024-01-20T10:30:00Z"
      }
    ]
  },
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 1,
      "total_pages": 1
    }
  }
}
```

### Get Site Details
```http
GET /api/v1/sites/site_123456
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "site": {
      "id": "site_123456",
      "site_code": "S2367",
      "name": "University Solar Installation",
      "address": "United States",
      "total_capacity_kw": 756.0,
      "number_of_inverters": 46,
      "installation_date": "2017-05-01",
      "component_summary": {
        "inverters": 46,
        "combiners": 12,
        "total_components": 58
      },
      "recent_activity_count": 15,
      "site_metadata": {},
      "created_at": "2024-01-20T10:30:00Z"
    }
  }
}
```

## Site Components Endpoints

### List Site Components
```http
GET /api/v1/sites/site_123456/components?type=inverter&page=1&limit=20
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "components": [
      {
        "id": "comp_123456",
        "site_id": "site_123456",
        "external_id": "31",
        "component_type": "inverter",
        "name": "Inverter 31",
        "label": "INV-31",
        "level": 3,
        "group_name": "Auditorium Inverters",
        "current_status": "operational",
        "specifications": {
          "manufacturer": "SOLECTRIA",
          "model": "PVI14TL-480",
          "serial_number": "11491604119"
        },
        "electrical_data": {
          "max_ac_power_kwac": 14,
          "ac_voltage_V": 208,
          "dc_voltage_V": 600
        },
        "drawing_title": "AUDITORIUM ROOF PV PANEL LAYOUT",
        "drawing_number": "PV101",
        "last_maintenance_date": "2024-04-09",
        "created_at": "2024-01-20T10:30:00Z"
      }
    ]
  },
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 46,
      "total_pages": 3
    }
  }
}
```

## Document Processing Endpoints

### Upload Document
```http
POST /api/v1/sites/site_123456/documents
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: (binary PDF, email, or text file)
document_type: "field_service_report"
document_date: "2024-04-09"
author_name: "Technician 1"

Response: 201 Created
{
  "success": true,
  "data": {
    "document": {
      "id": "doc_123456",
      "site_id": "site_123456",
      "document_type": "field_service_report",
      "title": "Inverter 31 Replacement Report",
      "processing_status": "processing",
      "document_date": "2024-04-09",
      "author_name": "Technician 1",
      "original_filename": "S2367_WO_00549595.pdf",
      "file_size": 2048576,
      "created_at": "2024-01-20T10:30:00Z"
    }
  }
}
```

### Get Document Processing Status
```http
GET /api/v1/documents/doc_123456/status
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "document_id": "doc_123456",
    "processing_status": "completed",
    "processing_started_at": "2024-01-20T10:30:00Z",
    "processing_completed_at": "2024-01-20T10:32:15Z",
    "extracted_actions_count": 3,
    "errors": []
  }
}
```

### List Documents
```http
GET /api/v1/sites/site_123456/documents?type=field_service_report&page=1&limit=20
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "documents": [
      {
        "id": "doc_123456",
        "document_type": "field_service_report",
        "title": "Inverter 31 Replacement Report",
        "document_date": "2024-04-09",
        "author_name": "Technician 1",
        "processing_status": "completed",
        "extracted_actions_count": 3,
        "created_at": "2024-01-20T10:30:00Z"
      }
    ]
  }
}
```

## Natural Language Query Endpoints

### Submit Query
```http
POST /api/v1/sites/site_123456/query
Authorization: Bearer <token>
Content-Type: application/json

{
  "query": "What maintenance work was completed on inverter 31 last month?",
  "include_related_events": true
}

Response: 200 OK
{
  "success": true,
  "data": {
    "answer": "Based on the maintenance records, inverter 31 (serial number 11491604119) had the following work completed last month:\n\n1. **April 9, 2024**: Complete inverter replacement due to persistent arc protect faults\n   - Old unit (S/N: 11491604119) was removed\n   - New unit (S/N: 11491543039) was installed\n   - Modbus ID set to 11\n   - Production confirmed both locally and remotely\n\n2. **March 25, 2024**: Arc protect troubleshooting\n   - Individual string testing performed\n   - All strings passed individual tests\n   - Solectria case #440586278 was opened\n   - Unit was scheduled for RMA replacement\n\nThe replacement was completed successfully and the inverter is now operational.",
    "query_type": "maintenance_history",
    "confidence": 0.94,
    "execution_time_ms": 850,
    "sources": [
      {
        "document_id": "doc_123456",
        "document_title": "Inverter 31 Replacement Report",
        "document_date": "2024-04-09",
        "relevant_excerpt": "Replaced failed PVI14TL inverter with new unit...",
        "relevance_score": 0.95
      }
    ],
    "related_events": [
      {
        "event_id": "evt_789012",
        "title": "Inverter 31 Replacement Completed",
        "start_time": "2024-04-09T10:00:00Z",
        "event_type": "maintenance_completed"
      }
    ]
  }
}
```

### Query Suggestions
```http
GET /api/v1/sites/site_123456/query-suggestions?q=inverter
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "suggestions": [
      "What's the status of inverter 31?",
      "Show me all arc fault incidents",
      "Which inverters need maintenance?",
      "What work is scheduled for next week?"
    ]
  }
}
```

## Timeline Visualization Endpoints

### Get Site Timeline
```http
GET /api/v1/sites/site_123456/timeline?start=2024-01-01&end=2024-12-31&types=maintenance,fault
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "timeline": {
      "events": [
        {
          "id": "evt_123456",
          "title": "Inverter 31 Replacement",
          "description": "Replaced failed inverter due to persistent arc protect faults",
          "start_time": "2024-04-09T08:00:00Z",
          "end_time": "2024-04-09T14:00:00Z",
          "event_type": "maintenance_completed",
          "priority": "high",
          "is_future": false,
          "component": {
            "id": "comp_123456",
            "external_id": "31",
            "name": "Inverter 31",
            "component_type": "inverter"
          },
          "sources": [
            {
              "document_id": "doc_123456",
              "document_title": "Inverter 31 Replacement Report"
            }
          ],
          "work_order_number": "00549595",
          "technician_assigned": "Technician 1, Technician 2",
          "metadata": {
            "old_serial": "11491604119",
            "new_serial": "11491543039",
            "modbus_id": 11
          }
        }
      ]
    },
    "summary": {
      "total_events": 45,
      "maintenance_events": 32,
      "fault_events": 8,
      "upcoming_events": 5,
      "critical_events": 2
    }
  }
}
```

### Get Component Timeline
```http
GET /api/v1/components/comp_123456/timeline
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "component": {
      "id": "comp_123456",
      "external_id": "31",
      "name": "Inverter 31",
      "component_type": "inverter"
    },
    "timeline": {
      "events": [
        {
          "id": "evt_123456",
          "title": "Inverter Replacement",
          "start_time": "2024-04-09T08:00:00Z",
          "event_type": "maintenance_completed",
          "work_order_number": "00549595"
        },
        {
          "id": "evt_789012",
          "title": "Arc Protect Fault",
          "start_time": "2024-03-25T14:30:00Z",
          "event_type": "fault_occurred",
          "priority": "high"
        }
      ]
    }
  }
}
```

## Extracted Actions Endpoints

### Get Action Details
```http
GET /api/v1/actions/act_123456
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "action": {
      "id": "act_123456",
      "document_id": "doc_123456",
      "site_id": "site_123456",
      "action_type": "replacement",
      "title": "Inverter 31 Replacement",
      "description": "Replaced failed PVI14TL inverter with new unit due to persistent arc protect faults",
      "action_date": "2024-04-09",
      "action_status": "completed",
      "technician_names": ["Technician 1", "Technician 2"],
      "work_order_number": "00549595",
      "outcome_description": "Inverter replacement completed successfully. Production confirmed locally and remotely with ITS.",
      "measurements": {
        "old_serial": "11491604119",
        "new_serial": "11491543039",
        "modbus_id": 11,
        "installation_time_hours": 6
      },
      "follow_up_actions": [
        "Monitor production for 48 hours",
        "Update asset database with new serial number"
      ],
      "primary_component": {
        "id": "comp_123456",
        "external_id": "31",
        "name": "Inverter 31"
      },
      "extraction_confidence": 0.96,
      "created_at": "2024-01-20T10:32:15Z"
    }
  }
}
```

### List Site Actions
```http
GET /api/v1/sites/site_123456/actions?type=maintenance&component_id=comp_123456&page=1&limit=20
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "actions": [
      {
        "id": "act_123456",
        "action_type": "replacement",
        "title": "Inverter 31 Replacement",
        "action_date": "2024-04-09",
        "technician_names": ["Technician 1", "Technician 2"],
        "work_order_number": "00549595",
        "primary_component": {
          "external_id": "31",
          "name": "Inverter 31"
        }
      }
    ]
  }
}
```

## Search Endpoints

### Full-Text Search
```http
GET /api/v1/sites/site_123456/search?q=arc+fault&type=documents&page=1&limit=20
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "results": [
      {
        "id": "doc_123456",
        "type": "document",
        "title": "Inverter 31 Troubleshooting Report",
        "excerpt": "Inverter was in standby with <em>arc fault</em> error. Cleared fault and inverter started up...",
        "score": 0.89,
        "document_date": "2024-02-06",
        "created_at": "2024-01-20T10:30:00Z"
      }
    ]
  },
  "meta": {
    "query": "arc fault",
    "total_results": 12,
    "search_time_ms": 45
  }
}
```

### Semantic Search
```http
POST /api/v1/sites/site_123456/search/semantic
Authorization: Bearer <token>
Content-Type: application/json

{
  "query": "inverters that frequently trip due to electrical issues",
  "limit": 10,
  "threshold": 0.7
}

Response: 200 OK
{
  "success": true,
  "data": {
    "results": [
      {
        "id": "comp_123456",
        "type": "component",
        "name": "Inverter 31",
        "similarity_score": 0.92,
        "context": "Multiple arc protect faults leading to replacement",
        "recent_issues": ["arc protect", "ground fault", "isolation error"]
      }
    ]
  }
}
```

## Analytics Endpoints

### Site Performance Analytics
```http
GET /api/v1/sites/site_123456/analytics/performance?period=30d
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "period": "30d",
    "metrics": {
      "total_work_orders": 15,
      "completed_actions": 42,
      "fault_incidents": 8,
      "maintenance_hours": 156,
      "most_active_components": [
        {
          "component_id": "comp_123456",
          "name": "Inverter 31",
          "action_count": 5
        }
      ],
      "technician_activity": [
        {
          "technician_name": "Technician 1",
          "work_orders": 8,
          "hours": 64
        }
      ]
    }
  }
}
```

### Query Analytics
```http
GET /api/v1/sites/site_123456/analytics/queries?period=7d
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "total_queries": 156,
    "avg_response_time_ms": 420,
    "most_common_query_types": [
      {
        "type": "component_status",
        "count": 45,
        "percentage": 28.8
      },
      {
        "type": "maintenance_history",
        "count": 38,
        "percentage": 24.4
      }
    ],
    "popular_queries": [
      "What's the status of inverter 31?",
      "Show me recent arc fault incidents",
      "What maintenance is due next week?"
    ]
  }
}
```

## Rate Limiting

### Headers
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1642680000
```

### Rate Limit Exceeded Response
```http
429 Too Many Requests
Retry-After: 60

{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please retry after 60 seconds.",
    "retry_after": 60
  }
}
```

## Pagination

### Query Parameters
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20, max: 100)
- `sort`: Sort field with direction (e.g., `-created_at`, `title`)

### Link Headers
```http
Link: </api/v1/engrams?page=2&limit=20>; rel="next",
      </api/v1/engrams?page=5&limit=20>; rel="last",
      </api/v1/engrams?page=1&limit=20>; rel="first"
```

## Filtering and Sorting

### Filter Operators
- `eq`: Equal to
- `ne`: Not equal to
- `gt`: Greater than
- `gte`: Greater than or equal to
- `lt`: Less than
- `lte`: Less than or equal to
- `in`: In array
- `contains`: Contains substring

### Example
```http
GET /api/v1/engrams?filter[created_at][gte]=2024-01-01&filter[tags][contains]=docker&sort=-created_at
```

## Webhooks

### Webhook Events
- `engram.created`
- `engram.updated`
- `engram.deleted`
- `workspace.created`
- `workspace.member.added`

### Webhook Payload
```json
{
  "id": "evt_123456",
  "type": "engram.created",
  "created": "2024-01-20T10:30:00Z",
  "data": {
    "object": {
      // Full engram object
    }
  }
}
```

### Webhook Security
- HMAC-SHA256 signature in `X-Webhook-Signature` header
- Timestamp in `X-Webhook-Timestamp` header
- Replay protection with 5-minute window

## API Best Practices

### 1. Idempotency
- Use idempotency keys for POST requests
- Header: `Idempotency-Key: unique-key-123`

### 2. Conditional Requests
- Support ETag for caching
- Support If-Modified-Since headers

### 3. Field Selection
- Use `fields` parameter to select specific fields
- Example: `?fields=id,title,created_at`

### 4. Bulk Operations
```http
POST /api/v1/engrams/bulk
Authorization: Bearer <token>
Content-Type: application/json

{
  "operations": [
    {
      "method": "create",
      "data": {
        "title": "Engram 1",
        "content": "..."
      }
    },
    {
      "method": "update",
      "id": "eng_123456",
      "data": {
        "title": "Updated Title"
      }
    }
  ]
}
```

### 5. API Health Check
```http
GET /api/v1/health

Response: 200 OK
{
  "status": "healthy",
  "version": "1.0.0",
  "services": {
    "database": "healthy",
    "redis": "healthy",
    "elasticsearch": "healthy",
    "document_processor": "healthy",
    "llm_service": "healthy"
  }
}
```

## Solar Asset Management Specific Features

### Component Status Tracking
The API provides specialized endpoints for tracking solar equipment status, maintenance history, and performance metrics. Each component (inverter, combiner, etc.) maintains its operational state and related events.

### Document Processing Pipeline
Field service reports, emails, and maintenance documents are automatically processed using AI to extract actionable information and relate it to specific site components.

### Natural Language Querying
Asset managers can query site information using natural language, making it easy to find maintenance history, fault analysis, and operational insights without complex database queries.

### Timeline Visualization
Events are automatically organized on interactive timelines, providing clear visibility into past maintenance activities and upcoming scheduled work.

### Privacy and Compliance
All LLM processing includes privacy controls for enterprise solar asset management companies, with options for data retention, PII stripping, and audit logging.

This API specification enables solar asset managers to efficiently query operational context, reduce manual data compilation, and focus on optimizing site performance rather than gathering information.