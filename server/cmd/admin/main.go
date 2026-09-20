package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/config"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/database"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/repository"
	"github.com/RodolfoIoLo/rodolfoiolo-site/server/internal/security"
	"golang.org/x/term"
)

func main() {
	username := flag.String("username", "", "unique login username")
	nickname := flag.String("nickname", "", "display name")
	email := flag.String("email", "", "email address")
	role := flag.String("role", "admin", "admin or editor")
	flag.Parse()

	if strings.TrimSpace(*username) == "" || strings.TrimSpace(*nickname) == "" || (*role != "admin" && *role != "editor") {
		fmt.Fprintln(os.Stderr, "username, nickname, and a valid role are required")
		os.Exit(2)
	}
	fmt.Fprint(os.Stderr, "Password: ")
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		fail(err)
	}
	hash, err := security.HashPassword(string(password))
	for index := range password {
		password[index] = 0
	}
	if err != nil {
		fail(err)
	}

	cfg, err := config.Load()
	if err != nil {
		fail(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := database.Open(ctx, database.Options{URL: cfg.DatabaseURL, MaxConns: 2, MinConns: 0, Timeout: cfg.DatabaseTimeout})
	if err != nil {
		fail(err)
	}
	defer pool.Close()
	user, err := repository.NewAuthRepository(pool).CreateUser(ctx, strings.TrimSpace(*username), hash, strings.TrimSpace(*nickname), strings.TrimSpace(*email), *role)
	if err != nil {
		fail(err)
	}
	fmt.Printf("created user id=%d username=%s role=%s\n", user.ID, user.Username, user.Role)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
