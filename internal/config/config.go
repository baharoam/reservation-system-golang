package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
	"github.com/baharoam/reservation/internal/models"
)

// AppConfig holds the application config
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	InProduction bool
	Session *scs.SessionManager
	ErrorLog      *log.Logger
	MailChan	chan models.MailData
}
