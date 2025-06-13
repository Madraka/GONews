# AI Integration API Documentation

## Overview

The News API now includes comprehensive AI integration capabilities to assist editors, writers, and commenters with content generation, moderation, and analysis. This includes both direct AI endpoints and an Agent API for n8n workflow automation.

## Features

### 🤖 AI-Powered Content Assistance
- **Smart Headlines**: AI-generated headlines based on article content
- **Content Generation**: Full article writing assistance with customizable style and tone
- **Content Improvement**: Suggestions for enhancing existing content
- **Automated Summarization**: Generate concise summaries of long-form content
- **Content Categorization**: Intelligent topic and category detection

### 🛡️ Content Moderation
- **AI Moderation**: Automated content screening for inappropriate material
- **Confidence Scoring**: AI confidence levels for moderation decisions
- **Category Detection**: Identification of specific content violation types
- **Severity Assessment**: Risk level evaluation (low, medium, high, critical)

### 📊 Analytics & Insights
- **Usage Statistics**: Track AI service consumption and costs
- **Suggestion History**: Access to previously generated AI suggestions
- **Content Analysis**: Sentiment, readability, and quality scoring

### 🔗 n8n Agent Integration
- **Workflow Automation**: Create automated editorial tasks
- **Scheduled Processing**: Time-based content generation and moderation
- **Webhook Support**: Real-time status updates for external systems
- **Batch Operations**: Bulk content processing capabilities

## Authentication

All AI endpoints require JWT authentication:

```http
Authorization: Bearer <your-jwt-token>
```

## Rate Limiting

- **AI Endpoints**: 3 requests per second
- **Agent API**: 5 requests per second

## AI Endpoints

### Generate Headlines

Generate AI-powered headlines for articles.

**Endpoint:** `POST /api/ai/headlines`

**Request Body:**
```json
{
  "content": "Article content or summary",
  "count": 3,
  "style": "news" // options: news, casual, formal, clickbait
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "headlines": [
      "AI-Generated Headline 1",
      "AI-Generated Headline 2", 
      "AI-Generated Headline 3"
    ],
    "confidence": 0.95,
    "suggestion_id": 123
  }
}
```

### Generate Content

Create AI-assisted article content.

**Endpoint:** `POST /api/ai/content`

**Request Body:**
```json
{
  "topic": "Climate Change Impact",
  "content_type": "article", // options: article, summary, introduction, conclusion
  "length": "medium", // options: short, medium, long
  "tone": "informative", // options: informative, casual, formal, persuasive
  "keywords": ["climate", "environment", "sustainability"]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "content": "Generated article content...",
    "word_count": 500,
    "confidence": 0.92,
    "suggestion_id": 124
  }
}
```

### Improve Content

Get AI suggestions for improving existing content.

**Endpoint:** `POST /api/ai/improve`

**Request Body:**
```json
{
  "content": "Original article content",
  "focus": "clarity", // options: clarity, engagement, readability, seo
  "target_audience": "general" // options: general, technical, casual
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "suggestions": [
      {
        "type": "grammar",
        "original": "Original text",
        "improved": "Improved text",
        "reason": "Improved clarity and flow"
      }
    ],
    "overall_score": 0.88,
    "suggestion_id": 125
  }
}
```

### Moderate Content

AI-powered content moderation and safety checks.

**Endpoint:** `POST /api/ai/moderate`

**Request Body:**
```json
{
  "content": "Content to moderate",
  "content_type": "article", // options: article, comment, headline
  "strictness": "medium" // options: low, medium, high, strict
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "is_approved": true,
    "confidence": 0.96,
    "reason": "Content meets community guidelines",
    "categories": ["safe"],
    "severity": "low",
    "suggestions": ["Consider adding sources for claims"]
  }
}
```

### Summarize Content

Generate concise summaries of long-form content.

**Endpoint:** `POST /api/ai/summarize`

**Request Body:**
```json
{
  "content": "Long article content to summarize...",
  "length": "short", // options: short, medium, long
  "focus": "key_points" // options: key_points, overview, conclusion
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "summary": "Concise summary of the content...",
    "key_points": ["Point 1", "Point 2", "Point 3"],
    "original_word_count": 1500,
    "summary_word_count": 150,
    "suggestion_id": 126
  }
}
```

### Categorize Content

Automatically categorize and tag content.

**Endpoint:** `POST /api/ai/categorize`

**Request Body:**
```json
{
  "content": "Article content to categorize",
  "categories": ["technology", "science", "business", "sports"]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "primary_category": "technology",
    "confidence": 0.94,
    "all_categories": [
      {"name": "technology", "confidence": 0.94},
      {"name": "science", "confidence": 0.76}
    ],
    "suggested_tags": ["AI", "innovation", "automation"]
  }
}
```

### Get AI Suggestions History

Retrieve user's AI suggestion history.

**Endpoint:** `GET /api/ai/suggestions`

**Query Parameters:**
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20)
- `type`: Filter by suggestion type (optional)

**Response:**
```json
{
  "success": true,
  "data": {
    "suggestions": [
      {
        "id": 123,
        "type": "headline",
        "input": "Original content",
        "suggestion": "AI suggestion",
        "confidence": 0.95,
        "created_at": "2025-05-27T10:30:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 100,
      "pages": 5
    }
  }
}
```

### Get AI Usage Statistics

View AI service usage and consumption.

**Endpoint:** `GET /api/ai/usage-stats`

**Query Parameters:**
- `period`: Time period (day, week, month) (default: week)
- `service_type`: Filter by service type (optional)

**Response:**
```json
{
  "success": true,
  "data": {
    "total_requests": 150,
    "total_tokens": 25000,
    "by_service": {
      "headline_generation": {"requests": 50, "tokens": 5000},
      "content_generation": {"requests": 30, "tokens": 15000},
      "moderation": {"requests": 70, "tokens": 5000}
    },
    "period": "week",
    "start_date": "2025-05-20",
    "end_date": "2025-05-27"
  }
}
```

## Agent API (n8n Integration)

The Agent API enables workflow automation and integration with n8n for scheduled and batch operations.

### Create Agent Task

Create an automated task for n8n workflow integration.

**Endpoint:** `POST /api/agent/tasks`

**Request Body:**
```json
{
  "task_type": "content_generation", // See task types below
  "priority": "medium", // options: low, medium, high, urgent
  "parameters": {
    "topic": "AI in Journalism",
    "length": "long",
    "tone": "professional"
  },
  "scheduled_at": "2025-05-27T15:00:00Z" // optional
}
```

**Task Types:**
- `content_generation`: Generate new articles
- `content_moderation`: Batch moderate content
- `content_analysis`: Analyze content quality
- `scheduled_summary`: Daily/weekly summaries
- `bulk_categorization`: Categorize multiple articles

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 456,
    "task_type": "content_generation",
    "status": "pending",
    "priority": "medium",
    "webhook_url": "https://your-n8n-instance.com/webhook/task-456",
    "created_at": "2025-05-27T10:30:00Z",
    "scheduled_at": "2025-05-27T15:00:00Z"
  }
}
```

### List Agent Tasks

Get user's agent tasks with pagination.

**Endpoint:** `GET /api/agent/tasks`

**Query Parameters:**
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 20)
- `status`: Filter by status (pending, running, completed, failed)
- `task_type`: Filter by task type

**Response:**
```json
{
  "success": true,
  "data": {
    "tasks": [
      {
        "id": 456,
        "task_type": "content_generation",
        "status": "completed",
        "priority": "medium",
        "progress": 100,
        "created_at": "2025-05-27T10:30:00Z",
        "completed_at": "2025-05-27T10:45:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "pages": 3
    }
  }
}
```

### Get Agent Task

Retrieve specific agent task details.

**Endpoint:** `GET /api/agent/tasks/{id}`

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 456,
    "task_type": "content_generation",
    "status": "completed",
    "priority": "medium",
    "parameters": {
      "topic": "AI in Journalism",
      "length": "long"
    },
    "result": {
      "content": "Generated article content...",
      "word_count": 1200
    },
    "progress": 100,
    "created_at": "2025-05-27T10:30:00Z",
    "completed_at": "2025-05-27T10:45:00Z"
  }
}
```

### Update Agent Task

Update task status (typically called by n8n webhooks).

**Endpoint:** `PUT /api/agent/tasks/{id}`

**Request Body:**
```json
{
  "status": "running",
  "progress": 50,
  "result": {} // optional partial results
}
```

### Delete Agent Task

Remove an agent task.

**Endpoint:** `DELETE /api/agent/tasks/{id}`

### Process Agent Task

Manually trigger task processing.

**Endpoint:** `POST /api/agent/tasks/{id}/process`

## Configuration

Add these environment variables to your `.env` file:

```env
# OpenAI Configuration
OPENAI_API_KEY=your-openai-api-key-here
OPENAI_MODEL=gpt-3.5-turbo
OPENAI_MAX_TOKENS=1000
OPENAI_TEMPERATURE=0.7
```

## Error Handling

All endpoints return consistent error responses:

```json
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Content cannot be empty",
    "details": {}
  }
}
```

Common error codes:
- `INVALID_REQUEST`: Bad request data
- `UNAUTHORIZED`: Invalid authentication
- `RATE_LIMITED`: Too many requests
- `AI_SERVICE_ERROR`: OpenAI API error
- `INSUFFICIENT_CREDITS`: Usage limit exceeded

## Integration Examples

### n8n Workflow Example

1. **Trigger**: Schedule or webhook
2. **HTTP Request**: POST to `/api/agent/tasks`
3. **Wait**: Monitor webhook for completion
4. **Process**: Use generated content in your workflow

### Client Integration Example

```javascript
// Generate headlines for an article
const response = await fetch('/api/ai/headlines', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`
  },
  body: JSON.stringify({
    content: articleContent,
    count: 3,
    style: 'news'
  })
});

const { data } = await response.json();
console.log('Generated headlines:', data.headlines);
```

## Best Practices

1. **Content Quality**: Provide clear, well-structured input for better AI results
2. **Rate Limiting**: Implement client-side rate limiting to avoid API limits
3. **Error Handling**: Always handle API errors gracefully
4. **Token Management**: Monitor token usage to control costs
5. **Caching**: Cache AI results when appropriate to reduce API calls
6. **User Feedback**: Allow users to rate AI suggestions for continuous improvement

## Support

For technical support or feature requests related to AI integration, please refer to the main API documentation or contact the development team.
