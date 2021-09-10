package drpv4

/*
 * Copyright RackN 2020
 */

import (
	"fmt"
	"log"

	"gitlab.com/rackn/provision/v4/api"
)

type Config struct {
	token    string
	username string
	password string
	endpoint string

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
	if c.token != "" {
		c.session, err = api.TokenSession(c.endpoint, c.token)
	} else {
		c.session, err = api.UserSession(c.endpoint, c.username, c.password)
	}
	if err != nil {
		log.Printf("[ERROR] Error creating session: %+v", err)
		return fmt.Errorf("Error creating session: %s", err)
	} else {
		log.Printf("[DEBUG] [Condig.validateAndConnect] Authenticated! Session Starting w %+v", c)
	}

	return nil
}
