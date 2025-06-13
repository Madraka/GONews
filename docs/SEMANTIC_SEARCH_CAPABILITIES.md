# Go News API - Semantic Search Capabilities Documentation

## Overview

The Go News API provides sophisticated semantic search capabilities powered by OpenAI embeddings and ElasticSearch, enabling users to search across multiple content types using natural language queries. The system supports both AI-powered semantic search and traditional text-based search with automatic fallback mechanisms.

## Searchable Content Types

### 1. Articles (`models.Article`)
Standard news articles with comprehensive metadata support.

**Key Fields:**
- `title` - Article headline
- `summary` - Brief article summary
- `content` - Full article content
- `category` - Article category (politics, sports, technology, etc.)
- `tags` - Associated keywords/tags
- `author` - Article author information
- `published_at` - Publication timestamp
- `updated_at` - Last modification timestamp

**Search Capabilities:**
- Full-text search across title, summary, and content
- Category-based filtering
- Tag-based filtering
- Author-based search
- Date range filtering

### 2. Breaking News Banners (`models.BreakingNewsBanner`)
Urgent news notifications displayed prominently.

**Key Fields:**
- `title` - Breaking news headline
- `content` - Breaking news content
- `priority` - Urgency level (1-10)
- `is_active` - Current display status
- `expires_at` - Expiration timestamp

**Search Capabilities:**
- Real-time urgent news discovery
- Priority-based filtering
- Active status filtering
- Expiration-aware search

### 3. Live News Streams (`models.LiveNewsStream`)
Real-time news feeds with continuous updates.

**Key Fields:**
- `title` - Stream title
- `description` - Stream description
- `status` - Stream status (active, inactive, scheduled)
- `scheduled_start` - Planned start time
- `actual_start` - Actual start time
- `stream_url` - Live stream URL

**Associated Updates (`models.LiveNewsUpdate`):**
- `content` - Update content
- `timestamp` - Update time
- `importance` - Update importance level

**Search Capabilities:**
- Live stream discovery
- Status-based filtering
- Time-based search
- Update content search

### 4. News Stories (`models.NewsStory`)
Instagram/Facebook-style ephemeral content.

**Key Fields:**
- `title` - Story title
- `content` - Story content
- `media_url` - Associated media
- `media_type` - Media format (image, video)
- `expires_at` - Story expiration
- `view_count` - Engagement metrics

**Search Capabilities:**
- Ephemeral content discovery
- Media type filtering
- Engagement-based ranking
- Expiration-aware search

### 5. Localized Articles (`models.LocalizedArticle`)
Multi-language content support.

**Key Fields:**
- `original_article_id` - Reference to source article
- `language` - Content language (EN, TR, ES)
- `title` - Localized title
- `summary` - Localized summary
- `content` - Localized content
- `translation_status` - Translation quality status

**Search Capabilities:**
- Multi-language search
- Translation status filtering
- Cross-language content discovery

## Search Infrastructure

### AI-Powered Semantic Search
- **Technology**: OpenAI text-embedding-ada-002 model
- **Vector Dimensions**: 1536
- **Storage**: ElasticSearch with vector similarity search
- **Similarity Algorithm**: Cosine similarity
- **Fallback**: Traditional text search when AI service unavailable

### Rate Limiting
- **Unauthenticated Users**: 5 AI searches per day
- **Authenticated Users**: 50 AI searches per day
- **Local Search**: Higher limits (100-1000 requests/day)

### Indexing Strategy
Articles are automatically indexed with:
- Combined text (title + summary + content)
- Metadata (category, tags, author)
- Timestamps (published_at, updated_at)
- Vector embeddings for semantic similarity

## API Endpoints

### 1. Semantic Search Endpoint
```
GET /api/v1/search
```

**Parameters:**
- `q` (required) - Search query
- `limit` (optional) - Number of results (default: 10, max: 50)
- `offset` (optional) - Pagination offset
- `content_type` (optional) - Filter by content type
- `category` (optional) - Filter by category
- `language` (optional) - Language preference

**Response:**
```json
{
  "results": [
    {
      "id": "uuid",
      "type": "article|breaking_news|live_stream|story",
      "title": "string",
      "summary": "string",
      "content": "string",
      "score": 0.95,
      "metadata": {
        "category": "string",
        "tags": ["string"],
        "author": "string",
        "published_at": "timestamp"
      }
    }
  ],
  "total": 100,
  "search_type": "semantic|text",
  "query_time_ms": 150
}
```

### 2. Localized Article Search
```
GET /api/v1/articles/search/localized
```

**Parameters:**
- `q` (required) - Search query
- `language` (required) - Target language (en, tr, es)
- `limit` (optional) - Number of results
- `category` (optional) - Category filter

### 3. Breaking News Search
```
GET /api/v1/breaking-news/search
```

**Parameters:**
- `q` (optional) - Search query
- `active_only` (optional) - Filter active banners only
- `priority_min` (optional) - Minimum priority level

### 4. Live News Search
```
GET /api/v1/live-news/search
```

**Parameters:**
- `q` (optional) - Search query
- `status` (optional) - Stream status filter
- `include_updates` (optional) - Include stream updates in search

### 5. News Stories Search
```
GET /api/v1/stories/search
```

**Parameters:**
- `q` (optional) - Search query
- `media_type` (optional) - Media type filter
- `active_only` (optional) - Non-expired stories only

## Search Examples

### Basic Semantic Search
```bash
curl "http://localhost:8080/api/v1/search?q=climate%20change%20policies"
```

### Category-Filtered Search
```bash
curl "http://localhost:8080/api/v1/search?q=election%20results&category=politics"
```

### Multi-Language Search
```bash
curl "http://localhost:8080/api/v1/articles/search/localized?q=teknoloji%20haberleri&language=tr"
```

### Breaking News Search
```bash
curl "http://localhost:8080/api/v1/breaking-news/search?active_only=true&priority_min=7"
```

### Live Stream Discovery
```bash
curl "http://localhost:8080/api/v1/live-news/search?status=active&include_updates=true"
```

## Technical Implementation Details

### Vector Embedding Process
1. Content preprocessing (title + summary + content)
2. OpenAI API call for embedding generation
3. ElasticSearch indexing with vector storage
4. Similarity search using cosine distance

### Search Flow
1. Query preprocessing and validation
2. Rate limit checking
3. AI embedding generation (if available)
4. ElasticSearch vector similarity search
5. Fallback to text search if needed
6. Result ranking and formatting
7. Response with metadata

### Performance Optimizations
- Embedding caching for common queries
- ElasticSearch index optimization
- Result pagination
- Asynchronous processing for bulk operations

## Error Handling

### Common Error Responses
- `400 Bad Request` - Invalid query parameters
- `429 Too Many Requests` - Rate limit exceeded
- `503 Service Unavailable` - AI service temporarily unavailable
- `500 Internal Server Error` - Search infrastructure issues

### Fallback Mechanisms
- Automatic fallback to text search when AI unavailable
- Cached results for repeated queries
- Graceful degradation with reduced functionality

## Monitoring and Analytics

### Available Metrics
- Search query volume and patterns
- AI vs. text search usage ratio
- Response times and performance
- Error rates and service health
- User engagement with search results

### Health Endpoints
- `/health` - Overall API health
- `/health/elasticsearch` - Search service health
- `/health/ai` - AI service connectivity

This documentation provides a comprehensive overview of the Go News API's semantic search capabilities, enabling developers to effectively utilize the search functionality across all available content types.
