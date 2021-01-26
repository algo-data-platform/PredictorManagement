package mail

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

type LoginAuth struct {
	username, password string
}

func NewLoginAuth(username, password string) smtp.Auth {
	return &LoginAuth{username, password}
}

func (a *LoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *LoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unknown fromServer")
		}
	}
	return nil, nil
}

type Mail struct {
	Auth smtp.Auth
	Addr string
}

// 初始化邮箱对象
func New(addr string, username, password string) *Mail {
	auth := NewLoginAuth(username, password)
	return &Mail{
		Auth: auth,
		Addr: addr,
	}
}

// 邮件发送
func (mail *Mail) SendMail(from_nickname, from string, to []string, subject string, msg []byte) error {
	c, err := smtp.Dial(mail.Addr)
	host, _, _ := net.SplitHostPort(mail.Addr)
	if err != nil {
		return fmt.Errorf("call dial err: %v", err)
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: host, InsecureSkipVerify: true}
		if err = c.StartTLS(config); err != nil {
			return fmt.Errorf("call start tls, err: %v", err)
		}
	}

	if mail.Auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(mail.Auth); err != nil {
				return fmt.Errorf("check auth with err: %v", err)
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}

	header := make(map[string]string)
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/html; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"
	header["From"] = from_nickname
	header["To"] = strings.Join(to, ",")
	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString(msg)
	_, err = w.Write([]byte(message))

	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
