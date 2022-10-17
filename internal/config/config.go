package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Addr			string		`env:"RUN_ADDRESS"`
	DatabaseDSN 	string		`env:"DATABASE_URI"`
	PasswordSalt	string		`env:"PASSWORD_SALT" envDefault:"PaSsW0rD"`
}

// DatabaseDSN scheme: "postgres://username:password@localhost:5432/database_name?sslmode=disable"

func (c* Config)Parse() error {
	flag.StringVar(&c.Addr,"a", "localhost:8081", "Host to listen on")
	flag.StringVar(&c.DatabaseDSN,"d", "", "The line with the address to connect to the database")
	flag.StringVar(&c.PasswordSalt,"p", "", "Password salt to create hashes for users's passwords")
	flag.Parse()

	//settings redefinition, if env variables are used
	err := env.Parse(c)

	return err
}