// Package app contains methods for launching shortener service.
package app

import (
	"context"
	"database/sql"
	"gophkeeper/internal/config"
	"gophkeeper/internal/delivery/http"
	"gophkeeper/internal/server"
	"gophkeeper/internal/service"
	"gophkeeper/internal/storage"
	"gophkeeper/pkg/auth"
	"gophkeeper/pkg/hash"
	"log"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	var cfg config.Config
	err := cfg.Parse()
	if err != nil {
		log.Fatal(err)
	}

	db, err := newInPSQL(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}

	storages := storage.NewStorages(db)

	hasher := hash.NewSHA1Hasher(cfg.PasswordSalt)
	tokenManager, err := auth.NewManager(cfg.PasswordSalt)
	if err != nil {
		log.Fatal(err)
	}

	deps := service.Deps{
		Storages: storages,
		Hasher: hasher,
		TokenManager: tokenManager,
		AccessTokenTTL: 1 * time.Minute,
		RefreshTokenTTL: 40 * 24 * time.Hour,
	}

	services := service.NewServices(deps)

	// HTTP server
	handlers := http.NewHandler(services, tokenManager)
	httpSrv := server.NewServer(&cfg, handlers.InitEcho())

	connectionsClosed := make(chan struct{})
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-interrupt
		if err := httpSrv.Stop(context.Background()); err != nil {
			log.Printf("HTTP server shutdown: %v", err)
		}

		if err := storages.Users.Close(); err != nil {
			log.Printf("Storage shutdown error: %v", err)
		}

		close(connectionsClosed)
	}()

	go func() {
		if err := httpSrv.Run(); err != nethttp.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	<-connectionsClosed
	log.Println("Server shutdown gracefully")
}

func newInPSQL(databaseDSN string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseDSN)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	if err = createTable(db); err != nil {
		log.Fatal(err)
	}
	return db, nil
}

func createTable(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS users (
		id serial primary key,
		login text not null unique,
		password text not null
    );
    CREATE TABLE IF NOT EXISTS sessions (
    	refresh_token text primary key unique,
		user_id int not null references users(id),
		expired_at timestamp
    );
	CREATE TABLE IF NOT EXISTS auth_data (
		id serial primary key,
		user_id int not null references users(id),
		login text not null,
		password text not null,
		metadata text
	);
	CREATE TABLE IF NOT EXISTS text_data (
		id serial primary key,
		user_id int not null references users(id),
		"data" text not null,
		metadata text
	);
	CREATE TABLE IF NOT EXISTS blob_data (
		id serial primary key,
		user_id int not null references users(id),
		"data" bytea not null,
		metadata text
	);
	CREATE TABLE IF NOT EXISTS card_data (
		id serial primary key,
		user_id int not null references users(id),
		card_number text not null unique,
		exp_date date not null, 
		cvv text not null,
		"name" text,
		surname text,
		metadata text
	);`
	_, err := db.Exec(query)
	return err
}
