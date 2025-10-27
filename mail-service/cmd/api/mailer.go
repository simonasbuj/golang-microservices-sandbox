package main

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain     string
	Host       string
	Port       int
	Username   string
	Password   string
	Encryption string
	FromName   string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
}

func (m *Mail) SendSMTPMessage(msg Message) error {
	if msg.From == "" {
		msg.From = m.FromName
	}

	msg.DataMap = map[string]any{
		"message": msg.Data,
	}

	htmlMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to build html message: %w", err)
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to build plain text message: %w", err)
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to smtp client: %w", err)
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject)

	email.SetBody(mail.TextPlain, plainMessage)
	email.AddAlternative(mail.TextHTML, htmlMessage)

	if len(msg.Attachments) > 0 {
		for _, a := range msg.Attachments {
			email.AddAttachment(a)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		return fmt.Errorf("failed to send an email; %w", err)
	}

	return nil
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := "./templates/mail.gohtml"

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %q, %w", templateToRender, err)
	}

	var tpl bytes.Buffer
	err = t.ExecuteTemplate(&tpl, "body", msg.DataMap)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", fmt.Errorf("failed to inline css: %w", err)
	}

	return formattedMessage, nil
}

func (m *Mail) inlineCSS(msg string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(msg, &options)
	if err != nil {
		return "", fmt.Errorf("failed to create premailer from string: %w", err)
	}

	html, err := prem.Transform()
	if err != nil {
		return "", fmt.Errorf("failed to transform premailer: %w", err)
	}

	return html, nil
}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	templateToRender := "./templates/mail.plain.gohtml"

	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %q, %w", templateToRender, err)
	}

	var tpl bytes.Buffer
	err = t.ExecuteTemplate(&tpl, "body", msg.DataMap)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

func (m *Mail) getEncryption(encryption string) mail.Encryption {
	switch encryption {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none", "":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
