package organized

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

// SeedTranslations creates multi-language translations for the website
func SeedTranslations(db *sqlx.DB) error {
	fmt.Println("🌐 Seeding translations...")

	// Common UI translations
	translations := []map[string]interface{}{
		// Navigation
		{"key": "nav.home", "language": "en", "value": "Home", "category": "navigation"},
		{"key": "nav.home", "language": "tr", "value": "Ana Sayfa", "category": "navigation"},
		{"key": "nav.home", "language": "es", "value": "Inicio", "category": "navigation"},

		{"key": "nav.about", "language": "en", "value": "About", "category": "navigation"},
		{"key": "nav.about", "language": "tr", "value": "Hakkımızda", "category": "navigation"},
		{"key": "nav.about", "language": "es", "value": "Acerca de", "category": "navigation"},

		{"key": "nav.contact", "language": "en", "value": "Contact", "category": "navigation"},
		{"key": "nav.contact", "language": "tr", "value": "İletişim", "category": "navigation"},
		{"key": "nav.contact", "language": "es", "value": "Contacto", "category": "navigation"},

		{"key": "nav.archive", "language": "en", "value": "Archive", "category": "navigation"},
		{"key": "nav.archive", "language": "tr", "value": "Arşiv", "category": "navigation"},
		{"key": "nav.archive", "language": "es", "value": "Archivo", "category": "navigation"},

		// Categories
		{"key": "category.politics", "language": "en", "value": "Politics", "category": "categories"},
		{"key": "category.politics", "language": "tr", "value": "Politika", "category": "categories"},
		{"key": "category.politics", "language": "es", "value": "Política", "category": "categories"},

		{"key": "category.technology", "language": "en", "value": "Technology", "category": "categories"},
		{"key": "category.technology", "language": "tr", "value": "Teknoloji", "category": "categories"},
		{"key": "category.technology", "language": "es", "value": "Tecnología", "category": "categories"},

		{"key": "category.business", "language": "en", "value": "Business", "category": "categories"},
		{"key": "category.business", "language": "tr", "value": "İş Dünyası", "category": "categories"},
		{"key": "category.business", "language": "es", "value": "Negocios", "category": "categories"},

		{"key": "category.sports", "language": "en", "value": "Sports", "category": "categories"},
		{"key": "category.sports", "language": "tr", "value": "Spor", "category": "categories"},
		{"key": "category.sports", "language": "es", "value": "Deportes", "category": "categories"},

		{"key": "category.health", "language": "en", "value": "Health", "category": "categories"},
		{"key": "category.health", "language": "tr", "value": "Sağlık", "category": "categories"},
		{"key": "category.health", "language": "es", "value": "Salud", "category": "categories"},

		{"key": "category.entertainment", "language": "en", "value": "Entertainment", "category": "categories"},
		{"key": "category.entertainment", "language": "tr", "value": "Eğlence", "category": "categories"},
		{"key": "category.entertainment", "language": "es", "value": "Entretenimiento", "category": "categories"},

		{"key": "category.education", "language": "en", "value": "Education", "category": "categories"},
		{"key": "category.education", "language": "tr", "value": "Eğitim", "category": "categories"},
		{"key": "category.education", "language": "es", "value": "Educación", "category": "categories"},

		{"key": "category.environment", "language": "en", "value": "Environment", "category": "categories"},
		{"key": "category.environment", "language": "tr", "value": "Çevre", "category": "categories"},
		{"key": "category.environment", "language": "es", "value": "Medio Ambiente", "category": "categories"},

		{"key": "category.travel", "language": "en", "value": "Travel", "category": "categories"},
		{"key": "category.travel", "language": "tr", "value": "Seyahat", "category": "categories"},
		{"key": "category.travel", "language": "es", "value": "Viajes", "category": "categories"},

		{"key": "category.lifestyle", "language": "en", "value": "Lifestyle", "category": "categories"},
		{"key": "category.lifestyle", "language": "tr", "value": "Yaşam Tarzı", "category": "categories"},
		{"key": "category.lifestyle", "language": "es", "value": "Estilo de Vida", "category": "categories"},

		{"key": "category.opinion", "language": "en", "value": "Opinion", "category": "categories"},
		{"key": "category.opinion", "language": "tr", "value": "Görüş", "category": "categories"},
		{"key": "category.opinion", "language": "es", "value": "Opinión", "category": "categories"},

		// Article interface
		{"key": "article.read_more", "language": "en", "value": "Read More", "category": "article"},
		{"key": "article.read_more", "language": "tr", "value": "Devamını Oku", "category": "article"},
		{"key": "article.read_more", "language": "es", "value": "Leer Más", "category": "article"},

		{"key": "article.published_on", "language": "en", "value": "Published on", "category": "article"},
		{"key": "article.published_on", "language": "tr", "value": "Yayınlanma tarihi", "category": "article"},
		{"key": "article.published_on", "language": "es", "value": "Publicado el", "category": "article"},

		{"key": "article.by_author", "language": "en", "value": "by", "category": "article"},
		{"key": "article.by_author", "language": "tr", "value": "yazan", "category": "article"},
		{"key": "article.by_author", "language": "es", "value": "por", "category": "article"},

		{"key": "article.views", "language": "en", "value": "views", "category": "article"},
		{"key": "article.views", "language": "tr", "value": "görüntülenme", "category": "article"},
		{"key": "article.views", "language": "es", "value": "vistas", "category": "article"},

		{"key": "article.likes", "language": "en", "value": "likes", "category": "article"},
		{"key": "article.likes", "language": "tr", "value": "beğeni", "category": "article"},
		{"key": "article.likes", "language": "es", "value": "me gusta", "category": "article"},

		{"key": "article.comments", "language": "en", "value": "comments", "category": "article"},
		{"key": "article.comments", "language": "tr", "value": "yorum", "category": "article"},
		{"key": "article.comments", "language": "es", "value": "comentarios", "category": "article"},

		{"key": "article.share", "language": "en", "value": "Share", "category": "article"},
		{"key": "article.share", "language": "tr", "value": "Paylaş", "category": "article"},
		{"key": "article.share", "language": "es", "value": "Compartir", "category": "article"},

		{"key": "article.tags", "language": "en", "value": "Tags", "category": "article"},
		{"key": "article.tags", "language": "tr", "value": "Etiketler", "category": "article"},
		{"key": "article.tags", "language": "es", "value": "Etiquetas", "category": "article"},

		{"key": "article.related", "language": "en", "value": "Related Articles", "category": "article"},
		{"key": "article.related", "language": "tr", "value": "İlgili Haberler", "category": "article"},
		{"key": "article.related", "language": "es", "value": "Artículos Relacionados", "category": "article"},

		// Search
		{"key": "search.placeholder", "language": "en", "value": "Search news...", "category": "search"},
		{"key": "search.placeholder", "language": "tr", "value": "Haber ara...", "category": "search"},
		{"key": "search.placeholder", "language": "es", "value": "Buscar noticias...", "category": "search"},

		{"key": "search.results", "language": "en", "value": "Search Results", "category": "search"},
		{"key": "search.results", "language": "tr", "value": "Arama Sonuçları", "category": "search"},
		{"key": "search.results", "language": "es", "value": "Resultados de Búsqueda", "category": "search"},

		{"key": "search.no_results", "language": "en", "value": "No results found", "category": "search"},
		{"key": "search.no_results", "language": "tr", "value": "Sonuç bulunamadı", "category": "search"},
		{"key": "search.no_results", "language": "es", "value": "No se encontraron resultados", "category": "search"},

		{"key": "search.filter_by", "language": "en", "value": "Filter by", "category": "search"},
		{"key": "search.filter_by", "language": "tr", "value": "Filtrele", "category": "search"},
		{"key": "search.filter_by", "language": "es", "value": "Filtrar por", "category": "search"},

		// Pagination
		{"key": "pagination.previous", "language": "en", "value": "Previous", "category": "pagination"},
		{"key": "pagination.previous", "language": "tr", "value": "Önceki", "category": "pagination"},
		{"key": "pagination.previous", "language": "es", "value": "Anterior", "category": "pagination"},

		{"key": "pagination.next", "language": "en", "value": "Next", "category": "pagination"},
		{"key": "pagination.next", "language": "tr", "value": "Sonraki", "category": "pagination"},
		{"key": "pagination.next", "language": "es", "value": "Siguiente", "category": "pagination"},

		{"key": "pagination.page", "language": "en", "value": "Page", "category": "pagination"},
		{"key": "pagination.page", "language": "tr", "value": "Sayfa", "category": "pagination"},
		{"key": "pagination.page", "language": "es", "value": "Página", "category": "pagination"},

		{"key": "pagination.of", "language": "en", "value": "of", "category": "pagination"},
		{"key": "pagination.of", "language": "tr", "value": "/", "category": "pagination"},
		{"key": "pagination.of", "language": "es", "value": "de", "category": "pagination"},

		// Forms
		{"key": "form.name", "language": "en", "value": "Name", "category": "forms"},
		{"key": "form.name", "language": "tr", "value": "İsim", "category": "forms"},
		{"key": "form.name", "language": "es", "value": "Nombre", "category": "forms"},

		{"key": "form.email", "language": "en", "value": "Email", "category": "forms"},
		{"key": "form.email", "language": "tr", "value": "E-posta", "category": "forms"},
		{"key": "form.email", "language": "es", "value": "Correo electrónico", "category": "forms"},

		{"key": "form.subject", "language": "en", "value": "Subject", "category": "forms"},
		{"key": "form.subject", "language": "tr", "value": "Konu", "category": "forms"},
		{"key": "form.subject", "language": "es", "value": "Asunto", "category": "forms"},

		{"key": "form.message", "language": "en", "value": "Message", "category": "forms"},
		{"key": "form.message", "language": "tr", "value": "Mesaj", "category": "forms"},
		{"key": "form.message", "language": "es", "value": "Mensaje", "category": "forms"},

		{"key": "form.submit", "language": "en", "value": "Submit", "category": "forms"},
		{"key": "form.submit", "language": "tr", "value": "Gönder", "category": "forms"},
		{"key": "form.submit", "language": "es", "value": "Enviar", "category": "forms"},

		{"key": "form.required", "language": "en", "value": "Required", "category": "forms"},
		{"key": "form.required", "language": "tr", "value": "Zorunlu", "category": "forms"},
		{"key": "form.required", "language": "es", "value": "Requerido", "category": "forms"},

		// Newsletter
		{"key": "newsletter.title", "language": "en", "value": "Subscribe to Newsletter", "category": "newsletter"},
		{"key": "newsletter.title", "language": "tr", "value": "Bültene Abone Ol", "category": "newsletter"},
		{"key": "newsletter.title", "language": "es", "value": "Suscribirse al Boletín", "category": "newsletter"},

		{"key": "newsletter.description", "language": "en", "value": "Get the latest news delivered to your inbox", "category": "newsletter"},
		{"key": "newsletter.description", "language": "tr", "value": "En son haberleri e-posta kutunuza alın", "category": "newsletter"},
		{"key": "newsletter.description", "language": "es", "value": "Recibe las últimas noticias en tu bandeja de entrada", "category": "newsletter"},

		{"key": "newsletter.subscribe", "language": "en", "value": "Subscribe", "category": "newsletter"},
		{"key": "newsletter.subscribe", "language": "tr", "value": "Abone Ol", "category": "newsletter"},
		{"key": "newsletter.subscribe", "language": "es", "value": "Suscribirse", "category": "newsletter"},

		{"key": "newsletter.success", "language": "en", "value": "Successfully subscribed!", "category": "newsletter"},
		{"key": "newsletter.success", "language": "tr", "value": "Başarıyla abone oldunuz!", "category": "newsletter"},
		{"key": "newsletter.success", "language": "es", "value": "¡Suscripción exitosa!", "category": "newsletter"},

		// Footer
		{"key": "footer.copyright", "language": "en", "value": "All rights reserved", "category": "footer"},
		{"key": "footer.copyright", "language": "tr", "value": "Tüm hakları saklıdır", "category": "footer"},
		{"key": "footer.copyright", "language": "es", "value": "Todos los derechos reservados", "category": "footer"},

		{"key": "footer.privacy", "language": "en", "value": "Privacy Policy", "category": "footer"},
		{"key": "footer.privacy", "language": "tr", "value": "Gizlilik Politikası", "category": "footer"},
		{"key": "footer.privacy", "language": "es", "value": "Política de Privacidad", "category": "footer"},

		{"key": "footer.terms", "language": "en", "value": "Terms of Service", "category": "footer"},
		{"key": "footer.terms", "language": "tr", "value": "Kullanım Koşulları", "category": "footer"},
		{"key": "footer.terms", "language": "es", "value": "Términos de Servicio", "category": "footer"},

		// Breaking news
		{"key": "breaking.title", "language": "en", "value": "Breaking News", "category": "breaking"},
		{"key": "breaking.title", "language": "tr", "value": "Son Dakika", "category": "breaking"},
		{"key": "breaking.title", "language": "es", "value": "Noticias de Última Hora", "category": "breaking"},

		{"key": "breaking.urgent", "language": "en", "value": "Urgent", "category": "breaking"},
		{"key": "breaking.urgent", "language": "tr", "value": "Acil", "category": "breaking"},
		{"key": "breaking.urgent", "language": "es", "value": "Urgente", "category": "breaking"},

		// Time formats
		{"key": "time.now", "language": "en", "value": "now", "category": "time"},
		{"key": "time.now", "language": "tr", "value": "şimdi", "category": "time"},
		{"key": "time.now", "language": "es", "value": "ahora", "category": "time"},

		{"key": "time.minute_ago", "language": "en", "value": "minute ago", "category": "time"},
		{"key": "time.minute_ago", "language": "tr", "value": "dakika önce", "category": "time"},
		{"key": "time.minute_ago", "language": "es", "value": "minuto atrás", "category": "time"},

		{"key": "time.minutes_ago", "language": "en", "value": "minutes ago", "category": "time"},
		{"key": "time.minutes_ago", "language": "tr", "value": "dakika önce", "category": "time"},
		{"key": "time.minutes_ago", "language": "es", "value": "minutos atrás", "category": "time"},

		{"key": "time.hour_ago", "language": "en", "value": "hour ago", "category": "time"},
		{"key": "time.hour_ago", "language": "tr", "value": "saat önce", "category": "time"},
		{"key": "time.hour_ago", "language": "es", "value": "hora atrás", "category": "time"},

		{"key": "time.hours_ago", "language": "en", "value": "hours ago", "category": "time"},
		{"key": "time.hours_ago", "language": "tr", "value": "saat önce", "category": "time"},
		{"key": "time.hours_ago", "language": "es", "value": "horas atrás", "category": "time"},

		{"key": "time.day_ago", "language": "en", "value": "day ago", "category": "time"},
		{"key": "time.day_ago", "language": "tr", "value": "gün önce", "category": "time"},
		{"key": "time.day_ago", "language": "es", "value": "día atrás", "category": "time"},

		{"key": "time.days_ago", "language": "en", "value": "days ago", "category": "time"},
		{"key": "time.days_ago", "language": "tr", "value": "gün önce", "category": "time"},
		{"key": "time.days_ago", "language": "es", "value": "días atrás", "category": "time"},

		// Social sharing
		{"key": "social.share_on", "language": "en", "value": "Share on", "category": "social"},
		{"key": "social.share_on", "language": "tr", "value": "Paylaş:", "category": "social"},
		{"key": "social.share_on", "language": "es", "value": "Compartir en", "category": "social"},

		{"key": "social.facebook", "language": "en", "value": "Facebook", "category": "social"},
		{"key": "social.facebook", "language": "tr", "value": "Facebook", "category": "social"},
		{"key": "social.facebook", "language": "es", "value": "Facebook", "category": "social"},

		{"key": "social.twitter", "language": "en", "value": "Twitter", "category": "social"},
		{"key": "social.twitter", "language": "tr", "value": "Twitter", "category": "social"},
		{"key": "social.twitter", "language": "es", "value": "Twitter", "category": "social"},

		{"key": "social.linkedin", "language": "en", "value": "LinkedIn", "category": "social"},
		{"key": "social.linkedin", "language": "tr", "value": "LinkedIn", "category": "social"},
		{"key": "social.linkedin", "language": "es", "value": "LinkedIn", "category": "social"},

		{"key": "social.whatsapp", "language": "en", "value": "WhatsApp", "category": "social"},
		{"key": "social.whatsapp", "language": "tr", "value": "WhatsApp", "category": "social"},
		{"key": "social.whatsapp", "language": "es", "value": "WhatsApp", "category": "social"},

		// Error messages
		{"key": "error.404_title", "language": "en", "value": "Page Not Found", "category": "errors"},
		{"key": "error.404_title", "language": "tr", "value": "Sayfa Bulunamadı", "category": "errors"},
		{"key": "error.404_title", "language": "es", "value": "Página No Encontrada", "category": "errors"},

		{"key": "error.404_message", "language": "en", "value": "The page you're looking for doesn't exist", "category": "errors"},
		{"key": "error.404_message", "language": "tr", "value": "Aradığınız sayfa mevcut değil", "category": "errors"},
		{"key": "error.404_message", "language": "es", "value": "La página que buscas no existe", "category": "errors"},

		{"key": "error.500_title", "language": "en", "value": "Server Error", "category": "errors"},
		{"key": "error.500_title", "language": "tr", "value": "Sunucu Hatası", "category": "errors"},
		{"key": "error.500_title", "language": "es", "value": "Error del Servidor", "category": "errors"},

		{"key": "error.500_message", "language": "en", "value": "Something went wrong on our end", "category": "errors"},
		{"key": "error.500_message", "language": "tr", "value": "Bizim tarafımızda bir hata oluştu", "category": "errors"},
		{"key": "error.500_message", "language": "es", "value": "Algo salió mal de nuestro lado", "category": "errors"},

		{"key": "error.go_home", "language": "en", "value": "Go Home", "category": "errors"},
		{"key": "error.go_home", "language": "tr", "value": "Ana Sayfaya Git", "category": "errors"},
		{"key": "error.go_home", "language": "es", "value": "Ir al Inicio", "category": "errors"},

		// Loading and status
		{"key": "status.loading", "language": "en", "value": "Loading...", "category": "status"},
		{"key": "status.loading", "language": "tr", "value": "Yükleniyor...", "category": "status"},
		{"key": "status.loading", "language": "es", "value": "Cargando...", "category": "status"},

		{"key": "status.no_content", "language": "en", "value": "No content available", "category": "status"},
		{"key": "status.no_content", "language": "tr", "value": "İçerik bulunmuyor", "category": "status"},
		{"key": "status.no_content", "language": "es", "value": "No hay contenido disponible", "category": "status"},

		{"key": "status.try_again", "language": "en", "value": "Try Again", "category": "status"},
		{"key": "status.try_again", "language": "tr", "value": "Tekrar Dene", "category": "status"},
		{"key": "status.try_again", "language": "es", "value": "Inténtalo de Nuevo", "category": "status"},

		// User actions
		{"key": "action.login", "language": "en", "value": "Login", "category": "actions"},
		{"key": "action.login", "language": "tr", "value": "Giriş Yap", "category": "actions"},
		{"key": "action.login", "language": "es", "value": "Iniciar Sesión", "category": "actions"},

		{"key": "action.logout", "language": "en", "value": "Logout", "category": "actions"},
		{"key": "action.logout", "language": "tr", "value": "Çıkış Yap", "category": "actions"},
		{"key": "action.logout", "language": "es", "value": "Cerrar Sesión", "category": "actions"},

		{"key": "action.register", "language": "en", "value": "Register", "category": "actions"},
		{"key": "action.register", "language": "tr", "value": "Kayıt Ol", "category": "actions"},
		{"key": "action.register", "language": "es", "value": "Registrarse", "category": "actions"},

		{"key": "action.save", "language": "en", "value": "Save", "category": "actions"},
		{"key": "action.save", "language": "tr", "value": "Kaydet", "category": "actions"},
		{"key": "action.save", "language": "es", "value": "Guardar", "category": "actions"},

		{"key": "action.cancel", "language": "en", "value": "Cancel", "category": "actions"},
		{"key": "action.cancel", "language": "tr", "value": "İptal", "category": "actions"},
		{"key": "action.cancel", "language": "es", "value": "Cancelar", "category": "actions"},

		{"key": "action.edit", "language": "en", "value": "Edit", "category": "actions"},
		{"key": "action.edit", "language": "tr", "value": "Düzenle", "category": "actions"},
		{"key": "action.edit", "language": "es", "value": "Editar", "category": "actions"},

		{"key": "action.delete", "language": "en", "value": "Delete", "category": "actions"},
		{"key": "action.delete", "language": "tr", "value": "Sil", "category": "actions"},
		{"key": "action.delete", "language": "es", "value": "Eliminar", "category": "actions"},
	}

	for _, translation := range translations {
		query := `
			INSERT INTO translations (
				key, language, value, category, created_at, updated_at
			) VALUES (
				:key, :language, :value, :category, NOW(), NOW()
			) ON CONFLICT (key, language) DO UPDATE SET
				value = :value,
				category = :category,
				updated_at = NOW()`

		stmt, err := db.PrepareNamed(query)
		if err != nil {
			return fmt.Errorf("error preparing translation query: %v", err)
		}
		defer func() {
			if err := stmt.Close(); err != nil {
				log.Printf("Warning: Failed to close prepared statement: %v", err)
			}
		}()

		_, err = stmt.Exec(translation)
		if err != nil {
			return fmt.Errorf("error inserting translation '%s' (%s): %v",
				translation["key"], translation["language"], err)
		}
	}

	// Count translations per language
	var counts []struct {
		Language string `db:"language"`
		Count    int    `db:"count"`
	}

	err := db.Select(&counts, `
		SELECT language, COUNT(*) as count 
		FROM translations 
		GROUP BY language 
		ORDER BY language
	`)
	if err == nil {
		fmt.Println("   Translation counts:")
		for _, count := range counts {
			fmt.Printf("     • %s: %d translations\n", count.Language, count.Count)
		}
	}

	fmt.Printf("✅ Successfully seeded %d translations across 3 languages\n", len(translations))
	return nil
}
