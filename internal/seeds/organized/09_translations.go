package organized

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

// SeedTranslations creates multi-language translations for the website
func SeedTranslations(db *sqlx.DB) error {
	fmt.Println("🌐 Seeding translations...")

	// First, seed UI Translation Categories
	categories := []map[string]interface{}{
		{"key": "navigation", "name": "Navigation", "description": "Menu items, navigation links", "sort_order": 1},
		{"key": "buttons", "name": "Buttons", "description": "Action buttons, links", "sort_order": 2},
		{"key": "forms", "name": "Forms", "description": "Form labels, placeholders, validation messages", "sort_order": 3},
		{"key": "messages", "name": "Messages", "description": "Success, error, info messages", "sort_order": 4},
		{"key": "content", "name": "Content", "description": "Content-related labels and text", "sort_order": 5},
		{"key": "auth", "name": "Authentication", "description": "Login, registration, password reset", "sort_order": 6},
		{"key": "admin", "name": "Admin Panel", "description": "Admin interface translations", "sort_order": 7},
		{"key": "search", "name": "Search", "description": "Search-related text", "sort_order": 8},
		{"key": "pagination", "name": "Pagination", "description": "Page navigation text", "sort_order": 9},
		{"key": "date_time", "name": "Date & Time", "description": "Date and time formats", "sort_order": 10},
		{"key": "social", "name": "Social Media", "description": "Social sharing text", "sort_order": 11},
		{"key": "comments", "name": "Comments", "description": "Comment system text", "sort_order": 12},
		{"key": "newsletter", "name": "Newsletter", "description": "Newsletter subscription text", "sort_order": 13},
		{"key": "footer", "name": "Footer", "description": "Footer content and links", "sort_order": 14},
		{"key": "metadata", "name": "Metadata", "description": "SEO and meta descriptions", "sort_order": 15},
	}

	for _, category := range categories {
		query := `INSERT INTO ui_translation_categories (key, name, description, sort_order, is_active, created_at, updated_at) 
				  VALUES (:key, :name, :description, :sort_order, true, NOW(), NOW()) 
				  ON CONFLICT (key) DO UPDATE SET 
				  	name = EXCLUDED.name, 
				  	description = EXCLUDED.description, 
				  	sort_order = EXCLUDED.sort_order,
				  	updated_at = NOW()`
		_, err := db.NamedExec(query, category)
		if err != nil {
			log.Printf("Failed to insert UI translation category: %v", err)
		}
	}

	// Seed Error Message Translations
	errorMessages := []map[string]interface{}{
		{"error_code": "validation_required", "language": "en", "title": "Required Field", "message": "This field is required", "user_message": "Please fill in this required field", "category": "validation"},
		{"error_code": "validation_required", "language": "tr", "title": "Zorunlu Alan", "message": "Bu alan zorunludur", "user_message": "Lütfen bu zorunlu alanı doldurun", "category": "validation"},
		{"error_code": "validation_required", "language": "es", "title": "Campo Requerido", "message": "Este campo es obligatorio", "user_message": "Por favor complete este campo obligatorio", "category": "validation"},

		{"error_code": "validation_email", "language": "en", "title": "Invalid Email", "message": "Please enter a valid email address", "user_message": "Please enter a valid email address", "category": "validation"},
		{"error_code": "validation_email", "language": "tr", "title": "Geçersiz Email", "message": "Geçerli bir email adresi girin", "user_message": "Geçerli bir email adresi girin", "category": "validation"},
		{"error_code": "validation_email", "language": "es", "title": "Email Inválido", "message": "Ingrese una dirección de email válida", "user_message": "Ingrese una dirección de email válida", "category": "validation"},

		{"error_code": "auth_failed", "language": "en", "title": "Authentication Failed", "message": "Invalid credentials provided", "user_message": "Invalid username or password", "category": "authentication"},
		{"error_code": "auth_failed", "language": "tr", "title": "Giriş Başarısız", "message": "Geçersiz kimlik bilgileri", "user_message": "Geçersiz kullanıcı adı veya şifre", "category": "authentication"},
		{"error_code": "auth_failed", "language": "es", "title": "Autenticación Fallida", "message": "Credenciales inválidas", "user_message": "Usuario o contraseña inválidos", "category": "authentication"},

		{"error_code": "server_error", "language": "en", "title": "Server Error", "message": "An internal server error occurred", "user_message": "Something went wrong. Please try again later.", "category": "system"},
		{"error_code": "server_error", "language": "tr", "title": "Sunucu Hatası", "message": "Sunucu hatası oluştu", "user_message": "Bir şeyler ters gitti. Lütfen daha sonra tekrar deneyin.", "category": "system"},
		{"error_code": "server_error", "language": "es", "title": "Error del Servidor", "message": "Ocurrió un error interno del servidor", "user_message": "Algo salió mal. Inténtelo de nuevo más tarde.", "category": "system"},
	}

	for _, errorMsg := range errorMessages {
		query := `INSERT INTO error_message_translations (error_code, language, title, message, user_message, category, is_active, created_at, updated_at) 
				  VALUES (:error_code, :language, :title, :message, :user_message, :category, true, NOW(), NOW()) 
				  ON CONFLICT (error_code, language) DO UPDATE SET 
				  	title = EXCLUDED.title,
				  	message = EXCLUDED.message, 
				  	user_message = EXCLUDED.user_message,
				  	updated_at = NOW()`
		_, err := db.NamedExec(query, errorMsg)
		if err != nil {
			log.Printf("Failed to insert error message translation: %v", err)
		}
	}

	// Seed Form Translations
	formTranslations := []map[string]interface{}{
		// Contact Form
		{"form_key": "contact", "field_key": "name", "language": "en", "label": "Full Name", "placeholder": "Enter your full name", "help_text": "Please provide your first and last name"},
		{"form_key": "contact", "field_key": "name", "language": "tr", "label": "Ad Soyad", "placeholder": "Adınızı ve soyadınızı girin", "help_text": "Lütfen adınızı ve soyadınızı belirtin"},
		{"form_key": "contact", "field_key": "name", "language": "es", "label": "Nombre Completo", "placeholder": "Ingrese su nombre completo", "help_text": "Proporcione su nombre y apellido"},

		{"form_key": "contact", "field_key": "email", "language": "en", "label": "Email Address", "placeholder": "Enter your email address", "help_text": "We will use this to contact you"},
		{"form_key": "contact", "field_key": "email", "language": "tr", "label": "Email Adresi", "placeholder": "Email adresinizi girin", "help_text": "Sizinle iletişime geçmek için kullanacağız"},
		{"form_key": "contact", "field_key": "email", "language": "es", "label": "Dirección de Email", "placeholder": "Ingrese su dirección de email", "help_text": "Usaremos esto için contactarlo"},

		{"form_key": "contact", "field_key": "subject", "language": "en", "label": "Subject", "placeholder": "Enter the subject", "help_text": "Brief description of your inquiry"},
		{"form_key": "contact", "field_key": "subject", "language": "tr", "label": "Konu", "placeholder": "Konuyu girin", "help_text": "Sorgunuzun kısa açıklaması"},
		{"form_key": "contact", "field_key": "subject", "language": "es", "label": "Asunto", "placeholder": "Ingrese el asunto", "help_text": "Breve descripción de su consulta"},

		{"form_key": "contact", "field_key": "message", "language": "en", "label": "Message", "placeholder": "Enter your message", "help_text": "Provide details about your inquiry"},
		{"form_key": "contact", "field_key": "message", "language": "tr", "label": "Mesaj", "placeholder": "Mesajınızı girin", "help_text": "Sorgunuz hakkında detay verin"},
		{"form_key": "contact", "field_key": "message", "language": "es", "label": "Mensaje", "placeholder": "Ingrese su mensaje", "help_text": "Proporcione detalles sobre su consulta"},

		// Newsletter Form
		{"form_key": "newsletter", "field_key": "email", "language": "en", "label": "Email Address", "placeholder": "Enter your email", "help_text": "Subscribe to our newsletter"},
		{"form_key": "newsletter", "field_key": "email", "language": "tr", "label": "Email Adresi", "placeholder": "Email adresinizi girin", "help_text": "Bültenimize abone olun"},
		{"form_key": "newsletter", "field_key": "email", "language": "es", "label": "Dirección de Email", "placeholder": "Ingrese su email", "help_text": "Suscríbase a nuestro boletín"},

		// Login Form
		{"form_key": "login", "field_key": "email", "language": "en", "label": "Email", "placeholder": "Enter your email", "help_text": "Your registered email address"},
		{"form_key": "login", "field_key": "email", "language": "tr", "label": "Email", "placeholder": "Email adresinizi girin", "help_text": "Kayıtlı email adresiniz"},
		{"form_key": "login", "field_key": "email", "language": "es", "label": "Email", "placeholder": "Ingrese su email", "help_text": "Su dirección de email registrada"},

		{"form_key": "login", "field_key": "password", "language": "en", "label": "Password", "placeholder": "Enter your password", "help_text": "Your account password"},
		{"form_key": "login", "field_key": "password", "language": "tr", "label": "Şifre", "placeholder": "Şifrenizi girin", "help_text": "Hesap şifreniz"},
		{"form_key": "login", "field_key": "password", "language": "es", "label": "Contraseña", "placeholder": "Ingrese su contraseña", "help_text": "Su contraseña de cuenta"},
	}

	for _, formTrans := range formTranslations {
		query := `INSERT INTO form_translations (form_key, field_key, language, label, placeholder, help_text, is_active, created_at, updated_at) 
				  VALUES (:form_key, :field_key, :language, :label, :placeholder, :help_text, true, NOW(), NOW()) 
				  ON CONFLICT (form_key, field_key, language) DO UPDATE SET 
				  	label = EXCLUDED.label, 
				  	placeholder = EXCLUDED.placeholder, 
				  	help_text = EXCLUDED.help_text,
				  	updated_at = NOW()`
		_, err := db.NamedExec(query, formTrans)
		if err != nil {
			log.Printf("Failed to insert form translation: %v", err)
		}
	}

	// Seed Email Template Translations
	emailTemplates := []map[string]interface{}{
		{"template_key": "welcome", "language": "en", "subject": "Welcome to Our News Platform", "plain_body": "Welcome to our news platform! Thank you for joining us.", "html_body": "<h1>Welcome!</h1><p>Thank you for joining our news platform.</p>", "preheader_text": "Welcome to our community"},
		{"template_key": "welcome", "language": "tr", "subject": "Haber Platformumuza Hoş Geldiniz", "plain_body": "Haber platformumuza hoş geldiniz! Bize katıldığınız için teşekkürler.", "html_body": "<h1>Hoş Geldiniz!</h1><p>Haber platformumuza katıldığınız için teşekkürler.</p>", "preheader_text": "Topluluğumuza hoş geldiniz"},
		{"template_key": "welcome", "language": "es", "subject": "Bienvenido a Nuestra Plataforma de Noticias", "plain_body": "¡Bienvenido a nuestra plataforma de noticias! Gracias por unirte a nosotros.", "html_body": "<h1>¡Bienvenido!</h1><p>Gracias por unirte a nuestra plataforma de noticias.</p>", "preheader_text": "Bienvenido a nuestra comunidad"},

		{"template_key": "password_reset", "language": "en", "subject": "Password Reset Request", "plain_body": "You requested a password reset. Click the link to reset your password.", "html_body": "<h1>Password Reset</h1><p>Click the link below to reset your password.</p>", "preheader_text": "Reset your password"},
		{"template_key": "password_reset", "language": "tr", "subject": "Şifre Sıfırlama Talebi", "plain_body": "Şifre sıfırlama talebinde bulundunuz. Şifrenizi sıfırlamak için bağlantıya tıklayın.", "html_body": "<h1>Şifre Sıfırlama</h1><p>Şifrenizi sıfırlamak için aşağıdaki bağlantıya tıklayın.</p>", "preheader_text": "Şifrenizi sıfırlayın"},
		{"template_key": "password_reset", "language": "es", "subject": "Solicitud de Restablecimiento de Contraseña", "plain_body": "Solicitó restablecer su contraseña. Haga clic en el enlace para restablecerla.", "html_body": "<h1>Restablecimiento de Contraseña</h1><p>Haga clic en el enlace a continuación para restablecer su contraseña.</p>", "preheader_text": "Restablezca su contraseña"},

		{"template_key": "newsletter", "language": "en", "subject": "Weekly Newsletter", "plain_body": "Here are this week's top stories.", "html_body": "<h1>Weekly Newsletter</h1><p>Here are this week's top stories.</p>", "preheader_text": "Your weekly news digest"},
		{"template_key": "newsletter", "language": "tr", "subject": "Haftalık Bülten", "plain_body": "Bu haftanın en önemli haberleri.", "html_body": "<h1>Haftalık Bülten</h1><p>Bu haftanın en önemli haberleri.</p>", "preheader_text": "Haftalık haber özetiniz"},
		{"template_key": "newsletter", "language": "es", "subject": "Boletín Semanal", "plain_body": "Aquí están las principales noticias de esta semana.", "html_body": "<h1>Boletín Semanal</h1><p>Aquí están las principales noticias de esta semana.</p>", "preheader_text": "Su resumen semanal de noticias"},
	}

	for _, emailTemplate := range emailTemplates {
		query := `INSERT INTO email_template_translations (template_key, language, subject, plain_body, html_body, preheader_text, is_active, created_at, updated_at) 
				  VALUES (:template_key, :language, :subject, :plain_body, :html_body, :preheader_text, true, NOW(), NOW()) 
				  ON CONFLICT (template_key, language) DO UPDATE SET 
				  	subject = EXCLUDED.subject, 
				  	plain_body = EXCLUDED.plain_body, 
				  	html_body = EXCLUDED.html_body,
				  	preheader_text = EXCLUDED.preheader_text,
				  	updated_at = NOW()`
		_, err := db.NamedExec(query, emailTemplate)
		if err != nil {
			log.Printf("Failed to insert email template translation: %v", err)
		}
	}

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
