package credential

import (
	"encoding/base64"
	"errors"
)

type Credential struct {
	Username string
	Password string
}

func SetCredential(username, password string) (*Credential, error) {
	if err := validateCredential(username, password); err != nil {
		return nil, err
	}
	cred := &Credential{
		Username: username,
		Password: password,
	}
	return cred, nil
}

func validateCredential(username, password string) error {
	if username == "" {
		return errors.New("username was empty")
	}
	if password == "" {
		return errors.New("password was empty")
	}
	return nil
}

func (c *Credential) GenerateToken() string {
	username := c.Username
	password := c.Password
	credential := username + ":" + password
	base64Token := base64.StdEncoding.EncodeToString([]byte(credential))
	return base64Token
}
