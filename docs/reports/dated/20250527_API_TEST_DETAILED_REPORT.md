# API Test Report - Detailed Results
*Generated: 2025-05-27 21:31:02*

**Total Tests**: 18 | **Passed**: 6 | **Failed**: 12 | **Success Rate**: 33.3%

## Test Results

### POST /register - ✅ PASSED
- **Status Code**: 200
- **Duration**: 23.662708ms

### POST /login - ✅ PASSED
- **Status Code**: 200
- **Duration**: 2.920167ms

### POST /api/ai/headlines - ❌ FAILED
- **Status Code**: 401
- **Duration**: 755.167µs
- **Response**: {"error":"Missing token"}

### POST /api/ai/content - ❌ FAILED
- **Status Code**: 401
- **Duration**: 529.292µs
- **Response**: {"error":"Missing token"}

### POST /api/ai/improve - ❌ FAILED
- **Status Code**: 401
- **Duration**: 712.25µs
- **Response**: {"error":"Missing token"}

### POST /api/ai/moderate - ❌ FAILED
- **Status Code**: 401
- **Duration**: 487.125µs
- **Response**: {"error":"Missing token"}

### POST /api/ai/summarize - ❌ FAILED
- **Status Code**: 401
- **Duration**: 352.708µs
- **Response**: {"error":"Missing token"}

### POST /api/ai/categorize - ❌ FAILED
- **Status Code**: 401
- **Duration**: 413.75µs
- **Response**: {"error":"Missing token"}

### GET /api/ai/suggestions - ❌ FAILED
- **Status Code**: 401
- **Duration**: 420.667µs
- **Response**: {"error":"Missing token"}

### GET /api/ai/usage-stats - ❌ FAILED
- **Status Code**: 401
- **Duration**: 357.833µs
- **Response**: {"error":"Missing token"}

### POST /api/agent/tasks - ❌ FAILED
- **Status Code**: 401
- **Duration**: 352.417µs
- **Response**: {"error":"Missing token"}

### GET /api/agent/tasks - ❌ FAILED
- **Status Code**: 401
- **Duration**: 353.416µs
- **Response**: {"error":"Missing token"}

### GET /health - ✅ PASSED
- **Status Code**: 200
- **Duration**: 3.708792ms
- **Response**: {"cache_healthy":true,"db_healthy":true,"status":"healthy","time":"2025-05-27 18:31:02.386697344 +0000 UTC m=+618.086134200","version":"1.0.0"}

### GET /api/articles - ✅ PASSED
- **Status Code**: 200
- **Duration**: 428.125µs

### GET /api/categories - ✅ PASSED
- **Status Code**: 200
- **Duration**: 19.925916ms

### GET /api/tags - ✅ PASSED
- **Status Code**: 200
- **Duration**: 3.021584ms

### GET /api/settings - ❌ FAILED
- **Status Code**: 429
- **Duration**: 872µs
- **Response**: {"error":"Too many requests","message":"Rate limit exceeded. Please try again later.","retry_after":1}

### GET /api/menus - ❌ FAILED
- **Status Code**: 429
- **Duration**: 582.5µs
- **Response**: {"error":"Too many requests","message":"Rate limit exceeded. Please try again later.","retry_after":1}

