package organized

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

// SeedContentTranslations creates translations for categories, tags, menus, pages etc.
func SeedContentTranslations(db *sqlx.DB) error {
	fmt.Println("🌐 Seeding content translations...")

	// Seed Category Translations
	categoryTranslations := []map[string]interface{}{
		// Politics Category
		{"category_id": 1, "language": "en", "name": "Politics", "slug": "politics", "description": "Political news and analysis", "meta_title": "Politics News", "meta_desc": "Latest political news and analysis"},
		{"category_id": 1, "language": "tr", "name": "Politika", "slug": "politika", "description": "Siyasi haberler ve analizler", "meta_title": "Politika Haberleri", "meta_desc": "En son siyasi haberler ve analizler"},
		{"category_id": 1, "language": "es", "name": "Política", "slug": "politica", "description": "Noticias y análisis políticos", "meta_title": "Noticias Políticas", "meta_desc": "Últimas noticias y análisis políticos"},

		// Technology Category
		{"category_id": 2, "language": "en", "name": "Technology", "slug": "technology", "description": "Latest tech news and innovations", "meta_title": "Technology News", "meta_desc": "Latest technology news and innovations"},
		{"category_id": 2, "language": "tr", "name": "Teknoloji", "slug": "teknoloji", "description": "En son teknoloji haberleri ve yenilikler", "meta_title": "Teknoloji Haberleri", "meta_desc": "En son teknoloji haberleri ve yenilikler"},
		{"category_id": 2, "language": "es", "name": "Tecnología", "slug": "tecnologia", "description": "Últimas noticias e innovaciones tecnológicas", "meta_title": "Noticias de Tecnología", "meta_desc": "Últimas noticias e innovaciones tecnológicas"},

		// Business Category
		{"category_id": 3, "language": "en", "name": "Business", "slug": "business", "description": "Business and economic news", "meta_title": "Business News", "meta_desc": "Latest business and economic news"},
		{"category_id": 3, "language": "tr", "name": "İş Dünyası", "slug": "is-dunyasi", "description": "İş dünyası ve ekonomi haberleri", "meta_title": "İş Dünyası Haberleri", "meta_desc": "En son iş dünyası ve ekonomi haberleri"},
		{"category_id": 3, "language": "es", "name": "Negocios", "slug": "negocios", "description": "Noticias de negocios y economía", "meta_title": "Noticias de Negocios", "meta_desc": "Últimas noticias de negocios y economía"},

		// Sports Category
		{"category_id": 4, "language": "en", "name": "Sports", "slug": "sports", "description": "Sports news and updates", "meta_title": "Sports News", "meta_desc": "Latest sports news and updates"},
		{"category_id": 4, "language": "tr", "name": "Spor", "slug": "spor", "description": "Spor haberleri ve güncellemeleri", "meta_title": "Spor Haberleri", "meta_desc": "En son spor haberleri ve güncellemeleri"},
		{"category_id": 4, "language": "es", "name": "Deportes", "slug": "deportes", "description": "Noticias y actualizaciones deportivas", "meta_title": "Noticias Deportivas", "meta_desc": "Últimas noticias y actualizaciones deportivas"},

		// Health Category
		{"category_id": 5, "language": "en", "name": "Health", "slug": "health", "description": "Health and wellness news", "meta_title": "Health News", "meta_desc": "Latest health and wellness news"},
		{"category_id": 5, "language": "tr", "name": "Sağlık", "slug": "saglik", "description": "Sağlık ve wellness haberleri", "meta_title": "Sağlık Haberleri", "meta_desc": "En son sağlık ve wellness haberleri"},
		{"category_id": 5, "language": "es", "name": "Salud", "slug": "salud", "description": "Noticias de salud y bienestar", "meta_title": "Noticias de Salud", "meta_desc": "Últimas noticias de salud y bienestar"},

		// Entertainment Category
		{"category_id": 6, "language": "en", "name": "Entertainment", "slug": "entertainment", "description": "Entertainment and celebrity news", "meta_title": "Entertainment News", "meta_desc": "Latest entertainment and celebrity news"},
		{"category_id": 6, "language": "tr", "name": "Eğlence", "slug": "eglence", "description": "Eğlence ve ünlü haberleri", "meta_title": "Eğlence Haberleri", "meta_desc": "En son eğlence ve ünlü haberleri"},
		{"category_id": 6, "language": "es", "name": "Entretenimiento", "slug": "entretenimiento", "description": "Noticias de entretenimiento y celebridades", "meta_title": "Noticias de Entretenimiento", "meta_desc": "Últimas noticias de entretenimiento y celebridades"},
	}

	for _, catTrans := range categoryTranslations {
		query := `INSERT INTO category_translations (category_id, language, name, slug, description, meta_title, meta_desc, is_active, created_at, updated_at) 
				  VALUES (:category_id, :language, :name, :slug, :description, :meta_title, :meta_desc, true, NOW(), NOW()) 
				  ON CONFLICT (category_id, language) DO UPDATE SET 
				  	name = EXCLUDED.name, 
				  	slug = EXCLUDED.slug,
				  	description = EXCLUDED.description,
				  	meta_title = EXCLUDED.meta_title,
				  	meta_desc = EXCLUDED.meta_desc,
				  	updated_at = NOW()`
		_, err := db.NamedExec(query, catTrans)
		if err != nil {
			log.Printf("Failed to insert category translation: %v", err)
		}
	}

	// Seed Tag Translations
	tagTranslations := []map[string]interface{}{
		// Breaking News Tag
		{"tag_id": 1, "language": "en", "name": "Breaking News", "slug": "breaking-news", "description": "Latest breaking news stories", "meta_title": "Breaking News", "meta_desc": "Latest breaking news stories"},
		{"tag_id": 1, "language": "tr", "name": "Son Dakika", "slug": "son-dakika", "description": "Son dakika haber başlıkları", "meta_title": "Son Dakika", "meta_desc": "Son dakika haber başlıkları"},
		{"tag_id": 1, "language": "es", "name": "Últimas Noticias", "slug": "ultimas-noticias", "description": "Últimas noticias de último momento", "meta_title": "Últimas Noticias", "meta_desc": "Últimas noticias de último momento"},

		// Featured Tag
		{"tag_id": 2, "language": "en", "name": "Featured", "slug": "featured", "description": "Featured articles and stories", "meta_title": "Featured Stories", "meta_desc": "Featured articles and stories"},
		{"tag_id": 2, "language": "tr", "name": "Öne Çıkan", "slug": "one-cikan", "description": "Öne çıkan makaleler ve haberler", "meta_title": "Öne Çıkan Haberler", "meta_desc": "Öne çıkan makaleler ve haberler"},
		{"tag_id": 2, "language": "es", "name": "Destacado", "slug": "destacado", "description": "Artículos y noticias destacadas", "meta_title": "Noticias Destacadas", "meta_desc": "Artículos y noticias destacadas"},

		// Trending Tag
		{"tag_id": 3, "language": "en", "name": "Trending", "slug": "trending", "description": "Trending topics and news", "meta_title": "Trending News", "meta_desc": "Trending topics and news"},
		{"tag_id": 3, "language": "tr", "name": "Gündem", "slug": "gundem", "description": "Gündemde olan konular ve haberler", "meta_title": "Gündem Haberleri", "meta_desc": "Gündemde olan konular ve haberler"},
		{"tag_id": 3, "language": "es", "name": "Tendencia", "slug": "tendencia", "description": "Temas y noticias en tendencia", "meta_title": "Noticias en Tendencia", "meta_desc": "Temas y noticias en tendencia"},

		// Analysis Tag
		{"tag_id": 4, "language": "en", "name": "Analysis", "slug": "analysis", "description": "In-depth analysis and commentary", "meta_title": "News Analysis", "meta_desc": "In-depth analysis and commentary"},
		{"tag_id": 4, "language": "tr", "name": "Analiz", "slug": "analiz", "description": "Derinlemesine analiz ve yorumlar", "meta_title": "Haber Analizi", "meta_desc": "Derinlemesine analiz ve yorumlar"},
		{"tag_id": 4, "language": "es", "name": "Análisis", "slug": "analisis", "description": "Análisis en profundidad y comentarios", "meta_title": "Análisis de Noticias", "meta_desc": "Análisis en profundidad y comentarios"},

		// Opinion Tag
		{"tag_id": 5, "language": "en", "name": "Opinion", "slug": "opinion", "description": "Opinion pieces and editorials", "meta_title": "Opinion Articles", "meta_desc": "Opinion pieces and editorials"},
		{"tag_id": 5, "language": "tr", "name": "Görüş", "slug": "gorus", "description": "Görüş yazıları ve köşe yazıları", "meta_title": "Görüş Yazıları", "meta_desc": "Görüş yazıları ve köşe yazıları"},
		{"tag_id": 5, "language": "es", "name": "Opinión", "slug": "opinion", "description": "Artículos de opinión y editoriales", "meta_title": "Artículos de Opinión", "meta_desc": "Artículos de opinión y editoriales"},
	}

	for _, tagTrans := range tagTranslations {
		query := `INSERT INTO tag_translations (tag_id, language, name, slug, description, meta_title, meta_desc, is_active, created_at, updated_at) 
				  VALUES (:tag_id, :language, :name, :slug, :description, :meta_title, :meta_desc, true, NOW(), NOW()) 
				  ON CONFLICT (tag_id, language) DO UPDATE SET 
				  	name = EXCLUDED.name, 
				  	slug = EXCLUDED.slug,
				  	description = EXCLUDED.description,
				  	meta_title = EXCLUDED.meta_title,
				  	meta_desc = EXCLUDED.meta_desc,
				  	updated_at = NOW()`
		_, err := db.NamedExec(query, tagTrans)
		if err != nil {
			log.Printf("Failed to insert tag translation: %v", err)
		}
	}

	// Seed Menu Translations
	menuTranslations := []map[string]interface{}{
		// Main Menu
		{"menu_id": 1, "language": "en", "name": "Main Menu", "description": "Main navigation menu"},
		{"menu_id": 1, "language": "tr", "name": "Ana Menü", "description": "Ana navigasyon menüsü"},
		{"menu_id": 1, "language": "es", "name": "Menú Principal", "description": "Menú de navegación principal"},

		// Footer Menu
		{"menu_id": 2, "language": "en", "name": "Footer Menu", "description": "Footer navigation menu"},
		{"menu_id": 2, "language": "tr", "name": "Alt Menü", "description": "Alt kısım navigasyon menüsü"},
		{"menu_id": 2, "language": "es", "name": "Menú de Pie", "description": "Menú de navegación del pie de página"},
	}

	for _, menuTrans := range menuTranslations {
		query := `INSERT INTO menu_translations (menu_id, language, name, description, is_active, created_at, updated_at) 
				  VALUES (:menu_id, :language, :name, :description, true, NOW(), NOW()) 
				  ON CONFLICT (menu_id, language) DO UPDATE SET 
				  	name = EXCLUDED.name, 
				  	description = EXCLUDED.description,
				  	updated_at = NOW()`
		_, err := db.NamedExec(query, menuTrans)
		if err != nil {
			log.Printf("Failed to insert menu translation: %v", err)
		}
	}

	// Seed Menu Item Translations
	menuItemTranslations := []map[string]interface{}{
		// Home
		{"menu_item_id": 1, "language": "en", "title": "Home", "url": "/"},
		{"menu_item_id": 1, "language": "tr", "title": "Ana Sayfa", "url": "/"},
		{"menu_item_id": 1, "language": "es", "title": "Inicio", "url": "/"},

		// Politics
		{"menu_item_id": 2, "language": "en", "title": "Politics", "url": "/category/politics"},
		{"menu_item_id": 2, "language": "tr", "title": "Politika", "url": "/kategori/politika"},
		{"menu_item_id": 2, "language": "es", "title": "Política", "url": "/categoria/politica"},

		// Technology
		{"menu_item_id": 3, "language": "en", "title": "Technology", "url": "/category/technology"},
		{"menu_item_id": 3, "language": "tr", "title": "Teknoloji", "url": "/kategori/teknoloji"},
		{"menu_item_id": 3, "language": "es", "title": "Tecnología", "url": "/categoria/tecnologia"},

		// Business
		{"menu_item_id": 4, "language": "en", "title": "Business", "url": "/category/business"},
		{"menu_item_id": 4, "language": "tr", "title": "İş Dünyası", "url": "/kategori/is-dunyasi"},
		{"menu_item_id": 4, "language": "es", "title": "Negocios", "url": "/categoria/negocios"},

		// Sports
		{"menu_item_id": 5, "language": "en", "title": "Sports", "url": "/category/sports"},
		{"menu_item_id": 5, "language": "tr", "title": "Spor", "url": "/kategori/spor"},
		{"menu_item_id": 5, "language": "es", "title": "Deportes", "url": "/categoria/deportes"},

		// About
		{"menu_item_id": 6, "language": "en", "title": "About", "url": "/about"},
		{"menu_item_id": 6, "language": "tr", "title": "Hakkımızda", "url": "/hakkimizda"},
		{"menu_item_id": 6, "language": "es", "title": "Acerca de", "url": "/acerca-de"},

		// Contact
		{"menu_item_id": 7, "language": "en", "title": "Contact", "url": "/contact"},
		{"menu_item_id": 7, "language": "tr", "title": "İletişim", "url": "/iletisim"},
		{"menu_item_id": 7, "language": "es", "title": "Contacto", "url": "/contacto"},
	}

	for _, menuItemTrans := range menuItemTranslations {
		query := `INSERT INTO menu_item_translations (menu_item_id, language, title, url, is_active, created_at, updated_at) 
				  VALUES (:menu_item_id, :language, :title, :url, true, NOW(), NOW()) 
				  ON CONFLICT (menu_item_id, language) DO UPDATE SET 
				  	title = EXCLUDED.title, 
				  	url = EXCLUDED.url,
				  	updated_at = NOW()`
		_, err := db.NamedExec(query, menuItemTrans)
		if err != nil {
			log.Printf("Failed to insert menu item translation: %v", err)
		}
	}

	// Seed Page Translations
	pageTranslations := []map[string]interface{}{
		// About Page
		{"page_id": 1, "language": "en", "title": "About Us", "slug": "about", "content": "Learn more about our news platform and mission.", "meta_title": "About Us", "meta_desc": "Learn more about our news platform and mission", "og_title": "About Us", "og_description": "Learn more about our news platform"},
		{"page_id": 1, "language": "tr", "title": "Hakkımızda", "slug": "hakkimizda", "content": "Haber platformumuz ve misyonumuz hakkında daha fazla bilgi edinin.", "meta_title": "Hakkımızda", "meta_desc": "Haber platformumuz ve misyonumuz hakkında daha fazla bilgi edinin", "og_title": "Hakkımızda", "og_description": "Haber platformumuz hakkında daha fazla bilgi"},
		{"page_id": 1, "language": "es", "title": "Acerca de Nosotros", "slug": "acerca-de", "content": "Conoce más sobre nuestra plataforma de noticias y misión.", "meta_title": "Acerca de Nosotros", "meta_desc": "Conoce más sobre nuestra plataforma de noticias y misión", "og_title": "Acerca de Nosotros", "og_description": "Conoce más sobre nuestra plataforma de noticias"},

		// Contact Page
		{"page_id": 2, "language": "en", "title": "Contact Us", "slug": "contact", "content": "Get in touch with our team.", "meta_title": "Contact Us", "meta_desc": "Get in touch with our news team", "og_title": "Contact Us", "og_description": "Get in touch with our team"},
		{"page_id": 2, "language": "tr", "title": "İletişim", "slug": "iletisim", "content": "Ekibimizle iletişime geçin.", "meta_title": "İletişim", "meta_desc": "Haber ekibimizle iletişime geçin", "og_title": "İletişim", "og_description": "Ekibimizle iletişime geçin"},
		{"page_id": 2, "language": "es", "title": "Contáctanos", "slug": "contacto", "content": "Ponte en contacto con nuestro equipo.", "meta_title": "Contáctanos", "meta_desc": "Ponte en contacto con nuestro equipo de noticias", "og_title": "Contáctanos", "og_description": "Ponte en contacto con nuestro equipo"},

		// Privacy Policy Page
		{"page_id": 3, "language": "en", "title": "Privacy Policy", "slug": "privacy", "content": "Our privacy policy and data protection information.", "meta_title": "Privacy Policy", "meta_desc": "Our privacy policy and data protection information", "og_title": "Privacy Policy", "og_description": "Our privacy policy"},
		{"page_id": 3, "language": "tr", "title": "Gizlilik Politikası", "slug": "gizlilik", "content": "Gizlilik politikamız ve veri koruma bilgileri.", "meta_title": "Gizlilik Politikası", "meta_desc": "Gizlilik politikamız ve veri koruma bilgileri", "og_title": "Gizlilik Politikası", "og_description": "Gizlilik politikamız"},
		{"page_id": 3, "language": "es", "title": "Política de Privacidad", "slug": "privacidad", "content": "Nuestra política de privacidad e información de protección de datos.", "meta_title": "Política de Privacidad", "meta_desc": "Nuestra política de privacidad e información de protección de datos", "og_title": "Política de Privacidad", "og_description": "Nuestra política de privacidad"},
	}

	for _, pageTrans := range pageTranslations {
		query := `INSERT INTO page_translations (page_id, language, title, slug, content, meta_title, meta_desc, og_title, og_description, is_active, created_at, updated_at) 
				  VALUES (:page_id, :language, :title, :slug, :content, :meta_title, :meta_desc, :og_title, :og_description, true, NOW(), NOW()) 
				  ON CONFLICT (page_id, language) DO UPDATE SET 
				  	title = EXCLUDED.title, 
				  	slug = EXCLUDED.slug,
				  	content = EXCLUDED.content,
				  	meta_title = EXCLUDED.meta_title,
				  	meta_desc = EXCLUDED.meta_desc,
				  	og_title = EXCLUDED.og_title,
				  	og_description = EXCLUDED.og_description,
				  	updated_at = NOW()`
		_, err := db.NamedExec(query, pageTrans)
		if err != nil {
			log.Printf("Failed to insert page translation: %v", err)
		}
	}

	fmt.Printf("✅ Successfully seeded content translations:\n")
	fmt.Printf("   • %d category translations\n", len(categoryTranslations))
	fmt.Printf("   • %d tag translations\n", len(tagTranslations))
	fmt.Printf("   • %d menu translations\n", len(menuTranslations))
	fmt.Printf("   • %d menu item translations\n", len(menuItemTranslations))
	fmt.Printf("   • %d page translations\n", len(pageTranslations))

	return nil
}
