package main

import (
	"errors"
	"io/ioutil"
	"log"

	"github.com/hbagdi/doko/dump"
	"github.com/hbagdi/doko/utils"
	"github.com/hbagdi/go-kong/kong"
	yaml "gopkg.in/yaml.v2"
)

var entities = []string{"key-auth", "hmac-auth", "jwt", "oauth2", "acl"}

func main() {
	client, err := kong.NewClient(nil, nil)
	if err != nil {
		log.Fatalln(err)
	}

	if err := Dump(client); err != nil {
		log.Fatalln(err)
	}
}

func Dump(client *kong.Client) error {
	ks, err := dump.Get(client)
	if err != nil {
		log.Fatalln(err)
	}

	// check if all services havea name or not
	for _, s := range ks.Services {
		if utils.Empty(s.Name) {
			return (errors.New("service with id '" + *s.ID + "' has no 'name' property." +
				" 'name' property is required if IDs are not being exported."))
		}
	}

	if err := removeTSAndIDs(ks); err != nil {
		log.Fatalln(err)
	}

	if err := outputToFile(ks); err != nil {
		log.Fatalln(err)
	}
	return nil
}

func removeTSAndIDs(state *dump.KongRawState) error {
	for _, s := range state.Services {
		s.ID = nil
		s.CreatedAt = nil
		s.UpdatedAt = nil
	}

	for _, r := range state.Routes {
		r.ID = nil
		r.CreatedAt = nil
		r.UpdatedAt = nil
	}

	for _, p := range state.Plugins {
		p.ID = nil
		p.CreatedAt = nil
	}

	for _, c := range state.Certificates {
		c.ID = nil
		c.CreatedAt = nil
	}

	for _, s := range state.SNIs {
		s.ID = nil
		s.CreatedAt = nil
	}

	for _, u := range state.Upstreams {
		u.ID = nil
		u.CreatedAt = nil
	}

	for _, t := range state.Targets {
		t.ID = nil
		t.CreatedAt = nil
	}

	for _, c := range state.Consumers {
		c.ID = nil
		c.CreatedAt = nil
	}
	return nil
}

func outputToFile(state *dump.KongRawState) error {
	c, err := yaml.Marshal(state)
	err = ioutil.WriteFile("test3.out", c, 0644)
	if err != nil {
		return err
	}

	return nil
}
