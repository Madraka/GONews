package organized

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// SeedPageTemplates creates page templates for the CMS system
func SeedPageTemplates(db *sqlx.DB) error {
	fmt.Println("📄 Seeding page templates...")

	templates := []map[string]interface{}{
		{
			"name":            "Default Page",
			"description":     "Standard page template with header, content area, and footer",
			"category":        "basic",
			"block_structure": generateDefaultPageTemplate(),
			"is_public":       true,
			"creator_id":      1, // Admin
		},
		{
			"name":            "Landing Page",
			"description":     "Hero section with call-to-action and feature highlights",
			"category":        "marketing",
			"block_structure": generateLandingPageTemplate(),
			"is_public":       true,
			"creator_id":      1, // Admin
		},
		{
			"name":            "About Page",
			"description":     "Template for about us pages with team and mission sections",
			"category":        "company",
			"block_structure": generateAboutPageTemplate(),
			"is_public":       true,
			"creator_id":      1, // Admin
		},
		{
			"name":            "Contact Page",
			"description":     "Contact form with location and contact information",
			"category":        "contact",
			"block_structure": generateContactPageTemplate(),
			"is_public":       true,
			"creator_id":      1, // Admin
		},
		{
			"name":            "News Archive",
			"description":     "Template for displaying news articles with pagination",
			"category":        "news",
			"block_structure": generateNewsArchiveTemplate(),
			"is_public":       true,
			"creator_id":      2, // Editor
		},
		{
			"name":            "Category Page",
			"description":     "Template for category-specific article listings",
			"category":        "news",
			"block_structure": generateCategoryPageTemplate(),
			"is_public":       true,
			"creator_id":      2, // Editor
		},
		{
			"name":            "Privacy Policy",
			"description":     "Legal page template for privacy policy content",
			"category":        "legal",
			"block_structure": generatePrivacyPolicyTemplate(),
			"is_public":       true,
			"creator_id":      1, // Admin
		},
		{
			"name":            "Terms of Service",
			"description":     "Legal page template for terms and conditions",
			"category":        "legal",
			"block_structure": generateTermsOfServiceTemplate(),
			"is_public":       true,
			"creator_id":      1, // Admin
		},
		{
			"name":            "FAQ Page",
			"description":     "Frequently asked questions with collapsible sections",
			"category":        "support",
			"block_structure": generateFAQPageTemplate(),
			"is_public":       true,
			"creator_id":      2, // Editor
		},
		{
			"name":            "Search Results",
			"description":     "Template for displaying search results with filters",
			"category":        "utility",
			"block_structure": generateSearchResultsTemplate(),
			"is_public":       true,
			"creator_id":      2, // Editor
		},
	}

	for _, template := range templates {
		query := `
			INSERT INTO page_templates (
				name, description, category, block_structure, is_public, creator_id, created_at, updated_at
			) VALUES (
				:name, :description, :category, :block_structure, :is_public, :creator_id, NOW(), NOW()
			) RETURNING id`

		var templateID int
		stmt, err := db.PrepareNamed(query)
		if err != nil {
			return fmt.Errorf("error preparing page template query: %v", err)
		}
		defer func() {
			if closeErr := stmt.Close(); closeErr != nil {
				fmt.Printf("Warning: Error closing statement: %v\n", closeErr)
			}
		}()

		err = stmt.Get(&templateID, template)
		if err != nil {
			return fmt.Errorf("error inserting page template '%s': %v", template["name"], err)
		}

		fmt.Printf("   ✓ Created template: %s (ID: %d)\n", template["name"], templateID)
	}

	fmt.Printf("✅ Successfully seeded %d page templates\n", len(templates))
	return nil
}

// generateDefaultPageTemplate creates a basic page template
func generateDefaultPageTemplate() string {
	return `{
	"sections": [
		{
			"type": "header",
			"component": "page-header",
			"props": {
				"title": "{{page.title}}",
				"subtitle": "{{page.subtitle}}",
				"showBreadcrumb": true
			}
		},
		{
			"type": "content",
			"component": "page-content",
			"props": {
				"content": "{{page.content}}",
				"allowHtml": true,
				"showToc": false
			}
		},
		{
			"type": "footer",
			"component": "page-footer",
			"props": {
				"showSocial": true,
				"showNewsletter": false
			}
		}
	],
	"layout": "default",
	"meta": {
		"responsive": true,
		"seo": true
	}
}`
}

// generateLandingPageTemplate creates a landing page template
func generateLandingPageTemplate() string {
	return `{
	"sections": [
		{
			"type": "hero",
			"component": "hero-section",
			"props": {
				"title": "{{page.hero_title}}",
				"subtitle": "{{page.hero_subtitle}}",
				"backgroundImage": "{{page.hero_image}}",
				"ctaButton": {
					"text": "{{page.cta_text}}",
					"url": "{{page.cta_url}}",
					"style": "primary"
				}
			}
		},
		{
			"type": "features",
			"component": "features-grid",
			"props": {
				"title": "{{page.features_title}}",
				"features": "{{page.features}}",
				"columns": 3
			}
		},
		{
			"type": "testimonials",
			"component": "testimonials-carousel",
			"props": {
				"title": "What Our Readers Say",
				"testimonials": "{{page.testimonials}}",
				"autoplay": true,
				"interval": 5000
			}
		},
		{
			"type": "newsletter",
			"component": "newsletter-signup",
			"props": {
				"title": "Stay Updated",
				"description": "Get the latest news delivered to your inbox",
				"placeholder": "Enter your email address",
				"buttonText": "Subscribe"
			}
		}
	],
	"layout": "fullwidth",
	"meta": {
		"responsive": true,
		"seo": true,
		"optimized": true
	}
}`
}

// generateAboutPageTemplate creates an about page template
func generateAboutPageTemplate() string {
	return `{
	"sections": [
		{
			"type": "header",
			"component": "page-header",
			"props": {
				"title": "{{page.title}}",
				"subtitle": "{{page.subtitle}}",
				"backgroundImage": "{{page.header_image}}"
			}
		},
		{
			"type": "mission",
			"component": "mission-section",
			"props": {
				"title": "Our Mission",
				"content": "{{page.mission}}",
				"image": "{{page.mission_image}}",
				"imagePosition": "right"
			}
		},
		{
			"type": "team",
			"component": "team-grid",
			"props": {
				"title": "Meet Our Team",
				"members": "{{page.team_members}}",
				"columns": 3,
				"showSocial": true
			}
		},
		{
			"type": "history",
			"component": "timeline",
			"props": {
				"title": "Our Story",
				"events": "{{page.timeline}}",
				"orientation": "vertical"
			}
		},
		{
			"type": "values",
			"component": "values-section",
			"props": {
				"title": "Our Values",
				"values": "{{page.values}}",
				"layout": "cards"
			}
		}
	],
	"layout": "default",
	"meta": {
		"responsive": true,
		"seo": true
	}
}`
}

// generateContactPageTemplate creates a contact page template
func generateContactPageTemplate() string {
	return `{
	"sections": [
		{
			"type": "header",
			"component": "page-header",
			"props": {
				"title": "Contact Us",
				"subtitle": "Get in touch with our team"
			}
		},
		{
			"type": "contact-info",
			"component": "contact-info",
			"props": {
				"title": "Get In Touch",
				"address": "{{site.address}}",
				"phone": "{{site.phone}}",
				"email": "{{site.email}}",
				"hours": "{{site.business_hours}}",
				"showMap": true
			}
		},
		{
			"type": "contact-form",
			"component": "contact-form",
			"props": {
				"title": "Send Us a Message",
				"fields": [
					{
						"name": "name",
						"type": "text",
						"label": "Full Name",
						"required": true
					},
					{
						"name": "email",
						"type": "email",
						"label": "Email Address",
						"required": true
					},
					{
						"name": "subject",
						"type": "text",
						"label": "Subject",
						"required": true
					},
					{
						"name": "message",
						"type": "textarea",
						"label": "Message",
						"required": true,
						"rows": 5
					}
				],
				"submitText": "Send Message",
				"successMessage": "Thank you for your message. We'll get back to you soon!"
			}
		}
	],
	"layout": "default",
	"meta": {
		"responsive": true,
		"seo": true
	}
}`
}

// generateNewsArchiveTemplate creates a news archive template
func generateNewsArchiveTemplate() string {
	return `{
	"sections": [
		{
			"type": "header",
			"component": "archive-header",
			"props": {
				"title": "News Archive",
				"subtitle": "All the latest news and updates",
				"showSearch": true,
				"showFilters": true
			}
		},
		{
			"type": "filters",
			"component": "news-filters",
			"props": {
				"categories": "{{categories}}",
				"tags": "{{popular_tags}}",
				"dateRange": true,
				"sortOptions": ["date", "popularity", "title"]
			}
		},
		{
			"type": "articles",
			"component": "articles-grid",
			"props": {
				"articles": "{{articles}}",
				"layout": "grid",
				"columns": 2,
				"showExcerpt": true,
				"showMeta": true,
				"showTags": true
			}
		},
		{
			"type": "pagination",
			"component": "pagination",
			"props": {
				"currentPage": "{{current_page}}",
				"totalPages": "{{total_pages}}",
				"showNumbers": true,
				"showPrevNext": true
			}
		}
	],
	"layout": "default",
	"meta": {
		"responsive": true,
		"seo": true,
		"paginated": true
	}
}`
}

// generateCategoryPageTemplate creates a category page template
func generateCategoryPageTemplate() string {
	return `{
	"sections": [
		{
			"type": "header",
			"component": "category-header",
			"props": {
				"title": "{{category.name}}",
				"description": "{{category.description}}",
				"image": "{{category.image}}",
				"articleCount": "{{category.article_count}}"
			}
		},
		{
			"type": "featured",
			"component": "featured-articles",
			"props": {
				"title": "Featured Articles",
				"articles": "{{featured_articles}}",
				"limit": 3,
				"layout": "carousel"
			}
		},
		{
			"type": "articles",
			"component": "articles-list",
			"props": {
				"articles": "{{category_articles}}",
				"layout": "list",
				"showExcerpt": true,
				"showAuthor": true,
				"showDate": true,
				"showComments": true
			}
		},
		{
			"type": "sidebar",
			"component": "category-sidebar",
			"props": {
				"relatedCategories": "{{related_categories}}",
				"popularTags": "{{popular_tags}}",
				"recentArticles": "{{recent_articles}}"
			}
		}
	],
	"layout": "sidebar-right",
	"meta": {
		"responsive": true,
		"seo": true,
		"category": true
	}
}`
}

// generatePrivacyPolicyTemplate creates a privacy policy template
func generatePrivacyPolicyTemplate() string {
	return `{
	"sections": [
		{
			"type": "header",
			"component": "legal-header",
			"props": {
				"title": "Privacy Policy",
				"lastUpdated": "{{page.last_updated}}",
				"effectiveDate": "{{page.effective_date}}"
			}
		},
		{
			"type": "content",
			"component": "legal-content",
			"props": {
				"content": "{{page.content}}",
				"showToc": true,
				"numberedSections": true,
				"printable": true
			}
		},
		{
			"type": "contact",
			"component": "legal-contact",
			"props": {
				"title": "Questions About This Policy?",
				"email": "{{site.privacy_email}}",
				"address": "{{site.legal_address}}"
			}
		}
	],
	"layout": "legal",
	"meta": {
		"responsive": true,
		"seo": true,
		"legal": true
	}
}`
}

// generateTermsOfServiceTemplate creates a terms of service template
func generateTermsOfServiceTemplate() string {
	return `{
	"sections": [
		{
			"type": "header",
			"component": "legal-header",
			"props": {
				"title": "Terms of Service",
				"lastUpdated": "{{page.last_updated}}",
				"effectiveDate": "{{page.effective_date}}"
			}
		},
		{
			"type": "content",
			"component": "legal-content",
			"props": {
				"content": "{{page.content}}",
				"showToc": true,
				"numberedSections": true,
				"printable": true
			}
		},
		{
			"type": "acceptance",
			"component": "terms-acceptance",
			"props": {
				"title": "Acceptance of Terms",
				"description": "By using our service, you agree to these terms and conditions."
			}
		}
	],
	"layout": "legal",
	"meta": {
		"responsive": true,
		"seo": true,
		"legal": true
	}
}`
}

// generateFAQPageTemplate creates an FAQ page template
func generateFAQPageTemplate() string {
	return `{
	"sections": [
		{
			"type": "header",
			"component": "faq-header",
			"props": {
				"title": "Frequently Asked Questions",
				"subtitle": "Find answers to common questions",
				"showSearch": true
			}
		},
		{
			"type": "categories",
			"component": "faq-categories",
			"props": {
				"categories": "{{faq_categories}}",
				"layout": "tabs"
			}
		},
		{
			"type": "questions",
			"component": "faq-accordion",
			"props": {
				"questions": "{{faq_questions}}",
				"searchable": true,
				"collapsible": true,
				"multipleOpen": false
			}
		},
		{
			"type": "contact",
			"component": "faq-contact",
			"props": {
				"title": "Still Have Questions?",
				"description": "Contact our support team for help",
				"contactUrl": "/contact",
				"supportEmail": "{{site.support_email}}"
			}
		}
	],
	"layout": "default",
	"meta": {
		"responsive": true,
		"seo": true,
		"searchable": true
	}
}`
}

// generateSearchResultsTemplate creates a search results template
func generateSearchResultsTemplate() string {
	return `{
	"sections": [
		{
			"type": "header",
			"component": "search-header",
			"props": {
				"title": "Search Results",
				"query": "{{search_query}}",
				"resultCount": "{{result_count}}",
				"showRefineSearch": true
			}
		},
		{
			"type": "filters",
			"component": "search-filters",
			"props": {
				"categories": "{{categories}}",
				"tags": "{{tags}}",
				"dateRange": true,
				"contentType": ["articles", "pages"],
				"sortOptions": ["relevance", "date", "popularity"]
			}
		},
		{
			"type": "results",
			"component": "search-results",
			"props": {
				"results": "{{search_results}}",
				"highlightQuery": true,
				"showExcerpt": true,
				"showMeta": true,
				"showScore": false
			}
		},
		{
			"type": "suggestions",
			"component": "search-suggestions",
			"props": {
				"title": "Did you mean?",
				"suggestions": "{{search_suggestions}}",
				"popularQueries": "{{popular_queries}}"
			}
		}
	],
	"layout": "default",
	"meta": {
		"responsive": true,
		"seo": false,
		"searchable": false
	}
}`
}
