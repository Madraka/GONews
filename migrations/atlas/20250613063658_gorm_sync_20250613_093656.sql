-- Set comment to schema: "public"
COMMENT ON SCHEMA "public" IS NULL;
-- Create "setting_translations" table
CREATE TABLE "public"."setting_translations" (
  "id" bigserial NOT NULL,
  "setting_key" character varying(100) NOT NULL,
  "language" character varying(5) NOT NULL,
  "value" text NOT NULL,
  "description" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_setting_translations_deleted_at" to table: "setting_translations"
CREATE INDEX "idx_setting_translations_deleted_at" ON "public"."setting_translations" ("deleted_at");
-- Create index "idx_setting_translations_language" to table: "setting_translations"
CREATE INDEX "idx_setting_translations_language" ON "public"."setting_translations" ("language");
-- Create index "idx_setting_translations_setting_key" to table: "setting_translations"
CREATE INDEX "idx_setting_translations_setting_key" ON "public"."setting_translations" ("setting_key");
-- Create "settings" table
CREATE TABLE "public"."settings" (
  "id" bigserial NOT NULL,
  "key" character varying(100) NOT NULL,
  "value" text NULL,
  "type" character varying(20) NOT NULL DEFAULT 'string',
  "description" text NULL,
  "group" character varying(50) NULL,
  "is_public" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_settings_key" UNIQUE ("key")
);
-- Create index "idx_settings_deleted_at" to table: "settings"
CREATE INDEX "idx_settings_deleted_at" ON "public"."settings" ("deleted_at");
-- Create "translation_queue" table
CREATE TABLE "public"."translation_queue" (
  "id" bigserial NOT NULL,
  "entity_type" text NOT NULL,
  "entity_id" bigint NOT NULL,
  "source_language" character varying(5) NOT NULL,
  "target_language" character varying(5) NOT NULL,
  "status" text NULL DEFAULT 'pending',
  "priority" bigint NULL DEFAULT 1,
  "error_message" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "translations" table
CREATE TABLE "public"."translations" (
  "id" bigserial NOT NULL,
  "key" character varying(100) NOT NULL,
  "language" character varying(5) NOT NULL,
  "value" text NOT NULL,
  "category" character varying(50) NOT NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_translation_key_lang" to table: "translations"
CREATE UNIQUE INDEX "idx_translation_key_lang" ON "public"."translations" ("key", "language");
-- Create index "idx_translations_category" to table: "translations"
CREATE INDEX "idx_translations_category" ON "public"."translations" ("category");
-- Create index "idx_translations_deleted_at" to table: "translations"
CREATE INDEX "idx_translations_deleted_at" ON "public"."translations" ("deleted_at");
-- Create "security_events" table
CREATE TABLE "public"."security_events" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "event_type" character varying(50) NOT NULL,
  "description" character varying(255) NULL,
  "ip" character varying(50) NULL,
  "user_agent" character varying(255) NULL,
  "metadata" text NULL,
  "timestamp" timestamptz NULL,
  "severity" character varying(20) NULL DEFAULT 'info',
  PRIMARY KEY ("id")
);
-- Create "user_sessions" table
CREATE TABLE "public"."user_sessions" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "token_id" character varying(255) NOT NULL,
  "ip" character varying(50) NULL,
  "user_agent" character varying(255) NULL,
  "device" character varying(100) NULL,
  "location" character varying(100) NULL,
  "created_at" timestamptz NULL,
  "expires_at" bigint NULL,
  "active" boolean NULL DEFAULT true,
  "revoked_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_user_sessions_deleted_at" to table: "user_sessions"
CREATE INDEX "idx_user_sessions_deleted_at" ON "public"."user_sessions" ("deleted_at");
-- Create "user_totps" table
CREATE TABLE "public"."user_totps" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "secret" character varying(255) NOT NULL,
  "backup_codes" text NULL,
  "enabled" boolean NULL DEFAULT false,
  "activated_at" timestamptz NULL,
  "last_used_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_user_totps_user_id" to table: "user_totps"
CREATE UNIQUE INDEX "idx_user_totps_user_id" ON "public"."user_totps" ("user_id");
-- Create "login_attempts" table
CREATE TABLE "public"."login_attempts" (
  "id" bigserial NOT NULL,
  "user_id" bigint NULL,
  "username" character varying(50) NOT NULL,
  "ip" character varying(50) NULL,
  "user_agent" character varying(255) NULL,
  "location" character varying(100) NULL,
  "timestamp" timestamptz NULL,
  "success" boolean NULL DEFAULT false,
  "failure_reason" character varying(255) NULL,
  PRIMARY KEY ("id")
);
-- Create "users" table
CREATE TABLE "public"."users" (
  "id" bigserial NOT NULL,
  "username" character varying(50) NOT NULL,
  "email" character varying(100) NOT NULL,
  "password" character varying(255) NOT NULL,
  "first_name" character varying(50) NULL,
  "last_name" character varying(50) NULL,
  "avatar" character varying(255) NULL,
  "bio" text NULL,
  "website" character varying(255) NULL,
  "location" character varying(100) NULL,
  "role" character varying(20) NOT NULL DEFAULT 'user',
  "status" character varying(20) NOT NULL DEFAULT 'active',
  "is_verified" boolean NULL DEFAULT false,
  "last_login_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_users_email" UNIQUE ("email"),
  CONSTRAINT "uni_users_username" UNIQUE ("username")
);
-- Create index "idx_users_created_at" to table: "users"
CREATE INDEX "idx_users_created_at" ON "public"."users" ("created_at");
-- Create index "idx_users_deleted_at" to table: "users"
CREATE INDEX "idx_users_deleted_at" ON "public"."users" ("deleted_at");
-- Create index "idx_users_is_verified" to table: "users"
CREATE INDEX "idx_users_is_verified" ON "public"."users" ("is_verified");
-- Create index "idx_users_last_login_at" to table: "users"
CREATE INDEX "idx_users_last_login_at" ON "public"."users" ("last_login_at");
-- Create index "idx_users_role" to table: "users"
CREATE INDEX "idx_users_role" ON "public"."users" ("role");
-- Create index "idx_users_status" to table: "users"
CREATE INDEX "idx_users_status" ON "public"."users" ("status");
-- Create "atlas_schema_revisions" table
CREATE TABLE "atlas_schema_revisions"."atlas_schema_revisions" (
  "version" character varying NOT NULL,
  "description" character varying NOT NULL,
  "type" bigint NOT NULL DEFAULT 2,
  "applied" bigint NOT NULL DEFAULT 0,
  "total" bigint NOT NULL DEFAULT 0,
  "executed_at" timestamptz NOT NULL,
  "execution_time" bigint NOT NULL,
  "error" text NULL,
  "error_stmt" text NULL,
  "hash" character varying NOT NULL,
  "partial_hashes" jsonb NULL,
  "operator_version" character varying NOT NULL,
  PRIMARY KEY ("version")
);
-- Create "agent_tasks" table
CREATE TABLE "public"."agent_tasks" (
  "id" bigserial NOT NULL,
  "task_type" character varying(50) NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "priority" bigint NULL DEFAULT 0,
  "input_data" jsonb NOT NULL,
  "output_data" jsonb NULL,
  "error_msg" text NULL,
  "webhook_url" character varying(500) NULL,
  "retry_count" bigint NULL DEFAULT 0,
  "max_retries" bigint NULL DEFAULT 3,
  "requested_by" bigint NULL,
  "started_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_agent_tasks_user" FOREIGN KEY ("requested_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_agent_tasks_deleted_at" to table: "agent_tasks"
CREATE INDEX "idx_agent_tasks_deleted_at" ON "public"."agent_tasks" ("deleted_at");
-- Create index "idx_agent_tasks_requested_by" to table: "agent_tasks"
CREATE INDEX "idx_agent_tasks_requested_by" ON "public"."agent_tasks" ("requested_by");
-- Create "ai_usage_stats" table
CREATE TABLE "public"."ai_usage_stats" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "service_type" character varying(50) NOT NULL,
  "request_count" bigint NULL DEFAULT 0,
  "tokens_used" bigint NULL DEFAULT 0,
  "date" date NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_ai_usage_stats_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_ai_usage_stats_date" to table: "ai_usage_stats"
CREATE INDEX "idx_ai_usage_stats_date" ON "public"."ai_usage_stats" ("date");
-- Create index "idx_ai_usage_stats_user_id" to table: "ai_usage_stats"
CREATE INDEX "idx_ai_usage_stats_user_id" ON "public"."ai_usage_stats" ("user_id");
-- Create "articles" table
CREATE TABLE "public"."articles" (
  "id" bigserial NOT NULL,
  "title" character varying(255) NOT NULL,
  "slug" character varying(255) NOT NULL,
  "summary" text NULL,
  "content" text NOT NULL,
  "content_type" character varying(20) NOT NULL DEFAULT 'legacy',
  "has_blocks" boolean NULL DEFAULT false,
  "blocks_version" bigint NULL DEFAULT 1,
  "author_id" bigint NOT NULL,
  "featured_image" character varying(255) NULL,
  "gallery" jsonb NULL,
  "status" character varying(20) NOT NULL DEFAULT 'draft',
  "published_at" timestamptz NULL,
  "scheduled_at" timestamptz NULL,
  "views" bigint NULL DEFAULT 0,
  "read_time" bigint NULL DEFAULT 0,
  "is_breaking" boolean NULL DEFAULT false,
  "is_featured" boolean NULL DEFAULT false,
  "is_sticky" boolean NULL DEFAULT false,
  "allow_comments" boolean NULL DEFAULT true,
  "meta_title" character varying(255) NULL,
  "meta_description" character varying(255) NULL,
  "source" character varying(255) NULL,
  "source_url" character varying(255) NULL,
  "language" character varying(5) NULL DEFAULT 'tr',
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_articles_slug" UNIQUE ("slug"),
  CONSTRAINT "fk_users_articles" FOREIGN KEY ("author_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_articles_author_id" to table: "articles"
CREATE INDEX "idx_articles_author_id" ON "public"."articles" ("author_id");
-- Create index "idx_articles_created_at" to table: "articles"
CREATE INDEX "idx_articles_created_at" ON "public"."articles" ("created_at");
-- Create index "idx_articles_deleted_at" to table: "articles"
CREATE INDEX "idx_articles_deleted_at" ON "public"."articles" ("deleted_at");
-- Create index "idx_articles_is_breaking" to table: "articles"
CREATE INDEX "idx_articles_is_breaking" ON "public"."articles" ("is_breaking");
-- Create index "idx_articles_is_featured" to table: "articles"
CREATE INDEX "idx_articles_is_featured" ON "public"."articles" ("is_featured");
-- Create index "idx_articles_language" to table: "articles"
CREATE INDEX "idx_articles_language" ON "public"."articles" ("language");
-- Create index "idx_articles_published_at" to table: "articles"
CREATE INDEX "idx_articles_published_at" ON "public"."articles" ("published_at");
-- Create index "idx_articles_status" to table: "articles"
CREATE INDEX "idx_articles_status" ON "public"."articles" ("status");
-- Create index "idx_articles_views" to table: "articles"
CREATE INDEX "idx_articles_views" ON "public"."articles" ("views");
-- Create "categories" table
CREATE TABLE "public"."categories" (
  "id" bigserial NOT NULL,
  "name" character varying(100) NOT NULL,
  "slug" character varying(100) NOT NULL,
  "description" text NULL,
  "color" character varying(7) NULL,
  "icon" character varying(50) NULL,
  "parent_id" bigint NULL,
  "sort_order" bigint NULL DEFAULT 0,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_categories_name" UNIQUE ("name"),
  CONSTRAINT "uni_categories_slug" UNIQUE ("slug"),
  CONSTRAINT "fk_categories_children" FOREIGN KEY ("parent_id") REFERENCES "public"."categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_categories_deleted_at" to table: "categories"
CREATE INDEX "idx_categories_deleted_at" ON "public"."categories" ("deleted_at");
-- Create index "idx_categories_parent_id" to table: "categories"
CREATE INDEX "idx_categories_parent_id" ON "public"."categories" ("parent_id");
-- Create "article_categories" table
CREATE TABLE "public"."article_categories" (
  "category_id" bigint NOT NULL,
  "article_id" bigint NOT NULL,
  PRIMARY KEY ("category_id", "article_id"),
  CONSTRAINT "fk_article_categories_article" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_article_categories_category" FOREIGN KEY ("category_id") REFERENCES "public"."categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "article_content_blocks" table
CREATE TABLE "public"."article_content_blocks" (
  "id" bigserial NOT NULL,
  "article_id" bigint NOT NULL,
  "block_type" character varying(50) NOT NULL,
  "content" text NULL,
  "settings" jsonb NULL,
  "position" bigint NOT NULL,
  "is_visible" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_articles_content_blocks" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_article_content_blocks_article_id" to table: "article_content_blocks"
CREATE INDEX "idx_article_content_blocks_article_id" ON "public"."article_content_blocks" ("article_id");
-- Create index "idx_article_content_blocks_block_type" to table: "article_content_blocks"
CREATE INDEX "idx_article_content_blocks_block_type" ON "public"."article_content_blocks" ("block_type");
-- Create index "idx_article_content_blocks_deleted_at" to table: "article_content_blocks"
CREATE INDEX "idx_article_content_blocks_deleted_at" ON "public"."article_content_blocks" ("deleted_at");
-- Create index "idx_article_content_blocks_position" to table: "article_content_blocks"
CREATE INDEX "idx_article_content_blocks_position" ON "public"."article_content_blocks" ("position");
-- Create "tags" table
CREATE TABLE "public"."tags" (
  "id" bigserial NOT NULL,
  "name" character varying(50) NOT NULL,
  "slug" character varying(50) NOT NULL,
  "description" text NULL,
  "color" character varying(7) NULL,
  "usage_count" bigint NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_tags_name" UNIQUE ("name"),
  CONSTRAINT "uni_tags_slug" UNIQUE ("slug")
);
-- Create index "idx_tags_deleted_at" to table: "tags"
CREATE INDEX "idx_tags_deleted_at" ON "public"."tags" ("deleted_at");
-- Create index "idx_tags_usage_count" to table: "tags"
CREATE INDEX "idx_tags_usage_count" ON "public"."tags" ("usage_count");
-- Create "article_tags" table
CREATE TABLE "public"."article_tags" (
  "tag_id" bigint NOT NULL,
  "article_id" bigint NOT NULL,
  PRIMARY KEY ("tag_id", "article_id"),
  CONSTRAINT "fk_article_tags_article" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_article_tags_tag" FOREIGN KEY ("tag_id") REFERENCES "public"."tags" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "article_translations" table
CREATE TABLE "public"."article_translations" (
  "id" bigserial NOT NULL,
  "article_id" bigint NOT NULL,
  "language" character varying(5) NOT NULL,
  "title" character varying(255) NOT NULL,
  "slug" character varying(255) NOT NULL,
  "summary" text NULL,
  "content" text NOT NULL,
  "meta_title" character varying(255) NULL,
  "meta_description" character varying(255) NULL,
  "translation_status" character varying(20) NOT NULL DEFAULT 'draft',
  "translator_id" bigint NULL,
  "translation_source" character varying(20) NULL DEFAULT 'manual',
  "quality_score" numeric(3,2) NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_article_translations_translator" FOREIGN KEY ("translator_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_articles_translations" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_article_translations_article_id" to table: "article_translations"
CREATE INDEX "idx_article_translations_article_id" ON "public"."article_translations" ("article_id");
-- Create index "idx_article_translations_deleted_at" to table: "article_translations"
CREATE INDEX "idx_article_translations_deleted_at" ON "public"."article_translations" ("deleted_at");
-- Create index "idx_article_translations_language" to table: "article_translations"
CREATE INDEX "idx_article_translations_language" ON "public"."article_translations" ("language");
-- Create index "idx_article_translations_translated_by" to table: "article_translations"
CREATE INDEX "idx_article_translations_translated_by" ON "public"."article_translations" ("translator_id");
-- Create "bookmarks" table
CREATE TABLE "public"."bookmarks" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "article_id" bigint NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_articles_bookmarks" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_users_bookmarks" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_bookmarks_article_id" to table: "bookmarks"
CREATE INDEX "idx_bookmarks_article_id" ON "public"."bookmarks" ("article_id");
-- Create index "idx_bookmarks_user_id" to table: "bookmarks"
CREATE INDEX "idx_bookmarks_user_id" ON "public"."bookmarks" ("user_id");
-- Create "breaking_news_banners" table
CREATE TABLE "public"."breaking_news_banners" (
  "id" bigserial NOT NULL,
  "title" character varying(255) NOT NULL,
  "content" text NULL,
  "article_id" bigint NULL,
  "priority" bigint NOT NULL DEFAULT 1,
  "style" character varying(50) NULL DEFAULT 'urgent',
  "text_color" character varying(7) NULL DEFAULT '#FFFFFF',
  "background_color" character varying(7) NULL DEFAULT '#DC2626',
  "start_time" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "end_time" timestamptz NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_breaking_news_banners_article" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_breaking_news_banners_article_id" to table: "breaking_news_banners"
CREATE INDEX "idx_breaking_news_banners_article_id" ON "public"."breaking_news_banners" ("article_id");
-- Create index "idx_breaking_news_banners_deleted_at" to table: "breaking_news_banners"
CREATE INDEX "idx_breaking_news_banners_deleted_at" ON "public"."breaking_news_banners" ("deleted_at");
-- Create "category_translations" table
CREATE TABLE "public"."category_translations" (
  "id" bigserial NOT NULL,
  "category_id" bigint NOT NULL,
  "language" character varying(5) NOT NULL,
  "name" character varying(100) NOT NULL,
  "slug" character varying(100) NOT NULL,
  "description" text NULL,
  "meta_title" character varying(255) NULL,
  "meta_desc" character varying(255) NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_category_translations_category" FOREIGN KEY ("category_id") REFERENCES "public"."categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_category_translations_category_id" to table: "category_translations"
CREATE INDEX "idx_category_translations_category_id" ON "public"."category_translations" ("category_id");
-- Create index "idx_category_translations_deleted_at" to table: "category_translations"
CREATE INDEX "idx_category_translations_deleted_at" ON "public"."category_translations" ("deleted_at");
-- Create index "idx_category_translations_language" to table: "category_translations"
CREATE INDEX "idx_category_translations_language" ON "public"."category_translations" ("language");
-- Create "comments" table
CREATE TABLE "public"."comments" (
  "id" bigserial NOT NULL,
  "article_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "parent_id" bigint NULL,
  "content" text NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'approved',
  "is_edited" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_articles_comments" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_comments_replies" FOREIGN KEY ("parent_id") REFERENCES "public"."comments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_users_comments" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_comments_article_id" to table: "comments"
CREATE INDEX "idx_comments_article_id" ON "public"."comments" ("article_id");
-- Create index "idx_comments_created_at" to table: "comments"
CREATE INDEX "idx_comments_created_at" ON "public"."comments" ("created_at");
-- Create index "idx_comments_deleted_at" to table: "comments"
CREATE INDEX "idx_comments_deleted_at" ON "public"."comments" ("deleted_at");
-- Create index "idx_comments_parent_id" to table: "comments"
CREATE INDEX "idx_comments_parent_id" ON "public"."comments" ("parent_id");
-- Create index "idx_comments_status" to table: "comments"
CREATE INDEX "idx_comments_status" ON "public"."comments" ("status");
-- Create index "idx_comments_user_id" to table: "comments"
CREATE INDEX "idx_comments_user_id" ON "public"."comments" ("user_id");
-- Create "content_analyses" table
CREATE TABLE "public"."content_analyses" (
  "id" bigserial NOT NULL,
  "article_id" bigint NOT NULL,
  "reading_level" character varying(20) NULL,
  "sentiment" character varying(20) NULL,
  "keywords" json NULL,
  "categories" json NULL,
  "tags" json NULL,
  "summary" text NULL,
  "read_time" bigint NULL DEFAULT 0,
  "quality" numeric(3,2) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_content_analyses_article" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_content_analyses_article_id" to table: "content_analyses"
CREATE INDEX "idx_content_analyses_article_id" ON "public"."content_analyses" ("article_id");
-- Create "content_suggestions" table
CREATE TABLE "public"."content_suggestions" (
  "id" bigserial NOT NULL,
  "type" character varying(50) NOT NULL,
  "input" text NOT NULL,
  "suggestion" text NOT NULL,
  "context" json NULL,
  "confidence" numeric(3,2) NULL,
  "user_id" bigint NULL,
  "article_id" bigint NULL,
  "used" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_content_suggestions_article" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_content_suggestions_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_content_suggestions_article_id" to table: "content_suggestions"
CREATE INDEX "idx_content_suggestions_article_id" ON "public"."content_suggestions" ("article_id");
-- Create index "idx_content_suggestions_user_id" to table: "content_suggestions"
CREATE INDEX "idx_content_suggestions_user_id" ON "public"."content_suggestions" ("user_id");
-- Create "follows" table
CREATE TABLE "public"."follows" (
  "id" bigserial NOT NULL,
  "follower_id" bigint NOT NULL,
  "following_id" bigint NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_followers" FOREIGN KEY ("following_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_users_following" FOREIGN KEY ("follower_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_follows_follower_id" to table: "follows"
CREATE INDEX "idx_follows_follower_id" ON "public"."follows" ("follower_id");
-- Create index "idx_follows_following_id" to table: "follows"
CREATE INDEX "idx_follows_following_id" ON "public"."follows" ("following_id");
-- Create "live_news_streams" table
CREATE TABLE "public"."live_news_streams" (
  "id" bigserial NOT NULL,
  "title" character varying(255) NOT NULL,
  "description" text NULL,
  "status" character varying(20) NOT NULL DEFAULT 'scheduled',
  "start_time" timestamptz NULL,
  "end_time" timestamptz NULL,
  "is_highlighted" boolean NULL DEFAULT false,
  "viewer_count" bigint NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_live_news_streams_deleted_at" to table: "live_news_streams"
CREATE INDEX "idx_live_news_streams_deleted_at" ON "public"."live_news_streams" ("deleted_at");
-- Create "live_news_updates" table
CREATE TABLE "public"."live_news_updates" (
  "id" bigserial NOT NULL,
  "stream_id" bigint NOT NULL,
  "title" character varying(255) NOT NULL,
  "content" text NOT NULL,
  "update_type" character varying(50) NULL DEFAULT 'update',
  "importance" character varying(20) NULL DEFAULT 'normal',
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_live_news_streams_updates" FOREIGN KEY ("stream_id") REFERENCES "public"."live_news_streams" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_live_news_updates_deleted_at" to table: "live_news_updates"
CREATE INDEX "idx_live_news_updates_deleted_at" ON "public"."live_news_updates" ("deleted_at");
-- Create index "idx_live_news_updates_stream_id" to table: "live_news_updates"
CREATE INDEX "idx_live_news_updates_stream_id" ON "public"."live_news_updates" ("stream_id");
-- Create "media" table
CREATE TABLE "public"."media" (
  "id" bigserial NOT NULL,
  "file_name" character varying(255) NOT NULL,
  "original_name" character varying(255) NOT NULL,
  "mime_type" character varying(100) NOT NULL,
  "size" bigint NOT NULL,
  "path" character varying(500) NOT NULL,
  "url" character varying(500) NOT NULL,
  "alt_text" character varying(255) NULL,
  "caption" text NULL,
  "uploaded_by" bigint NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_media_uploader" FOREIGN KEY ("uploaded_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_media_deleted_at" to table: "media"
CREATE INDEX "idx_media_deleted_at" ON "public"."media" ("deleted_at");
-- Create "menus" table
CREATE TABLE "public"."menus" (
  "id" bigserial NOT NULL,
  "name" character varying(100) NOT NULL,
  "slug" character varying(100) NOT NULL,
  "location" character varying(50) NOT NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_menus_slug" UNIQUE ("slug")
);
-- Create index "idx_menus_deleted_at" to table: "menus"
CREATE INDEX "idx_menus_deleted_at" ON "public"."menus" ("deleted_at");
-- Create "menu_items" table
CREATE TABLE "public"."menu_items" (
  "id" bigserial NOT NULL,
  "menu_id" bigint NOT NULL,
  "parent_id" bigint NULL,
  "title" character varying(100) NOT NULL,
  "url" character varying(255) NULL,
  "category_id" bigint NULL,
  "icon" character varying(50) NULL,
  "target" character varying(20) NULL DEFAULT '_self',
  "sort_order" bigint NULL DEFAULT 0,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_menu_items_category" FOREIGN KEY ("category_id") REFERENCES "public"."categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_menu_items_children" FOREIGN KEY ("parent_id") REFERENCES "public"."menu_items" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_menus_items" FOREIGN KEY ("menu_id") REFERENCES "public"."menus" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_menu_items_category_id" to table: "menu_items"
CREATE INDEX "idx_menu_items_category_id" ON "public"."menu_items" ("category_id");
-- Create index "idx_menu_items_deleted_at" to table: "menu_items"
CREATE INDEX "idx_menu_items_deleted_at" ON "public"."menu_items" ("deleted_at");
-- Create index "idx_menu_items_menu_id" to table: "menu_items"
CREATE INDEX "idx_menu_items_menu_id" ON "public"."menu_items" ("menu_id");
-- Create index "idx_menu_items_parent_id" to table: "menu_items"
CREATE INDEX "idx_menu_items_parent_id" ON "public"."menu_items" ("parent_id");
-- Create "menu_item_translations" table
CREATE TABLE "public"."menu_item_translations" (
  "id" bigserial NOT NULL,
  "menu_item_id" bigint NOT NULL,
  "language" character varying(5) NOT NULL,
  "title" character varying(100) NOT NULL,
  "url" character varying(255) NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_menu_item_translations_menu_item" FOREIGN KEY ("menu_item_id") REFERENCES "public"."menu_items" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_menu_item_translations_deleted_at" to table: "menu_item_translations"
CREATE INDEX "idx_menu_item_translations_deleted_at" ON "public"."menu_item_translations" ("deleted_at");
-- Create index "idx_menu_item_translations_language" to table: "menu_item_translations"
CREATE INDEX "idx_menu_item_translations_language" ON "public"."menu_item_translations" ("language");
-- Create index "idx_menu_item_translations_menu_item_id" to table: "menu_item_translations"
CREATE INDEX "idx_menu_item_translations_menu_item_id" ON "public"."menu_item_translations" ("menu_item_id");
-- Create "menu_translations" table
CREATE TABLE "public"."menu_translations" (
  "id" bigserial NOT NULL,
  "menu_id" bigint NOT NULL,
  "language" character varying(5) NOT NULL,
  "name" character varying(100) NOT NULL,
  "description" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_menu_translations_menu" FOREIGN KEY ("menu_id") REFERENCES "public"."menus" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_menu_translations_deleted_at" to table: "menu_translations"
CREATE INDEX "idx_menu_translations_deleted_at" ON "public"."menu_translations" ("deleted_at");
-- Create index "idx_menu_translations_language" to table: "menu_translations"
CREATE INDEX "idx_menu_translations_language" ON "public"."menu_translations" ("language");
-- Create index "idx_menu_translations_menu_id" to table: "menu_translations"
CREATE INDEX "idx_menu_translations_menu_id" ON "public"."menu_translations" ("menu_id");
-- Create "moderation_results" table
CREATE TABLE "public"."moderation_results" (
  "id" bigserial NOT NULL,
  "content_type" character varying(20) NOT NULL,
  "content_id" bigint NOT NULL,
  "content" text NOT NULL,
  "is_approved" boolean NULL DEFAULT false,
  "confidence" numeric(3,2) NULL,
  "reason" text NULL,
  "categories" json NULL,
  "severity" character varying(20) NULL DEFAULT 'low',
  "reviewed_by" bigint NULL,
  "reviewed_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_moderation_results_reviewer" FOREIGN KEY ("reviewed_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_moderation_results_reviewed_by" to table: "moderation_results"
CREATE INDEX "idx_moderation_results_reviewed_by" ON "public"."moderation_results" ("reviewed_by");
-- Create "news_stories" table
CREATE TABLE "public"."news_stories" (
  "id" bigserial NOT NULL,
  "headline" character varying(255) NOT NULL,
  "image_url" character varying(255) NOT NULL,
  "background_color" character varying(7) NULL DEFAULT '#000000',
  "text_color" character varying(7) NULL DEFAULT '#FFFFFF',
  "duration" bigint NULL DEFAULT 5,
  "article_id" bigint NULL,
  "external_url" character varying(255) NULL,
  "sort_order" bigint NULL DEFAULT 0,
  "start_time" timestamptz NOT NULL,
  "view_count" bigint NULL DEFAULT 0,
  "create_user_id" bigint NOT NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_news_stories_article" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_news_stories_article_id" to table: "news_stories"
CREATE INDEX "idx_news_stories_article_id" ON "public"."news_stories" ("article_id");
-- Create index "idx_news_stories_deleted_at" to table: "news_stories"
CREATE INDEX "idx_news_stories_deleted_at" ON "public"."news_stories" ("deleted_at");
-- Create "newsletters" table
CREATE TABLE "public"."newsletters" (
  "id" bigserial NOT NULL,
  "title" character varying(255) NOT NULL,
  "subject" character varying(255) NOT NULL,
  "content" text NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'draft',
  "sent_at" timestamptz NULL,
  "scheduled_at" timestamptz NULL,
  "created_by" bigint NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_newsletters_creator" FOREIGN KEY ("created_by") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_newsletters_deleted_at" to table: "newsletters"
CREATE INDEX "idx_newsletters_deleted_at" ON "public"."newsletters" ("deleted_at");
-- Create "notifications" table
CREATE TABLE "public"."notifications" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "type" character varying(50) NOT NULL,
  "title" character varying(255) NOT NULL,
  "message" text NOT NULL,
  "data" json NULL,
  "is_read" boolean NULL DEFAULT false,
  "read_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_notifications_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_notifications_deleted_at" to table: "notifications"
CREATE INDEX "idx_notifications_deleted_at" ON "public"."notifications" ("deleted_at");
-- Create index "idx_notifications_user_id" to table: "notifications"
CREATE INDEX "idx_notifications_user_id" ON "public"."notifications" ("user_id");
-- Create "notification_translations" table
CREATE TABLE "public"."notification_translations" (
  "id" bigserial NOT NULL,
  "notification_id" bigint NOT NULL,
  "language" character varying(5) NOT NULL,
  "title" character varying(255) NOT NULL,
  "message" text NOT NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_notification_translations_notification" FOREIGN KEY ("notification_id") REFERENCES "public"."notifications" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_notification_translations_deleted_at" to table: "notification_translations"
CREATE INDEX "idx_notification_translations_deleted_at" ON "public"."notification_translations" ("deleted_at");
-- Create index "idx_notification_translations_language" to table: "notification_translations"
CREATE INDEX "idx_notification_translations_language" ON "public"."notification_translations" ("language");
-- Create index "idx_notification_translations_notification_id" to table: "notification_translations"
CREATE INDEX "idx_notification_translations_notification_id" ON "public"."notification_translations" ("notification_id");
-- Create "pages" table
CREATE TABLE "public"."pages" (
  "id" bigserial NOT NULL,
  "title" character varying(255) NOT NULL,
  "slug" character varying(255) NOT NULL,
  "meta_title" character varying(255) NULL,
  "meta_desc" character varying(255) NULL,
  "template" character varying(50) NULL DEFAULT 'default',
  "layout" character varying(50) NULL DEFAULT 'container',
  "status" character varying(20) NOT NULL DEFAULT 'draft',
  "language" character varying(5) NULL DEFAULT 'tr',
  "parent_id" bigint NULL,
  "sort_order" bigint NULL DEFAULT 0,
  "featured_image" character varying(255) NULL,
  "excerpt_text" text NULL,
  "seo_settings" jsonb NULL,
  "page_settings" jsonb NULL,
  "layout_data" jsonb NULL,
  "author_id" bigint NOT NULL,
  "published_at" timestamptz NULL,
  "scheduled_at" timestamptz NULL,
  "views" bigint NULL DEFAULT 0,
  "is_homepage" boolean NULL DEFAULT false,
  "is_landing_page" boolean NULL DEFAULT false,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_pages_slug" UNIQUE ("slug"),
  CONSTRAINT "fk_pages_author" FOREIGN KEY ("author_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_pages_children" FOREIGN KEY ("parent_id") REFERENCES "public"."pages" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_pages_author_id" to table: "pages"
CREATE INDEX "idx_pages_author_id" ON "public"."pages" ("author_id");
-- Create index "idx_pages_deleted_at" to table: "pages"
CREATE INDEX "idx_pages_deleted_at" ON "public"."pages" ("deleted_at");
-- Create index "idx_pages_parent_id" to table: "pages"
CREATE INDEX "idx_pages_parent_id" ON "public"."pages" ("parent_id");
-- Create "page_content_blocks" table
CREATE TABLE "public"."page_content_blocks" (
  "id" bigserial NOT NULL,
  "page_id" bigint NOT NULL,
  "container_id" bigint NULL,
  "block_type" character varying(50) NOT NULL,
  "content" text NULL,
  "settings" jsonb NULL,
  "styles" jsonb NULL,
  "position" bigint NOT NULL,
  "is_visible" boolean NULL DEFAULT true,
  "is_container" boolean NULL DEFAULT false,
  "container_type" character varying(30) NULL,
  "grid_settings" jsonb NULL,
  "responsive_data" jsonb NULL,
  "ai_generated" boolean NULL DEFAULT false,
  "performance_data" jsonb NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_page_content_blocks_child_blocks" FOREIGN KEY ("container_id") REFERENCES "public"."page_content_blocks" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_pages_content_blocks" FOREIGN KEY ("page_id") REFERENCES "public"."pages" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_page_content_blocks_block_type" to table: "page_content_blocks"
CREATE INDEX "idx_page_content_blocks_block_type" ON "public"."page_content_blocks" ("block_type");
-- Create index "idx_page_content_blocks_container_id" to table: "page_content_blocks"
CREATE INDEX "idx_page_content_blocks_container_id" ON "public"."page_content_blocks" ("container_id");
-- Create index "idx_page_content_blocks_deleted_at" to table: "page_content_blocks"
CREATE INDEX "idx_page_content_blocks_deleted_at" ON "public"."page_content_blocks" ("deleted_at");
-- Create index "idx_page_content_blocks_page_id" to table: "page_content_blocks"
CREATE INDEX "idx_page_content_blocks_page_id" ON "public"."page_content_blocks" ("page_id");
-- Create index "idx_page_content_blocks_position" to table: "page_content_blocks"
CREATE INDEX "idx_page_content_blocks_position" ON "public"."page_content_blocks" ("position");
-- Create "page_templates" table
CREATE TABLE "public"."page_templates" (
  "id" bigserial NOT NULL,
  "name" character varying(100) NOT NULL,
  "description" text NULL,
  "category" character varying(50) NULL,
  "thumbnail" character varying(255) NULL,
  "preview_image" character varying(255) NULL,
  "block_structure" jsonb NOT NULL,
  "default_styles" jsonb NULL,
  "is_public" boolean NULL DEFAULT false,
  "is_premium" boolean NULL DEFAULT false,
  "usage_count" bigint NULL DEFAULT 0,
  "rating" numeric NULL DEFAULT 0,
  "tags" jsonb NULL,
  "creator_id" bigint NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_page_templates_creator" FOREIGN KEY ("creator_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_page_templates_creator_id" to table: "page_templates"
CREATE INDEX "idx_page_templates_creator_id" ON "public"."page_templates" ("creator_id");
-- Create index "idx_page_templates_deleted_at" to table: "page_templates"
CREATE INDEX "idx_page_templates_deleted_at" ON "public"."page_templates" ("deleted_at");
-- Create "related_articles" table
CREATE TABLE "public"."related_articles" (
  "id" bigserial NOT NULL,
  "article_id" bigint NOT NULL,
  "related_id" bigint NOT NULL,
  "score" bigint NULL DEFAULT 0,
  "type" character varying(20) NULL DEFAULT 'auto',
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_articles_related" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_related_articles_related" FOREIGN KEY ("related_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_related_articles_article_id" to table: "related_articles"
CREATE INDEX "idx_related_articles_article_id" ON "public"."related_articles" ("article_id");
-- Create index "idx_related_articles_related_id" to table: "related_articles"
CREATE INDEX "idx_related_articles_related_id" ON "public"."related_articles" ("related_id");
-- Create "story_groups" table
CREATE TABLE "public"."story_groups" (
  "id" bigserial NOT NULL,
  "title" character varying(100) NOT NULL,
  "description" character varying(255) NULL,
  "cover_image_url" character varying(255) NULL,
  "sort_order" bigint NULL DEFAULT 0,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_story_groups_deleted_at" to table: "story_groups"
CREATE INDEX "idx_story_groups_deleted_at" ON "public"."story_groups" ("deleted_at");
-- Create "story_group_items" table
CREATE TABLE "public"."story_group_items" (
  "story_group_id" bigint NOT NULL,
  "story_id" bigint NOT NULL,
  "sort_order" bigint NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("story_group_id", "story_id"),
  CONSTRAINT "fk_story_group_items_story" FOREIGN KEY ("story_id") REFERENCES "public"."news_stories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_story_group_items_story_group" FOREIGN KEY ("story_group_id") REFERENCES "public"."story_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "story_views" table
CREATE TABLE "public"."story_views" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "story_id" bigint NOT NULL,
  "viewed_at" timestamptz NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_story_views_story" FOREIGN KEY ("story_id") REFERENCES "public"."news_stories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_story_views_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_story_views_story_id" to table: "story_views"
CREATE INDEX "idx_story_views_story_id" ON "public"."story_views" ("story_id");
-- Create index "idx_story_views_user_id" to table: "story_views"
CREATE INDEX "idx_story_views_user_id" ON "public"."story_views" ("user_id");
-- Create "subscriptions" table
CREATE TABLE "public"."subscriptions" (
  "id" bigserial NOT NULL,
  "user_id" bigint NULL,
  "email" character varying(100) NOT NULL,
  "type" character varying(20) NOT NULL,
  "category_id" bigint NULL,
  "tag_id" bigint NULL,
  "is_active" boolean NULL DEFAULT true,
  "token" character varying(255) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "uni_subscriptions_token" UNIQUE ("token"),
  CONSTRAINT "fk_subscriptions_category" FOREIGN KEY ("category_id") REFERENCES "public"."categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_subscriptions_tag" FOREIGN KEY ("tag_id") REFERENCES "public"."tags" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_users_subscriptions" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_subscriptions_category_id" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_category_id" ON "public"."subscriptions" ("category_id");
-- Create index "idx_subscriptions_deleted_at" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_deleted_at" ON "public"."subscriptions" ("deleted_at");
-- Create index "idx_subscriptions_tag_id" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_tag_id" ON "public"."subscriptions" ("tag_id");
-- Create index "idx_subscriptions_user_id" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_user_id" ON "public"."subscriptions" ("user_id");
-- Create "tag_translations" table
CREATE TABLE "public"."tag_translations" (
  "id" bigserial NOT NULL,
  "tag_id" bigint NOT NULL,
  "language" character varying(5) NOT NULL,
  "name" character varying(50) NOT NULL,
  "slug" character varying(50) NOT NULL,
  "description" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_tag_translations_tag" FOREIGN KEY ("tag_id") REFERENCES "public"."tags" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_tag_translations_deleted_at" to table: "tag_translations"
CREATE INDEX "idx_tag_translations_deleted_at" ON "public"."tag_translations" ("deleted_at");
-- Create index "idx_tag_translations_language" to table: "tag_translations"
CREATE INDEX "idx_tag_translations_language" ON "public"."tag_translations" ("language");
-- Create index "idx_tag_translations_tag_id" to table: "tag_translations"
CREATE INDEX "idx_tag_translations_tag_id" ON "public"."tag_translations" ("tag_id");
-- Create "user_article_interactions" table
CREATE TABLE "public"."user_article_interactions" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "article_id" bigint NOT NULL,
  "interaction_type" character varying(20) NOT NULL,
  "duration" bigint NULL,
  "completion_rate" numeric(3,2) NULL,
  "platform" character varying(20) NULL,
  "user_agent" character varying(255) NULL,
  "ip_address" character varying(45) NULL,
  "referrer_url" character varying(500) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_articles_user_interactions" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_users_article_interactions" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_user_article_interactions_article_id" to table: "user_article_interactions"
CREATE INDEX "idx_user_article_interactions_article_id" ON "public"."user_article_interactions" ("article_id");
-- Create index "idx_user_article_interactions_created_at" to table: "user_article_interactions"
CREATE INDEX "idx_user_article_interactions_created_at" ON "public"."user_article_interactions" ("created_at");
-- Create index "idx_user_article_interactions_deleted_at" to table: "user_article_interactions"
CREATE INDEX "idx_user_article_interactions_deleted_at" ON "public"."user_article_interactions" ("deleted_at");
-- Create index "idx_user_article_interactions_interaction_type" to table: "user_article_interactions"
CREATE INDEX "idx_user_article_interactions_interaction_type" ON "public"."user_article_interactions" ("interaction_type");
-- Create index "idx_user_article_interactions_user_id" to table: "user_article_interactions"
CREATE INDEX "idx_user_article_interactions_user_id" ON "public"."user_article_interactions" ("user_id");
-- Create "videos" table
CREATE TABLE "public"."videos" (
  "id" bigserial NOT NULL,
  "title" character varying(200) NOT NULL,
  "description" text NULL,
  "video_url" character varying(500) NOT NULL,
  "thumbnail_url" character varying(500) NULL,
  "duration" bigint NULL,
  "file_size" bigint NULL,
  "resolution" character varying(20) NULL,
  "category_id" bigint NULL,
  "tags" text NULL,
  "user_id" bigint NOT NULL,
  "is_generated" boolean NULL DEFAULT false,
  "status" character varying(20) NULL DEFAULT 'pending',
  "is_public" boolean NULL DEFAULT true,
  "is_featured" boolean NULL DEFAULT false,
  "ai_generated" boolean NULL DEFAULT false,
  "ai_confidence" numeric NULL DEFAULT 0,
  "content_warning" character varying(100) NULL,
  "view_count" bigint NULL DEFAULT 0,
  "like_count" bigint NULL DEFAULT 0,
  "dislike_count" bigint NULL DEFAULT 0,
  "comment_count" bigint NULL DEFAULT 0,
  "share_count" bigint NULL DEFAULT 0,
  "published_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_videos_category" FOREIGN KEY ("category_id") REFERENCES "public"."categories" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_videos_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_videos_status" CHECK ((status)::text = ANY (ARRAY[('draft'::character varying)::text, ('pending'::character varying)::text, ('published'::character varying)::text, ('rejected'::character varying)::text, ('archived'::character varying)::text]))
);
-- Create index "idx_videos_ai_generated" to table: "videos"
CREATE INDEX "idx_videos_ai_generated" ON "public"."videos" ("ai_generated");
-- Create index "idx_videos_category_id" to table: "videos"
CREATE INDEX "idx_videos_category_id" ON "public"."videos" ("category_id");
-- Create index "idx_videos_created_at" to table: "videos"
CREATE INDEX "idx_videos_created_at" ON "public"."videos" ("created_at");
-- Create index "idx_videos_deleted_at" to table: "videos"
CREATE INDEX "idx_videos_deleted_at" ON "public"."videos" ("deleted_at");
-- Create index "idx_videos_is_featured" to table: "videos"
CREATE INDEX "idx_videos_is_featured" ON "public"."videos" ("is_featured");
-- Create index "idx_videos_is_public" to table: "videos"
CREATE INDEX "idx_videos_is_public" ON "public"."videos" ("is_public");
-- Create index "idx_videos_like_count" to table: "videos"
CREATE INDEX "idx_videos_like_count" ON "public"."videos" ("like_count");
-- Create index "idx_videos_published_at" to table: "videos"
CREATE INDEX "idx_videos_published_at" ON "public"."videos" ("published_at");
-- Create index "idx_videos_status" to table: "videos"
CREATE INDEX "idx_videos_status" ON "public"."videos" ("status");
-- Create index "idx_videos_user_id" to table: "videos"
CREATE INDEX "idx_videos_user_id" ON "public"."videos" ("user_id");
-- Create index "idx_videos_view_count" to table: "videos"
CREATE INDEX "idx_videos_view_count" ON "public"."videos" ("view_count");
-- Create "video_comments" table
CREATE TABLE "public"."video_comments" (
  "id" bigserial NOT NULL,
  "video_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "content" text NOT NULL,
  "parent_id" bigint NULL,
  "status" character varying(20) NULL DEFAULT 'active',
  "is_edited" boolean NULL DEFAULT false,
  "edited_at" timestamptz NULL,
  "ai_moderated" boolean NULL DEFAULT false,
  "ai_confidence" numeric NULL DEFAULT 0,
  "toxicity_score" numeric NULL DEFAULT 0,
  "like_count" bigint NULL DEFAULT 0,
  "dislike_count" bigint NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_video_comments_replies" FOREIGN KEY ("parent_id") REFERENCES "public"."video_comments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_video_comments_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_videos_comments" FOREIGN KEY ("video_id") REFERENCES "public"."videos" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_video_comments_status" CHECK ((status)::text = ANY (ARRAY[('active'::character varying)::text, ('hidden'::character varying)::text, ('deleted'::character varying)::text, ('flagged'::character varying)::text]))
);
-- Create index "idx_video_comments_deleted_at" to table: "video_comments"
CREATE INDEX "idx_video_comments_deleted_at" ON "public"."video_comments" ("deleted_at");
-- Create index "idx_video_comments_parent_id" to table: "video_comments"
CREATE INDEX "idx_video_comments_parent_id" ON "public"."video_comments" ("parent_id");
-- Create index "idx_video_comments_user_id" to table: "video_comments"
CREATE INDEX "idx_video_comments_user_id" ON "public"."video_comments" ("user_id");
-- Create index "idx_video_comments_video_id" to table: "video_comments"
CREATE INDEX "idx_video_comments_video_id" ON "public"."video_comments" ("video_id");
-- Create "video_comment_votes" table
CREATE TABLE "public"."video_comment_votes" (
  "id" bigserial NOT NULL,
  "comment_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "type" character varying(10) NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_video_comment_votes_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_video_comments_votes" FOREIGN KEY ("comment_id") REFERENCES "public"."video_comments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_video_comment_votes_type" CHECK ((type)::text = ANY (ARRAY[('like'::character varying)::text, ('dislike'::character varying)::text]))
);
-- Create index "idx_video_comment_votes_comment_id" to table: "video_comment_votes"
CREATE INDEX "idx_video_comment_votes_comment_id" ON "public"."video_comment_votes" ("comment_id");
-- Create index "idx_video_comment_votes_deleted_at" to table: "video_comment_votes"
CREATE INDEX "idx_video_comment_votes_deleted_at" ON "public"."video_comment_votes" ("deleted_at");
-- Create index "idx_video_comment_votes_user_id" to table: "video_comment_votes"
CREATE INDEX "idx_video_comment_votes_user_id" ON "public"."video_comment_votes" ("user_id");
-- Create "video_playlists" table
CREATE TABLE "public"."video_playlists" (
  "id" bigserial NOT NULL,
  "name" character varying(100) NOT NULL,
  "description" text NULL,
  "user_id" bigint NOT NULL,
  "is_public" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_video_playlists_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_video_playlists_deleted_at" to table: "video_playlists"
CREATE INDEX "idx_video_playlists_deleted_at" ON "public"."video_playlists" ("deleted_at");
-- Create index "idx_video_playlists_user_id" to table: "video_playlists"
CREATE INDEX "idx_video_playlists_user_id" ON "public"."video_playlists" ("user_id");
-- Create "video_playlist_items" table
CREATE TABLE "public"."video_playlist_items" (
  "id" bigserial NOT NULL,
  "playlist_id" bigint NOT NULL,
  "video_id" bigint NOT NULL,
  "order" bigint NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_video_playlist_items_video" FOREIGN KEY ("video_id") REFERENCES "public"."videos" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_video_playlists_items" FOREIGN KEY ("playlist_id") REFERENCES "public"."video_playlists" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_video_playlist_items_playlist_id" to table: "video_playlist_items"
CREATE INDEX "idx_video_playlist_items_playlist_id" ON "public"."video_playlist_items" ("playlist_id");
-- Create index "idx_video_playlist_items_video_id" to table: "video_playlist_items"
CREATE INDEX "idx_video_playlist_items_video_id" ON "public"."video_playlist_items" ("video_id");
-- Create "video_processing_jobs" table
CREATE TABLE "public"."video_processing_jobs" (
  "id" bigserial NOT NULL,
  "video_id" bigint NOT NULL,
  "job_type" character varying(50) NOT NULL,
  "status" character varying(20) NULL DEFAULT 'pending',
  "progress" bigint NULL DEFAULT 0,
  "error_msg" text NULL,
  "parameters" text NULL,
  "result" text NULL,
  "started_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_video_processing_jobs_video" FOREIGN KEY ("video_id") REFERENCES "public"."videos" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_video_processing_jobs_video_id" to table: "video_processing_jobs"
CREATE INDEX "idx_video_processing_jobs_video_id" ON "public"."video_processing_jobs" ("video_id");
-- Create "video_views" table
CREATE TABLE "public"."video_views" (
  "id" bigserial NOT NULL,
  "video_id" bigint NOT NULL,
  "user_id" bigint NULL,
  "ip_address" character varying(45) NULL,
  "user_agent" character varying(500) NULL,
  "duration" bigint NULL,
  "watch_percent" numeric NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_video_views_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_videos_views" FOREIGN KEY ("video_id") REFERENCES "public"."videos" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_video_views_user_id" to table: "video_views"
CREATE INDEX "idx_video_views_user_id" ON "public"."video_views" ("user_id");
-- Create index "idx_video_views_video_id" to table: "video_views"
CREATE INDEX "idx_video_views_video_id" ON "public"."video_views" ("video_id");
-- Create "video_votes" table
CREATE TABLE "public"."video_votes" (
  "id" bigserial NOT NULL,
  "video_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "type" character varying(10) NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_video_votes_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_videos_votes" FOREIGN KEY ("video_id") REFERENCES "public"."videos" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "chk_video_votes_type" CHECK ((type)::text = ANY (ARRAY[('like'::character varying)::text, ('dislike'::character varying)::text]))
);
-- Create index "idx_video_votes_deleted_at" to table: "video_votes"
CREATE INDEX "idx_video_votes_deleted_at" ON "public"."video_votes" ("deleted_at");
-- Create index "idx_video_votes_user_id" to table: "video_votes"
CREATE INDEX "idx_video_votes_user_id" ON "public"."video_votes" ("user_id");
-- Create index "idx_video_votes_video_id" to table: "video_votes"
CREATE INDEX "idx_video_votes_video_id" ON "public"."video_votes" ("video_id");
-- Create "votes" table
CREATE TABLE "public"."votes" (
  "id" bigserial NOT NULL,
  "user_id" bigint NOT NULL,
  "article_id" bigint NULL,
  "comment_id" bigint NULL,
  "type" character varying(10) NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_articles_votes" FOREIGN KEY ("article_id") REFERENCES "public"."articles" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_comments_votes" FOREIGN KEY ("comment_id") REFERENCES "public"."comments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_users_votes" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_votes_article_id" to table: "votes"
CREATE INDEX "idx_votes_article_id" ON "public"."votes" ("article_id");
-- Create index "idx_votes_comment_id" to table: "votes"
CREATE INDEX "idx_votes_comment_id" ON "public"."votes" ("comment_id");
-- Create index "idx_votes_created_at" to table: "votes"
CREATE INDEX "idx_votes_created_at" ON "public"."votes" ("created_at");
-- Create index "idx_votes_type" to table: "votes"
CREATE INDEX "idx_votes_type" ON "public"."votes" ("type");
-- Create index "idx_votes_user_id" to table: "votes"
CREATE INDEX "idx_votes_user_id" ON "public"."votes" ("user_id");
