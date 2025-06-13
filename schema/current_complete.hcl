table "atlas_schema_revisions" {
  schema = schema.atlas_schema_revisions
  column "version" {
    null = false
    type = character_varying
  }
  column "description" {
    null = false
    type = character_varying
  }
  column "type" {
    null    = false
    type    = bigint
    default = 2
  }
  column "applied" {
    null    = false
    type    = bigint
    default = 0
  }
  column "total" {
    null    = false
    type    = bigint
    default = 0
  }
  column "executed_at" {
    null = false
    type = timestamptz
  }
  column "execution_time" {
    null = false
    type = bigint
  }
  column "error" {
    null = true
    type = text
  }
  column "error_stmt" {
    null = true
    type = text
  }
  column "hash" {
    null = false
    type = character_varying
  }
  column "partial_hashes" {
    null = true
    type = jsonb
  }
  column "operator_version" {
    null = false
    type = character_varying
  }
  primary_key {
    columns = [column.version]
  }
}
table "agent_tasks" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "task_type" {
    null = false
    type = character_varying(50)
  }
  column "status" {
    null    = false
    type    = character_varying(20)
    default = "pending"
  }
  column "priority" {
    null    = true
    type    = bigint
    default = 0
  }
  column "input_data" {
    null = false
    type = jsonb
  }
  column "output_data" {
    null = true
    type = jsonb
  }
  column "error_msg" {
    null = true
    type = text
  }
  column "webhook_url" {
    null = true
    type = character_varying(500)
  }
  column "retry_count" {
    null    = true
    type    = bigint
    default = 0
  }
  column "max_retries" {
    null    = true
    type    = bigint
    default = 3
  }
  column "requested_by" {
    null = true
    type = bigint
  }
  column "started_at" {
    null = true
    type = timestamptz
  }
  column "completed_at" {
    null = true
    type = timestamptz
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_agent_tasks_user" {
    columns     = [column.requested_by]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_agent_tasks_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_agent_tasks_requested_by" {
    columns = [column.requested_by]
  }
}
table "ai_usage_stats" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "service_type" {
    null = false
    type = character_varying(50)
  }
  column "request_count" {
    null    = true
    type    = bigint
    default = 0
  }
  column "tokens_used" {
    null    = true
    type    = bigint
    default = 0
  }
  column "date" {
    null = false
    type = date
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_ai_usage_stats_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_ai_usage_stats_date" {
    columns = [column.date]
  }
  index "idx_ai_usage_stats_user_id" {
    columns = [column.user_id]
  }
}
table "article_categories" {
  schema = schema.public
  column "category_id" {
    null = false
    type = bigint
  }
  column "article_id" {
    null = false
    type = bigint
  }
  primary_key {
    columns = [column.category_id, column.article_id]
  }
  foreign_key "fk_article_categories_article" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_article_categories_category" {
    columns     = [column.category_id]
    ref_columns = [table.categories.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "article_content_blocks" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "article_id" {
    null = false
    type = bigint
  }
  column "block_type" {
    null = false
    type = character_varying(50)
  }
  column "content" {
    null = true
    type = text
  }
  column "settings" {
    null = true
    type = jsonb
  }
  column "position" {
    null = false
    type = bigint
  }
  column "is_visible" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_articles_content_blocks" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_article_content_blocks_article_id" {
    columns = [column.article_id]
  }
  index "idx_article_content_blocks_block_type" {
    columns = [column.block_type]
  }
  index "idx_article_content_blocks_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_article_content_blocks_position" {
    columns = [column.position]
  }
}
table "article_tags" {
  schema = schema.public
  column "tag_id" {
    null = false
    type = bigint
  }
  column "article_id" {
    null = false
    type = bigint
  }
  primary_key {
    columns = [column.tag_id, column.article_id]
  }
  foreign_key "fk_article_tags_article" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_article_tags_tag" {
    columns     = [column.tag_id]
    ref_columns = [table.tags.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "article_translations" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "article_id" {
    null = false
    type = bigint
  }
  column "language" {
    null = false
    type = character_varying(5)
  }
  column "title" {
    null = false
    type = character_varying(255)
  }
  column "slug" {
    null = false
    type = character_varying(255)
  }
  column "summary" {
    null = true
    type = text
  }
  column "content" {
    null = false
    type = text
  }
  column "meta_title" {
    null = true
    type = character_varying(255)
  }
  column "meta_description" {
    null = true
    type = character_varying(255)
  }
  column "translation_status" {
    null    = false
    type    = character_varying(20)
    default = "draft"
  }
  column "translator_id" {
    null = true
    type = bigint
  }
  column "translation_source" {
    null    = true
    type    = character_varying(20)
    default = "manual"
  }
  column "quality_score" {
    null = true
    type = numeric(3,2)
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_article_translations_translator" {
    columns     = [column.translator_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_articles_translations" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_article_translations_article_id" {
    columns = [column.article_id]
  }
  index "idx_article_translations_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_article_translations_language" {
    columns = [column.language]
  }
  index "idx_article_translations_translated_by" {
    columns = [column.translator_id]
  }
}
table "articles" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "title" {
    null = false
    type = character_varying(255)
  }
  column "slug" {
    null = false
    type = character_varying(255)
  }
  column "summary" {
    null = true
    type = text
  }
  column "content" {
    null = false
    type = text
  }
  column "content_type" {
    null    = false
    type    = character_varying(20)
    default = "legacy"
  }
  column "has_blocks" {
    null    = true
    type    = boolean
    default = false
  }
  column "blocks_version" {
    null    = true
    type    = bigint
    default = 1
  }
  column "author_id" {
    null = false
    type = bigint
  }
  column "featured_image" {
    null = true
    type = character_varying(255)
  }
  column "gallery" {
    null = true
    type = jsonb
  }
  column "status" {
    null    = false
    type    = character_varying(20)
    default = "draft"
  }
  column "published_at" {
    null = true
    type = timestamptz
  }
  column "scheduled_at" {
    null = true
    type = timestamptz
  }
  column "views" {
    null    = true
    type    = bigint
    default = 0
  }
  column "read_time" {
    null    = true
    type    = bigint
    default = 0
  }
  column "is_breaking" {
    null    = true
    type    = boolean
    default = false
  }
  column "is_featured" {
    null    = true
    type    = boolean
    default = false
  }
  column "is_sticky" {
    null    = true
    type    = boolean
    default = false
  }
  column "allow_comments" {
    null    = true
    type    = boolean
    default = true
  }
  column "meta_title" {
    null = true
    type = character_varying(255)
  }
  column "meta_description" {
    null = true
    type = character_varying(255)
  }
  column "source" {
    null = true
    type = character_varying(255)
  }
  column "source_url" {
    null = true
    type = character_varying(255)
  }
  column "language" {
    null    = true
    type    = character_varying(5)
    default = "tr"
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_users_articles" {
    columns     = [column.author_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_articles_author_id" {
    columns = [column.author_id]
  }
  index "idx_articles_created_at" {
    columns = [column.created_at]
  }
  index "idx_articles_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_articles_is_breaking" {
    columns = [column.is_breaking]
  }
  index "idx_articles_is_featured" {
    columns = [column.is_featured]
  }
  index "idx_articles_language" {
    columns = [column.language]
  }
  index "idx_articles_published_at" {
    columns = [column.published_at]
  }
  index "idx_articles_status" {
    columns = [column.status]
  }
  index "idx_articles_views" {
    columns = [column.views]
  }
  unique "uni_articles_slug" {
    columns = [column.slug]
  }
}
table "bookmarks" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "article_id" {
    null = false
    type = bigint
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_articles_bookmarks" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_users_bookmarks" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_bookmarks_article_id" {
    columns = [column.article_id]
  }
  index "idx_bookmarks_user_id" {
    columns = [column.user_id]
  }
}
table "breaking_news_banners" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "title" {
    null = false
    type = character_varying(255)
  }
  column "content" {
    null = true
    type = text
  }
  column "article_id" {
    null = true
    type = bigint
  }
  column "priority" {
    null    = false
    type    = bigint
    default = 1
  }
  column "style" {
    null    = true
    type    = character_varying(50)
    default = "urgent"
  }
  column "text_color" {
    null    = true
    type    = character_varying(7)
    default = "#FFFFFF"
  }
  column "background_color" {
    null    = true
    type    = character_varying(7)
    default = "#DC2626"
  }
  column "start_time" {
    null    = false
    type    = timestamptz
    default = sql("CURRENT_TIMESTAMP")
  }
  column "end_time" {
    null = true
    type = timestamptz
  }
  column "is_active" {
    null    = false
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_breaking_news_banners_article" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_breaking_news_banners_article_id" {
    columns = [column.article_id]
  }
  index "idx_breaking_news_banners_deleted_at" {
    columns = [column.deleted_at]
  }
}
table "categories" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "name" {
    null = false
    type = character_varying(100)
  }
  column "slug" {
    null = false
    type = character_varying(100)
  }
  column "description" {
    null = true
    type = text
  }
  column "color" {
    null = true
    type = character_varying(7)
  }
  column "icon" {
    null = true
    type = character_varying(50)
  }
  column "parent_id" {
    null = true
    type = bigint
  }
  column "sort_order" {
    null    = true
    type    = bigint
    default = 0
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_categories_children" {
    columns     = [column.parent_id]
    ref_columns = [table.categories.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_categories_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_categories_parent_id" {
    columns = [column.parent_id]
  }
  unique "uni_categories_name" {
    columns = [column.name]
  }
  unique "uni_categories_slug" {
    columns = [column.slug]
  }
}
table "category_translations" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "category_id" {
    null = false
    type = bigint
  }
  column "language" {
    null = false
    type = character_varying(5)
  }
  column "name" {
    null = false
    type = character_varying(100)
  }
  column "slug" {
    null = false
    type = character_varying(100)
  }
  column "description" {
    null = true
    type = text
  }
  column "meta_title" {
    null = true
    type = character_varying(255)
  }
  column "meta_desc" {
    null = true
    type = character_varying(255)
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_category_translations_category" {
    columns     = [column.category_id]
    ref_columns = [table.categories.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_category_translations_category_id" {
    columns = [column.category_id]
  }
  index "idx_category_translations_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_category_translations_language" {
    columns = [column.language]
  }
}
table "comments" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "article_id" {
    null = false
    type = bigint
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "parent_id" {
    null = true
    type = bigint
  }
  column "content" {
    null = false
    type = text
  }
  column "status" {
    null    = false
    type    = character_varying(20)
    default = "approved"
  }
  column "is_edited" {
    null    = true
    type    = boolean
    default = false
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_articles_comments" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_comments_replies" {
    columns     = [column.parent_id]
    ref_columns = [table.comments.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_users_comments" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_comments_article_id" {
    columns = [column.article_id]
  }
  index "idx_comments_created_at" {
    columns = [column.created_at]
  }
  index "idx_comments_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_comments_parent_id" {
    columns = [column.parent_id]
  }
  index "idx_comments_status" {
    columns = [column.status]
  }
  index "idx_comments_user_id" {
    columns = [column.user_id]
  }
}
table "content_analyses" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "article_id" {
    null = false
    type = bigint
  }
  column "reading_level" {
    null = true
    type = character_varying(20)
  }
  column "sentiment" {
    null = true
    type = character_varying(20)
  }
  column "keywords" {
    null = true
    type = json
  }
  column "categories" {
    null = true
    type = json
  }
  column "tags" {
    null = true
    type = json
  }
  column "summary" {
    null = true
    type = text
  }
  column "read_time" {
    null    = true
    type    = bigint
    default = 0
  }
  column "quality" {
    null = true
    type = numeric(3,2)
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_content_analyses_article" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_content_analyses_article_id" {
    columns = [column.article_id]
  }
}
table "content_suggestions" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "type" {
    null = false
    type = character_varying(50)
  }
  column "input" {
    null = false
    type = text
  }
  column "suggestion" {
    null = false
    type = text
  }
  column "context" {
    null = true
    type = json
  }
  column "confidence" {
    null = true
    type = numeric(3,2)
  }
  column "user_id" {
    null = true
    type = bigint
  }
  column "article_id" {
    null = true
    type = bigint
  }
  column "used" {
    null    = true
    type    = boolean
    default = false
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_content_suggestions_article" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_content_suggestions_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_content_suggestions_article_id" {
    columns = [column.article_id]
  }
  index "idx_content_suggestions_user_id" {
    columns = [column.user_id]
  }
}
table "follows" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "follower_id" {
    null = false
    type = bigint
  }
  column "following_id" {
    null = false
    type = bigint
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_users_followers" {
    columns     = [column.following_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_users_following" {
    columns     = [column.follower_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_follows_follower_id" {
    columns = [column.follower_id]
  }
  index "idx_follows_following_id" {
    columns = [column.following_id]
  }
}
table "live_news_streams" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "title" {
    null = false
    type = character_varying(255)
  }
  column "description" {
    null = true
    type = text
  }
  column "status" {
    null    = false
    type    = character_varying(20)
    default = "scheduled"
  }
  column "start_time" {
    null = true
    type = timestamptz
  }
  column "end_time" {
    null = true
    type = timestamptz
  }
  column "is_highlighted" {
    null    = true
    type    = boolean
    default = false
  }
  column "viewer_count" {
    null    = true
    type    = bigint
    default = 0
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_live_news_streams_deleted_at" {
    columns = [column.deleted_at]
  }
}
table "live_news_updates" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "stream_id" {
    null = false
    type = bigint
  }
  column "title" {
    null = false
    type = character_varying(255)
  }
  column "content" {
    null = false
    type = text
  }
  column "update_type" {
    null    = true
    type    = character_varying(50)
    default = "update"
  }
  column "importance" {
    null    = true
    type    = character_varying(20)
    default = "normal"
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_live_news_streams_updates" {
    columns     = [column.stream_id]
    ref_columns = [table.live_news_streams.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_live_news_updates_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_live_news_updates_stream_id" {
    columns = [column.stream_id]
  }
}
table "login_attempts" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = true
    type = bigint
  }
  column "username" {
    null = false
    type = character_varying(50)
  }
  column "ip" {
    null = true
    type = character_varying(50)
  }
  column "user_agent" {
    null = true
    type = character_varying(255)
  }
  column "location" {
    null = true
    type = character_varying(100)
  }
  column "timestamp" {
    null = true
    type = timestamptz
  }
  column "success" {
    null    = true
    type    = boolean
    default = false
  }
  column "failure_reason" {
    null = true
    type = character_varying(255)
  }
  primary_key {
    columns = [column.id]
  }
}
table "media" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "file_name" {
    null = false
    type = character_varying(255)
  }
  column "original_name" {
    null = false
    type = character_varying(255)
  }
  column "mime_type" {
    null = false
    type = character_varying(100)
  }
  column "size" {
    null = false
    type = bigint
  }
  column "path" {
    null = false
    type = character_varying(500)
  }
  column "url" {
    null = false
    type = character_varying(500)
  }
  column "alt_text" {
    null = true
    type = character_varying(255)
  }
  column "caption" {
    null = true
    type = text
  }
  column "uploaded_by" {
    null = false
    type = bigint
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_media_uploader" {
    columns     = [column.uploaded_by]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_media_deleted_at" {
    columns = [column.deleted_at]
  }
}
table "menu_item_translations" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "menu_item_id" {
    null = false
    type = bigint
  }
  column "language" {
    null = false
    type = character_varying(5)
  }
  column "title" {
    null = false
    type = character_varying(100)
  }
  column "url" {
    null = true
    type = character_varying(255)
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_menu_item_translations_menu_item" {
    columns     = [column.menu_item_id]
    ref_columns = [table.menu_items.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_menu_item_translations_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_menu_item_translations_language" {
    columns = [column.language]
  }
  index "idx_menu_item_translations_menu_item_id" {
    columns = [column.menu_item_id]
  }
}
table "menu_items" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "menu_id" {
    null = false
    type = bigint
  }
  column "parent_id" {
    null = true
    type = bigint
  }
  column "title" {
    null = false
    type = character_varying(100)
  }
  column "url" {
    null = true
    type = character_varying(255)
  }
  column "category_id" {
    null = true
    type = bigint
  }
  column "icon" {
    null = true
    type = character_varying(50)
  }
  column "target" {
    null    = true
    type    = character_varying(20)
    default = "_self"
  }
  column "sort_order" {
    null    = true
    type    = bigint
    default = 0
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_menu_items_category" {
    columns     = [column.category_id]
    ref_columns = [table.categories.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_menu_items_children" {
    columns     = [column.parent_id]
    ref_columns = [table.menu_items.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_menus_items" {
    columns     = [column.menu_id]
    ref_columns = [table.menus.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_menu_items_category_id" {
    columns = [column.category_id]
  }
  index "idx_menu_items_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_menu_items_menu_id" {
    columns = [column.menu_id]
  }
  index "idx_menu_items_parent_id" {
    columns = [column.parent_id]
  }
}
table "menu_translations" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "menu_id" {
    null = false
    type = bigint
  }
  column "language" {
    null = false
    type = character_varying(5)
  }
  column "name" {
    null = false
    type = character_varying(100)
  }
  column "description" {
    null = true
    type = text
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_menu_translations_menu" {
    columns     = [column.menu_id]
    ref_columns = [table.menus.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_menu_translations_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_menu_translations_language" {
    columns = [column.language]
  }
  index "idx_menu_translations_menu_id" {
    columns = [column.menu_id]
  }
}
table "menus" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "name" {
    null = false
    type = character_varying(100)
  }
  column "slug" {
    null = false
    type = character_varying(100)
  }
  column "location" {
    null = false
    type = character_varying(50)
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_menus_deleted_at" {
    columns = [column.deleted_at]
  }
  unique "uni_menus_slug" {
    columns = [column.slug]
  }
}
table "moderation_results" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "content_type" {
    null = false
    type = character_varying(20)
  }
  column "content_id" {
    null = false
    type = bigint
  }
  column "content" {
    null = false
    type = text
  }
  column "is_approved" {
    null    = true
    type    = boolean
    default = false
  }
  column "confidence" {
    null = true
    type = numeric(3,2)
  }
  column "reason" {
    null = true
    type = text
  }
  column "categories" {
    null = true
    type = json
  }
  column "severity" {
    null    = true
    type    = character_varying(20)
    default = "low"
  }
  column "reviewed_by" {
    null = true
    type = bigint
  }
  column "reviewed_at" {
    null = true
    type = timestamptz
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_moderation_results_reviewer" {
    columns     = [column.reviewed_by]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_moderation_results_reviewed_by" {
    columns = [column.reviewed_by]
  }
}
table "news_stories" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "headline" {
    null = false
    type = character_varying(255)
  }
  column "image_url" {
    null = false
    type = character_varying(255)
  }
  column "background_color" {
    null    = true
    type    = character_varying(7)
    default = "#000000"
  }
  column "text_color" {
    null    = true
    type    = character_varying(7)
    default = "#FFFFFF"
  }
  column "duration" {
    null    = true
    type    = bigint
    default = 5
  }
  column "article_id" {
    null = true
    type = bigint
  }
  column "external_url" {
    null = true
    type = character_varying(255)
  }
  column "sort_order" {
    null    = true
    type    = bigint
    default = 0
  }
  column "start_time" {
    null = false
    type = timestamptz
  }
  column "view_count" {
    null    = true
    type    = bigint
    default = 0
  }
  column "create_user_id" {
    null = false
    type = bigint
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_news_stories_article" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_news_stories_article_id" {
    columns = [column.article_id]
  }
  index "idx_news_stories_deleted_at" {
    columns = [column.deleted_at]
  }
}
table "newsletters" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "title" {
    null = false
    type = character_varying(255)
  }
  column "subject" {
    null = false
    type = character_varying(255)
  }
  column "content" {
    null = false
    type = text
  }
  column "status" {
    null    = false
    type    = character_varying(20)
    default = "draft"
  }
  column "sent_at" {
    null = true
    type = timestamptz
  }
  column "scheduled_at" {
    null = true
    type = timestamptz
  }
  column "created_by" {
    null = false
    type = bigint
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_newsletters_creator" {
    columns     = [column.created_by]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_newsletters_deleted_at" {
    columns = [column.deleted_at]
  }
}
table "notification_translations" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "notification_id" {
    null = false
    type = bigint
  }
  column "language" {
    null = false
    type = character_varying(5)
  }
  column "title" {
    null = false
    type = character_varying(255)
  }
  column "message" {
    null = false
    type = text
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_notification_translations_notification" {
    columns     = [column.notification_id]
    ref_columns = [table.notifications.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_notification_translations_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_notification_translations_language" {
    columns = [column.language]
  }
  index "idx_notification_translations_notification_id" {
    columns = [column.notification_id]
  }
}
table "notifications" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "type" {
    null = false
    type = character_varying(50)
  }
  column "title" {
    null = false
    type = character_varying(255)
  }
  column "message" {
    null = false
    type = text
  }
  column "data" {
    null = true
    type = json
  }
  column "is_read" {
    null    = true
    type    = boolean
    default = false
  }
  column "read_at" {
    null = true
    type = timestamptz
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_notifications_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_notifications_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_notifications_user_id" {
    columns = [column.user_id]
  }
}
table "page_content_blocks" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "page_id" {
    null = false
    type = bigint
  }
  column "container_id" {
    null = true
    type = bigint
  }
  column "block_type" {
    null = false
    type = character_varying(50)
  }
  column "content" {
    null = true
    type = text
  }
  column "settings" {
    null = true
    type = jsonb
  }
  column "styles" {
    null = true
    type = jsonb
  }
  column "position" {
    null = false
    type = bigint
  }
  column "is_visible" {
    null    = true
    type    = boolean
    default = true
  }
  column "is_container" {
    null    = true
    type    = boolean
    default = false
  }
  column "container_type" {
    null = true
    type = character_varying(30)
  }
  column "grid_settings" {
    null = true
    type = jsonb
  }
  column "responsive_data" {
    null = true
    type = jsonb
  }
  column "ai_generated" {
    null    = true
    type    = boolean
    default = false
  }
  column "performance_data" {
    null = true
    type = jsonb
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_page_content_blocks_child_blocks" {
    columns     = [column.container_id]
    ref_columns = [table.page_content_blocks.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_pages_content_blocks" {
    columns     = [column.page_id]
    ref_columns = [table.pages.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_page_content_blocks_block_type" {
    columns = [column.block_type]
  }
  index "idx_page_content_blocks_container_id" {
    columns = [column.container_id]
  }
  index "idx_page_content_blocks_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_page_content_blocks_page_id" {
    columns = [column.page_id]
  }
  index "idx_page_content_blocks_position" {
    columns = [column.position]
  }
}
table "page_templates" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "name" {
    null = false
    type = character_varying(100)
  }
  column "description" {
    null = true
    type = text
  }
  column "category" {
    null = true
    type = character_varying(50)
  }
  column "thumbnail" {
    null = true
    type = character_varying(255)
  }
  column "preview_image" {
    null = true
    type = character_varying(255)
  }
  column "block_structure" {
    null = false
    type = jsonb
  }
  column "default_styles" {
    null = true
    type = jsonb
  }
  column "is_public" {
    null    = true
    type    = boolean
    default = false
  }
  column "is_premium" {
    null    = true
    type    = boolean
    default = false
  }
  column "usage_count" {
    null    = true
    type    = bigint
    default = 0
  }
  column "rating" {
    null    = true
    type    = numeric
    default = 0
  }
  column "tags" {
    null = true
    type = jsonb
  }
  column "creator_id" {
    null = false
    type = bigint
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_page_templates_creator" {
    columns     = [column.creator_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_page_templates_creator_id" {
    columns = [column.creator_id]
  }
  index "idx_page_templates_deleted_at" {
    columns = [column.deleted_at]
  }
}
table "pages" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "title" {
    null = false
    type = character_varying(255)
  }
  column "slug" {
    null = false
    type = character_varying(255)
  }
  column "meta_title" {
    null = true
    type = character_varying(255)
  }
  column "meta_desc" {
    null = true
    type = character_varying(255)
  }
  column "template" {
    null    = true
    type    = character_varying(50)
    default = "default"
  }
  column "layout" {
    null    = true
    type    = character_varying(50)
    default = "container"
  }
  column "status" {
    null    = false
    type    = character_varying(20)
    default = "draft"
  }
  column "language" {
    null    = true
    type    = character_varying(5)
    default = "tr"
  }
  column "parent_id" {
    null = true
    type = bigint
  }
  column "sort_order" {
    null    = true
    type    = bigint
    default = 0
  }
  column "featured_image" {
    null = true
    type = character_varying(255)
  }
  column "excerpt_text" {
    null = true
    type = text
  }
  column "seo_settings" {
    null = true
    type = jsonb
  }
  column "page_settings" {
    null = true
    type = jsonb
  }
  column "layout_data" {
    null = true
    type = jsonb
  }
  column "author_id" {
    null = false
    type = bigint
  }
  column "published_at" {
    null = true
    type = timestamptz
  }
  column "scheduled_at" {
    null = true
    type = timestamptz
  }
  column "views" {
    null    = true
    type    = bigint
    default = 0
  }
  column "is_homepage" {
    null    = true
    type    = boolean
    default = false
  }
  column "is_landing_page" {
    null    = true
    type    = boolean
    default = false
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_pages_author" {
    columns     = [column.author_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_pages_children" {
    columns     = [column.parent_id]
    ref_columns = [table.pages.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_pages_author_id" {
    columns = [column.author_id]
  }
  index "idx_pages_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_pages_parent_id" {
    columns = [column.parent_id]
  }
  unique "uni_pages_slug" {
    columns = [column.slug]
  }
}
table "related_articles" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "article_id" {
    null = false
    type = bigint
  }
  column "related_id" {
    null = false
    type = bigint
  }
  column "score" {
    null    = true
    type    = bigint
    default = 0
  }
  column "type" {
    null    = true
    type    = character_varying(20)
    default = "auto"
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_articles_related" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_related_articles_related" {
    columns     = [column.related_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_related_articles_article_id" {
    columns = [column.article_id]
  }
  index "idx_related_articles_related_id" {
    columns = [column.related_id]
  }
}
table "security_events" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "event_type" {
    null = false
    type = character_varying(50)
  }
  column "description" {
    null = true
    type = character_varying(255)
  }
  column "ip" {
    null = true
    type = character_varying(50)
  }
  column "user_agent" {
    null = true
    type = character_varying(255)
  }
  column "metadata" {
    null = true
    type = text
  }
  column "timestamp" {
    null = true
    type = timestamptz
  }
  column "severity" {
    null    = true
    type    = character_varying(20)
    default = "info"
  }
  primary_key {
    columns = [column.id]
  }
}
table "setting_translations" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "setting_key" {
    null = false
    type = character_varying(100)
  }
  column "language" {
    null = false
    type = character_varying(5)
  }
  column "value" {
    null = false
    type = text
  }
  column "description" {
    null = true
    type = text
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_setting_translations_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_setting_translations_language" {
    columns = [column.language]
  }
  index "idx_setting_translations_setting_key" {
    columns = [column.setting_key]
  }
}
table "settings" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "key" {
    null = false
    type = character_varying(100)
  }
  column "value" {
    null = true
    type = text
  }
  column "type" {
    null    = false
    type    = character_varying(20)
    default = "string"
  }
  column "description" {
    null = true
    type = text
  }
  column "group" {
    null = true
    type = character_varying(50)
  }
  column "is_public" {
    null    = true
    type    = boolean
    default = false
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_settings_deleted_at" {
    columns = [column.deleted_at]
  }
  unique "uni_settings_key" {
    columns = [column.key]
  }
}
table "story_group_items" {
  schema = schema.public
  column "story_group_id" {
    null = false
    type = bigint
  }
  column "story_id" {
    null = false
    type = bigint
  }
  column "sort_order" {
    null    = true
    type    = bigint
    default = 0
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.story_group_id, column.story_id]
  }
  foreign_key "fk_story_group_items_story" {
    columns     = [column.story_id]
    ref_columns = [table.news_stories.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_story_group_items_story_group" {
    columns     = [column.story_group_id]
    ref_columns = [table.story_groups.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "story_groups" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "title" {
    null = false
    type = character_varying(100)
  }
  column "description" {
    null = true
    type = character_varying(255)
  }
  column "cover_image_url" {
    null = true
    type = character_varying(255)
  }
  column "sort_order" {
    null    = true
    type    = bigint
    default = 0
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_story_groups_deleted_at" {
    columns = [column.deleted_at]
  }
}
table "story_views" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "story_id" {
    null = false
    type = bigint
  }
  column "viewed_at" {
    null = false
    type = timestamptz
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_story_views_story" {
    columns     = [column.story_id]
    ref_columns = [table.news_stories.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_story_views_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_story_views_story_id" {
    columns = [column.story_id]
  }
  index "idx_story_views_user_id" {
    columns = [column.user_id]
  }
}
table "subscriptions" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = true
    type = bigint
  }
  column "email" {
    null = false
    type = character_varying(100)
  }
  column "type" {
    null = false
    type = character_varying(20)
  }
  column "category_id" {
    null = true
    type = bigint
  }
  column "tag_id" {
    null = true
    type = bigint
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "token" {
    null = true
    type = character_varying(255)
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_subscriptions_category" {
    columns     = [column.category_id]
    ref_columns = [table.categories.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_subscriptions_tag" {
    columns     = [column.tag_id]
    ref_columns = [table.tags.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_users_subscriptions" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_subscriptions_category_id" {
    columns = [column.category_id]
  }
  index "idx_subscriptions_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_subscriptions_tag_id" {
    columns = [column.tag_id]
  }
  index "idx_subscriptions_user_id" {
    columns = [column.user_id]
  }
  unique "uni_subscriptions_token" {
    columns = [column.token]
  }
}
table "tag_translations" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "tag_id" {
    null = false
    type = bigint
  }
  column "language" {
    null = false
    type = character_varying(5)
  }
  column "name" {
    null = false
    type = character_varying(50)
  }
  column "slug" {
    null = false
    type = character_varying(50)
  }
  column "description" {
    null = true
    type = text
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_tag_translations_tag" {
    columns     = [column.tag_id]
    ref_columns = [table.tags.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_tag_translations_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_tag_translations_language" {
    columns = [column.language]
  }
  index "idx_tag_translations_tag_id" {
    columns = [column.tag_id]
  }
}
table "tags" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "name" {
    null = false
    type = character_varying(50)
  }
  column "slug" {
    null = false
    type = character_varying(50)
  }
  column "description" {
    null = true
    type = text
  }
  column "color" {
    null = true
    type = character_varying(7)
  }
  column "usage_count" {
    null    = true
    type    = bigint
    default = 0
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_tags_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_tags_usage_count" {
    columns = [column.usage_count]
  }
  unique "uni_tags_name" {
    columns = [column.name]
  }
  unique "uni_tags_slug" {
    columns = [column.slug]
  }
}
table "translation_queue" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "entity_type" {
    null = false
    type = text
  }
  column "entity_id" {
    null = false
    type = bigint
  }
  column "source_language" {
    null = false
    type = character_varying(5)
  }
  column "target_language" {
    null = false
    type = character_varying(5)
  }
  column "status" {
    null    = true
    type    = text
    default = "pending"
  }
  column "priority" {
    null    = true
    type    = bigint
    default = 1
  }
  column "error_message" {
    null = true
    type = text
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
}
table "translations" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "key" {
    null = false
    type = character_varying(100)
  }
  column "language" {
    null = false
    type = character_varying(5)
  }
  column "value" {
    null = false
    type = text
  }
  column "category" {
    null = false
    type = character_varying(50)
  }
  column "is_active" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_translation_key_lang" {
    unique  = true
    columns = [column.key, column.language]
  }
  index "idx_translations_category" {
    columns = [column.category]
  }
  index "idx_translations_deleted_at" {
    columns = [column.deleted_at]
  }
}
table "user_article_interactions" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "article_id" {
    null = false
    type = bigint
  }
  column "interaction_type" {
    null = false
    type = character_varying(20)
  }
  column "duration" {
    null = true
    type = bigint
  }
  column "completion_rate" {
    null = true
    type = numeric(3,2)
  }
  column "platform" {
    null = true
    type = character_varying(20)
  }
  column "user_agent" {
    null = true
    type = character_varying(255)
  }
  column "ip_address" {
    null = true
    type = character_varying(45)
  }
  column "referrer_url" {
    null = true
    type = character_varying(500)
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_articles_user_interactions" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_users_article_interactions" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_user_article_interactions_article_id" {
    columns = [column.article_id]
  }
  index "idx_user_article_interactions_created_at" {
    columns = [column.created_at]
  }
  index "idx_user_article_interactions_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_user_article_interactions_interaction_type" {
    columns = [column.interaction_type]
  }
  index "idx_user_article_interactions_user_id" {
    columns = [column.user_id]
  }
}
table "user_sessions" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "token_id" {
    null = false
    type = character_varying(255)
  }
  column "ip" {
    null = true
    type = character_varying(50)
  }
  column "user_agent" {
    null = true
    type = character_varying(255)
  }
  column "device" {
    null = true
    type = character_varying(100)
  }
  column "location" {
    null = true
    type = character_varying(100)
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "expires_at" {
    null = true
    type = bigint
  }
  column "active" {
    null    = true
    type    = boolean
    default = true
  }
  column "revoked_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_user_sessions_deleted_at" {
    columns = [column.deleted_at]
  }
}
table "user_totps" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "secret" {
    null = false
    type = character_varying(255)
  }
  column "backup_codes" {
    null = true
    type = text
  }
  column "enabled" {
    null    = true
    type    = boolean
    default = false
  }
  column "activated_at" {
    null = true
    type = timestamptz
  }
  column "last_used_at" {
    null = true
    type = timestamptz
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_user_totps_user_id" {
    unique  = true
    columns = [column.user_id]
  }
}
table "users" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "username" {
    null = false
    type = character_varying(50)
  }
  column "email" {
    null = false
    type = character_varying(100)
  }
  column "password" {
    null = false
    type = character_varying(255)
  }
  column "first_name" {
    null = true
    type = character_varying(50)
  }
  column "last_name" {
    null = true
    type = character_varying(50)
  }
  column "avatar" {
    null = true
    type = character_varying(255)
  }
  column "bio" {
    null = true
    type = text
  }
  column "website" {
    null = true
    type = character_varying(255)
  }
  column "location" {
    null = true
    type = character_varying(100)
  }
  column "role" {
    null    = false
    type    = character_varying(20)
    default = "user"
  }
  column "status" {
    null    = false
    type    = character_varying(20)
    default = "active"
  }
  column "is_verified" {
    null    = true
    type    = boolean
    default = false
  }
  column "last_login_at" {
    null = true
    type = timestamptz
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  index "idx_users_created_at" {
    columns = [column.created_at]
  }
  index "idx_users_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_users_is_verified" {
    columns = [column.is_verified]
  }
  index "idx_users_last_login_at" {
    columns = [column.last_login_at]
  }
  index "idx_users_role" {
    columns = [column.role]
  }
  index "idx_users_status" {
    columns = [column.status]
  }
  unique "uni_users_email" {
    columns = [column.email]
  }
  unique "uni_users_username" {
    columns = [column.username]
  }
}
table "video_comment_votes" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "comment_id" {
    null = false
    type = bigint
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "type" {
    null = false
    type = character_varying(10)
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_video_comment_votes_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_video_comments_votes" {
    columns     = [column.comment_id]
    ref_columns = [table.video_comments.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_video_comment_votes_comment_id" {
    columns = [column.comment_id]
  }
  index "idx_video_comment_votes_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_video_comment_votes_user_id" {
    columns = [column.user_id]
  }
  check "chk_video_comment_votes_type" {
    expr = "((type)::text = ANY ((ARRAY['like'::character varying, 'dislike'::character varying])::text[]))"
  }
}
table "video_comments" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "video_id" {
    null = false
    type = bigint
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "content" {
    null = false
    type = text
  }
  column "parent_id" {
    null = true
    type = bigint
  }
  column "status" {
    null    = true
    type    = character_varying(20)
    default = "active"
  }
  column "is_edited" {
    null    = true
    type    = boolean
    default = false
  }
  column "edited_at" {
    null = true
    type = timestamptz
  }
  column "ai_moderated" {
    null    = true
    type    = boolean
    default = false
  }
  column "ai_confidence" {
    null    = true
    type    = numeric
    default = 0
  }
  column "toxicity_score" {
    null    = true
    type    = numeric
    default = 0
  }
  column "like_count" {
    null    = true
    type    = bigint
    default = 0
  }
  column "dislike_count" {
    null    = true
    type    = bigint
    default = 0
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_video_comments_replies" {
    columns     = [column.parent_id]
    ref_columns = [table.video_comments.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_video_comments_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_videos_comments" {
    columns     = [column.video_id]
    ref_columns = [table.videos.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_video_comments_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_video_comments_parent_id" {
    columns = [column.parent_id]
  }
  index "idx_video_comments_user_id" {
    columns = [column.user_id]
  }
  index "idx_video_comments_video_id" {
    columns = [column.video_id]
  }
  check "chk_video_comments_status" {
    expr = "((status)::text = ANY ((ARRAY['active'::character varying, 'hidden'::character varying, 'deleted'::character varying, 'flagged'::character varying])::text[]))"
  }
}
table "video_playlist_items" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "playlist_id" {
    null = false
    type = bigint
  }
  column "video_id" {
    null = false
    type = bigint
  }
  column "order" {
    null    = true
    type    = bigint
    default = 0
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_video_playlist_items_video" {
    columns     = [column.video_id]
    ref_columns = [table.videos.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_video_playlists_items" {
    columns     = [column.playlist_id]
    ref_columns = [table.video_playlists.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_video_playlist_items_playlist_id" {
    columns = [column.playlist_id]
  }
  index "idx_video_playlist_items_video_id" {
    columns = [column.video_id]
  }
}
table "video_playlists" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "name" {
    null = false
    type = character_varying(100)
  }
  column "description" {
    null = true
    type = text
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "is_public" {
    null    = true
    type    = boolean
    default = true
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_video_playlists_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_video_playlists_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_video_playlists_user_id" {
    columns = [column.user_id]
  }
}
table "video_processing_jobs" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "video_id" {
    null = false
    type = bigint
  }
  column "job_type" {
    null = false
    type = character_varying(50)
  }
  column "status" {
    null    = true
    type    = character_varying(20)
    default = "pending"
  }
  column "progress" {
    null    = true
    type    = bigint
    default = 0
  }
  column "error_msg" {
    null = true
    type = text
  }
  column "parameters" {
    null = true
    type = text
  }
  column "result" {
    null = true
    type = text
  }
  column "started_at" {
    null = true
    type = timestamptz
  }
  column "completed_at" {
    null = true
    type = timestamptz
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_video_processing_jobs_video" {
    columns     = [column.video_id]
    ref_columns = [table.videos.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_video_processing_jobs_video_id" {
    columns = [column.video_id]
  }
}
table "video_views" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "video_id" {
    null = false
    type = bigint
  }
  column "user_id" {
    null = true
    type = bigint
  }
  column "ip_address" {
    null = true
    type = character_varying(45)
  }
  column "user_agent" {
    null = true
    type = character_varying(500)
  }
  column "duration" {
    null = true
    type = bigint
  }
  column "watch_percent" {
    null = true
    type = numeric
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_video_views_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_videos_views" {
    columns     = [column.video_id]
    ref_columns = [table.videos.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_video_views_user_id" {
    columns = [column.user_id]
  }
  index "idx_video_views_video_id" {
    columns = [column.video_id]
  }
}
table "video_votes" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "video_id" {
    null = false
    type = bigint
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "type" {
    null = false
    type = character_varying(10)
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_video_votes_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_videos_votes" {
    columns     = [column.video_id]
    ref_columns = [table.videos.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_video_votes_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_video_votes_user_id" {
    columns = [column.user_id]
  }
  index "idx_video_votes_video_id" {
    columns = [column.video_id]
  }
  check "chk_video_votes_type" {
    expr = "((type)::text = ANY ((ARRAY['like'::character varying, 'dislike'::character varying])::text[]))"
  }
}
table "videos" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "title" {
    null = false
    type = character_varying(200)
  }
  column "description" {
    null = true
    type = text
  }
  column "video_url" {
    null = false
    type = character_varying(500)
  }
  column "thumbnail_url" {
    null = true
    type = character_varying(500)
  }
  column "duration" {
    null = true
    type = bigint
  }
  column "file_size" {
    null = true
    type = bigint
  }
  column "resolution" {
    null = true
    type = character_varying(20)
  }
  column "category_id" {
    null = true
    type = bigint
  }
  column "tags" {
    null = true
    type = text
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "is_generated" {
    null    = true
    type    = boolean
    default = false
  }
  column "status" {
    null    = true
    type    = character_varying(20)
    default = "pending"
  }
  column "is_public" {
    null    = true
    type    = boolean
    default = true
  }
  column "is_featured" {
    null    = true
    type    = boolean
    default = false
  }
  column "ai_generated" {
    null    = true
    type    = boolean
    default = false
  }
  column "ai_confidence" {
    null    = true
    type    = numeric
    default = 0
  }
  column "content_warning" {
    null = true
    type = character_varying(100)
  }
  column "view_count" {
    null    = true
    type    = bigint
    default = 0
  }
  column "like_count" {
    null    = true
    type    = bigint
    default = 0
  }
  column "dislike_count" {
    null    = true
    type    = bigint
    default = 0
  }
  column "comment_count" {
    null    = true
    type    = bigint
    default = 0
  }
  column "share_count" {
    null    = true
    type    = bigint
    default = 0
  }
  column "published_at" {
    null = true
    type = timestamptz
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  column "updated_at" {
    null = true
    type = timestamptz
  }
  column "deleted_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_videos_category" {
    columns     = [column.category_id]
    ref_columns = [table.categories.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_videos_user" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_videos_ai_generated" {
    columns = [column.ai_generated]
  }
  index "idx_videos_category_id" {
    columns = [column.category_id]
  }
  index "idx_videos_created_at" {
    columns = [column.created_at]
  }
  index "idx_videos_deleted_at" {
    columns = [column.deleted_at]
  }
  index "idx_videos_is_featured" {
    columns = [column.is_featured]
  }
  index "idx_videos_is_public" {
    columns = [column.is_public]
  }
  index "idx_videos_like_count" {
    columns = [column.like_count]
  }
  index "idx_videos_published_at" {
    columns = [column.published_at]
  }
  index "idx_videos_status" {
    columns = [column.status]
  }
  index "idx_videos_user_id" {
    columns = [column.user_id]
  }
  index "idx_videos_view_count" {
    columns = [column.view_count]
  }
  check "chk_videos_status" {
    expr = "((status)::text = ANY ((ARRAY['draft'::character varying, 'pending'::character varying, 'published'::character varying, 'rejected'::character varying, 'archived'::character varying])::text[]))"
  }
}
table "votes" {
  schema = schema.public
  column "id" {
    null = false
    type = bigserial
  }
  column "user_id" {
    null = false
    type = bigint
  }
  column "article_id" {
    null = true
    type = bigint
  }
  column "comment_id" {
    null = true
    type = bigint
  }
  column "type" {
    null = false
    type = character_varying(10)
  }
  column "created_at" {
    null = true
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "fk_articles_votes" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_comments_votes" {
    columns     = [column.comment_id]
    ref_columns = [table.comments.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "fk_users_votes" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "idx_votes_article_id" {
    columns = [column.article_id]
  }
  index "idx_votes_comment_id" {
    columns = [column.comment_id]
  }
  index "idx_votes_created_at" {
    columns = [column.created_at]
  }
  index "idx_votes_type" {
    columns = [column.type]
  }
  index "idx_votes_user_id" {
    columns = [column.user_id]
  }
}
schema "atlas_schema_revisions" {
}
schema "public" {
  comment = "standard public schema"
}
