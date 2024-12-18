package httpserver

import (
	"app/pkg/auth"
	"app/pkg/config"
	"app/pkg/emailsender"
	"app/pkg/repository"
	"app/pkg/repository/postgres"
	"net/http"

	"github.com/gorilla/mux"
)

func StartHttpServer(port string) error {
	var DB repository.Storage
	var err error

	dbConfig := postgres.PostgresConfig{
		Host:     config.GetEnv("DB_HOST", "localhost"),
		Port:     config.GetEnv("DB_PORT", "5432"),
		Database: config.GetEnv("DB_NAME", "postgres"),
		User:     config.GetEnv("DB_USER", "postgres"),
		Password: config.GetEnv("DB_PASSWORD", "postgres"),
		Sslmode:  config.GetEnv("DB_SSLMODE", "disable"),
	}

	DB, err = postgres.NewPostgresConnect(&dbConfig)
	if err != nil {
		return err
	}
	defer DB.Close()

	EmailSender := emailsender.NewEmailSender(&emailsender.AuthData{
		Identity: config.GetEnv("EMAIL_SENDER_IDENTITY", "identity"),
		Username: config.GetEnv("EMAIL_SENDER_USERNAME", "username"),
		Password: config.GetEnv("EMAIL_SENDER_PASSWORD", "password"),
		Host:     config.GetEnv("EMAIL_SENDER_HOST", "host"),
		Email:    config.GetEnv("EMAIL_SENDER_EMAIL", "email"),
		Addr:     config.GetEnv("EMAIL_SENDER_ADDR", "addr"),
	})

	r := mux.NewRouter()

	r.HandleFunc("/Auth", auth.GetTokenHandler(DB, config.GetEnv("SECRET_KEY", "secret"))).Methods("GET")
	r.HandleFunc("/Refresh", auth.RefreshHandler(DB, EmailSender, config.GetEnv("SECRET_KEY", "secret"))).Methods("POST")

	return http.ListenAndServe(port, r)

}
