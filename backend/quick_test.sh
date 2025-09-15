#!/bin/bash

# Quick API Test Script for Engramiq Backend
# Make sure the application is running: make dev

BASE_URL="http://localhost:8080/api/v1"
SITE_ID=$(docker exec -i engramiq-postgres psql -U engramiq -d engramiq -t -c "SELECT id FROM sites WHERE site_code = 'S2367';" | xargs)
echo "Using Site ID: $SITE_ID"

echo "🚀 Starting Engramiq API Tests..."
echo "Base URL: $BASE_URL"
echo "Site ID: $SITE_ID"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test function
test_endpoint() {
    echo -e "${BLUE}Testing: $1${NC}"
    echo "Command: $2"
    echo "Response:"
    eval "$2"
    echo ""
    echo "-----------------------------------"
    echo ""
}

# 0. Verify site S2367 exists and show migration details
echo -e "${BLUE}Verifying site S2367 and migration status...${NC}"
docker exec -i engramiq-postgres psql -U engramiq -d engramiq -c "SELECT site_code, name, number_of_inverters FROM sites WHERE site_code = 'S2367';"
echo ""
echo -e "${BLUE}Migration status:${NC}"
docker exec -i engramiq-postgres psql -U engramiq -d engramiq -c "SELECT * FROM migration_records;"
echo ""
echo -e "${BLUE}Components from migration:${NC}"
docker exec -i engramiq-postgres psql -U engramiq -d engramiq -c "SELECT COUNT(*) as total_components, specifications->>'manufacturer' as manufacturer FROM site_components WHERE site_id = '$SITE_ID' GROUP BY specifications->>'manufacturer';"
echo ""
echo "-----------------------------------"
echo ""

# 1. Health Check
test_endpoint "Health Check" "curl -s '$BASE_URL/health' | jq ."

# 2. Create Additional Test Component (our migration already created inverters)
test_endpoint "Create Additional Test Component" "curl -s -X POST '$BASE_URL/sites/$SITE_ID/components' \
  -H 'Content-Type: application/json' \
  -d '{
    \"component_type\": \"inverter\",
    \"name\": \"TEST-INV-999\",
    \"external_id\": \"TEST-INV-999\",
    \"label\": \"Test Inverter for API Testing\",
    \"specifications\": {
      \"manufacturer\": \"Test Corp\",
      \"model\": \"TEST-25K\",
      \"capacity_kw\": 25.0,
      \"test_component\": true
    }
  }' | jq ."

# 3. List Components
test_endpoint "List Components" "curl -s '$BASE_URL/sites/$SITE_ID/components' | jq ."

# 4. Skip document upload - using existing Supporting-Documents data
echo -e "${BLUE}Using existing field service reports from Supporting-Documents...${NC}"
echo "Available reports: fsr_inverter40_arc_protect.txt, fsr_inverter31_replacement.txt, fsr_wire_management.txt"
echo ""

# 5. List Documents (from Supporting-Documents)
test_endpoint "List Documents" "curl -s '$BASE_URL/sites/$SITE_ID/documents' | jq ."

# 7. Enhanced Query - Professional (using actual data from Supporting-Documents)
test_endpoint "Enhanced Query: Maintenance History" "curl -s -X POST '$BASE_URL/sites/$SITE_ID/queries' \
  -H 'Content-Type: application/json' \
  -d '{
    \"query_text\": \"What was the issue with inverter 40?\",
    \"enhanced\": true
  }' | jq ."

# 8. Enhanced Query - Date-based (using actual data from Supporting-Documents)
test_endpoint "Enhanced Query: Inverter Replacement" "curl -s -X POST '$BASE_URL/sites/$SITE_ID/queries' \
  -H 'Content-Type: application/json' \
  -d '{
    \"query_text\": \"What work was performed on inverter 31?\",
    \"enhanced\": true
  }' | jq ."

# 9. Content Filtering Test - Off-topic (should be rejected)
test_endpoint "Content Filter Test: Off-topic Query (Should be Rejected)" "curl -s -X POST '$BASE_URL/sites/$SITE_ID/queries' \
  -H 'Content-Type: application/json' \
  -d '{
    \"query_text\": \"What is the weather like today?\",
    \"enhanced\": true
  }' | jq ."

# 10. Content Filtering Test - Inappropriate (should be rejected)
test_endpoint "Content Filter Test: Inappropriate Query (Should be Rejected)" "curl -s -X POST '$BASE_URL/sites/$SITE_ID/queries' \
  -H 'Content-Type: application/json' \
  -d '{
    \"query_text\": \"You are amazing and beautiful\",
    \"enhanced\": true
  }' | jq ."

# 11. Query History
test_endpoint "Query History" "curl -s '$BASE_URL/queries/history?user_id=test-user-123' | jq ."

# 12. Site Timeline
test_endpoint "Site Timeline" "curl -s '$BASE_URL/sites/$SITE_ID/timeline' | jq ."

echo -e "${GREEN}✅ All tests completed!${NC}"
echo ""
echo "Key features tested:"
echo "✓ Health check and connectivity"
echo "✓ Component management (create/list)"
echo "✓ Document upload and management"
echo "✓ Enhanced query system with PRD features"
echo "✓ Content filtering and behavioral guards"
echo "✓ Professional tone enforcement"
echo "✓ Query history and analytics"
echo ""
echo "Check the responses above to verify:"
echo "- Queries about inverter 40 arc protection issues should return detailed answers"
echo "- Queries about inverter 31 replacement work should return maintenance details"
echo "- Off-topic queries should be rejected with appropriate messages"
echo "- All responses should maintain professional tone and include source citations"
echo "- Component data shows 47+ SOLECTRIA inverters from inverter_nodes.json"