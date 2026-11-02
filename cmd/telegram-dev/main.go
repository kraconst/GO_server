package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"unicode"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"golang.org/x/term"

	"github.com/kraconst/GO_server/internal/telegram"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Telegram:", err)
		os.Exit(1)
	}
}

func run() error {
	limit := flag.Int("limit", 20, "maximum number of dialogs (1–100)")
	flag.Parse()
	if flag.NArg() != 0 {
		return errors.New("unexpected positional arguments")
	}
	if *limit < 1 || *limit > 100 {
		return errors.New("-limit must be between 1 and 100")
	}
	id, err := strconv.Atoi(os.Getenv("TELEGRAM_API_ID"))
	if err != nil || id <= 0 {
		return errors.New("set TELEGRAM_API_ID to a positive integer")
	}
	phone := strings.TrimSpace(os.Getenv("TELEGRAM_PHONE"))
	if !validPhone(phone) {
		return errors.New("set TELEGRAM_PHONE in international format, e.g. +79991234567")
	}
	path := os.Getenv("TELEGRAM_SESSION_PATH")
	if path == "" {
		path = "./var/telegram/" + strings.TrimPrefix(phone, "+") + ".session"
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return errors.New("run this command in an interactive terminal")
	}
	client, err := telegram.New(telegram.Config{APIID: id, APIHash: os.Getenv("TELEGRAM_API_HASH"), SessionPath: path})
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	profile, dialogs, err := client.Inspect(ctx, phone, &prompts{phone: phone}, *limit)
	if err != nil {
		return err
	}
	fmt.Printf("Connected: %q (@%s), ID %d\nSession: %s\n", profile.Name, profile.Username, profile.ID, path)
	fmt.Printf("Dialogs (up to %d; first page):\n", *limit)
	for _, d := range dialogs {
		fmt.Printf("%s\t%d\t%q\n", d.Kind, d.ID, d.Name)
	}
	return nil
}

func validPhone(phone string) bool {
	if len(phone) < 8 || len(phone) > 16 || phone[0] != '+' {
		return false
	}
	for _, r := range phone[1:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

type prompts struct{ phone string }

func (p *prompts) Phone(context.Context) (string, error) { return p.phone, nil }
func (p *prompts) Code(_ context.Context, code *tg.AuthSentCode) (string, error) {
	fmt.Fprintf(os.Stderr, "Telegram selected code delivery: %T\n", code.Type)
	return secret("Login code: ")
}
func (p *prompts) Password(context.Context) (string, error) { return secret("2FA password: ") }
func (p *prompts) SignUp(context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("create the Telegram account in an official app first")
}
func (p *prompts) AcceptTermsOfService(context.Context, tg.HelpTermsOfService) error {
	return errors.New("account registration is not supported; use an existing Telegram account")
}

func secret(label string) (string, error) {
	fmt.Fprint(os.Stderr, label)
	value, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	defer func() {
		for i := range value {
			value[i] = 0
		}
	}()
	if len(value) == 0 || strings.IndexFunc(string(value), unicode.IsControl) >= 0 {
		return "", errors.New("input is empty or contains control characters")
	}
	return string(value), nil
}
