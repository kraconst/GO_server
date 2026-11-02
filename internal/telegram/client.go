// Package telegram provides the MTProto connection used by the development CLI.
package telegram

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gotd/td/session"
	gotd "github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
)

type Config struct {
	APIID       int
	APIHash     string
	SessionPath string
}

// Client owns one account's connection. Use a separate session path per account.
type Client struct{ client *gotd.Client }

func New(cfg Config) (*Client, error) {
	if cfg.APIID <= 0 || strings.TrimSpace(cfg.APIHash) == "" {
		return nil, fmt.Errorf("TELEGRAM_API_ID must be positive and TELEGRAM_API_HASH must be set")
	}
	if strings.TrimSpace(cfg.SessionPath) == "" {
		return nil, fmt.Errorf("session path must be set")
	}
	if err := os.MkdirAll(filepath.Dir(cfg.SessionPath), 0700); err != nil {
		return nil, fmt.Errorf("create session directory: %w", err)
	}
	return &Client{client: gotd.NewClient(cfg.APIID, cfg.APIHash, gotd.Options{
		SessionStorage: &session.FileStorage{Path: cfg.SessionPath},
		NoUpdates:      true,
	})}, nil
}

type Profile struct {
	ID       int64
	Name     string
	Username string
}

type Dialog struct {
	ID   int64
	Kind string
	Name string
}

// Inspect restores a session or signs in, then fetches one bounded page of dialogs.
// It closes the connection on return and never logs out the saved session.
func (c *Client) Inspect(ctx context.Context, phone string, authenticator auth.UserAuthenticator, limit int) (Profile, []Dialog, error) {
	var profile Profile
	var dialogs []Dialog
	if limit < 1 || limit > 100 {
		return profile, nil, fmt.Errorf("dialog limit must be between 1 and 100")
	}
	err := c.client.Run(ctx, func(ctx context.Context) error {
		if err := c.client.Auth().IfNecessary(ctx, auth.NewFlow(authenticator, auth.SendCodeOptions{})); err != nil {
			return fmt.Errorf("authorize Telegram account: %w", err)
		}
		self, err := c.client.Self(ctx)
		if err != nil {
			return fmt.Errorf("get own profile: %w", err)
		}
		// Prevent accidentally inspecting another account's saved session.
		if self.Phone != strings.TrimPrefix(phone, "+") {
			return fmt.Errorf("saved session belongs to another phone; use a separate TELEGRAM_SESSION_PATH")
		}
		profile = Profile{ID: self.ID, Name: strings.TrimSpace(self.FirstName + " " + self.LastName), Username: self.Username}
		result, err := c.client.API().MessagesGetDialogs(ctx, &tg.MessagesGetDialogsRequest{OffsetPeer: &tg.InputPeerEmpty{}, Limit: limit})
		if err != nil {
			return fmt.Errorf("get dialogs: %w", err)
		}
		var items []tg.DialogClass
		var users []tg.UserClass
		var chats []tg.ChatClass
		switch r := result.(type) {
		case *tg.MessagesDialogs:
			items, users, chats = r.Dialogs, r.Users, r.Chats
		case *tg.MessagesDialogsSlice:
			items, users, chats = r.Dialogs, r.Users, r.Chats
		default:
			return fmt.Errorf("unexpected dialogs response: %T", result)
		}
		names := make(map[int64]string)
		for _, u := range users {
			if u, ok := u.(*tg.User); ok {
				names[u.ID] = strings.TrimSpace(u.FirstName + " " + u.LastName)
			}
		}
		chatNames := make(map[int64]string)
		for _, chat := range chats {
			switch chat := chat.(type) {
			case *tg.Chat:
				chatNames[chat.ID] = chat.Title
			case *tg.Channel:
				chatNames[chat.ID] = chat.Title
			}
		}
		for _, item := range items {
			d, ok := item.(*tg.Dialog)
			if !ok {
				continue
			}
			var dialog Dialog
			switch p := d.Peer.(type) {
			case *tg.PeerUser:
				dialog = Dialog{ID: p.UserID, Kind: "user", Name: names[p.UserID]}
			case *tg.PeerChat:
				dialog = Dialog{ID: p.ChatID, Kind: "group", Name: chatNames[p.ChatID]}
			case *tg.PeerChannel:
				dialog = Dialog{ID: p.ChannelID, Kind: "channel", Name: chatNames[p.ChannelID]}
			default:
				continue
			}
			dialogs = append(dialogs, dialog)
		}
		return nil
	})
	return profile, dialogs, err
}
