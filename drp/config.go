package drp

import (
	"log"

	"gitlab.com/rackn/provision/v4/api"
)

type Config struct {
	Token    string
	Username string
	Password string
	Url      string

	session *api.Client
}

/*
 * Builds a client object for this config
 */
func (c *Config) validateAndConnect() error {
	log.Println("[DEBUG] [Config.validateAndConnect] Configuring the DRP API client")

	if c.session != nil {
		return nil
	}
	var err error
	if c.Token != "" {
		c.session, err = api.TokenSession(c.Url, c.Token)
	} else {
		c.session, err = api.UserSession(c.Url, c.Username, c.Password)
	}
	if err != nil {
		log.Printf("[ERROR] Error creating session: %v", err)
		return err
	}

	return nil
}
