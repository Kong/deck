// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"errors"
	"io/ioutil"
	"log"

	"github.com/hbagdi/doko/dump"
	"github.com/hbagdi/doko/utils"
	"github.com/hbagdi/go-kong/kong"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := kong.NewClient(nil, nil)
		if err != nil {
			log.Fatalln(err)
		}
		return d(client)
	},
}

var withID bool

func init() {
	rootCmd.AddCommand(dumpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dumpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dumpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	dumpCmd.Flags().BoolVar(&withID, "with-ids", false, "export all entities with IDs")
}

func d(client *kong.Client) error {
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

	if err := dropFields(ks, !withID, true); err != nil {
		log.Fatalln(err)
	}

	if err := outputToFile(ks); err != nil {
		log.Fatalln(err)
	}
	return nil
}

func dropFields(state *dump.KongRawState, id, ts bool) error {
	for _, s := range state.Services {
		if id {
			s.ID = nil
		}
		if ts {
			s.CreatedAt = nil
			s.UpdatedAt = nil
		}
	}

	for _, r := range state.Routes {
		if id {
			r.ID = nil
		}
		if ts {
			r.CreatedAt = nil
			r.UpdatedAt = nil
		}
	}

	for _, p := range state.Plugins {
		if id {
			p.ID = nil
		}
		if ts {
			p.CreatedAt = nil
		}
	}

	for _, c := range state.Certificates {
		if id {
			c.ID = nil
		}
		if ts {
			c.CreatedAt = nil
		}
	}

	for _, s := range state.SNIs {
		if id {
			s.ID = nil
		}
		if ts {
			s.CreatedAt = nil
		}
	}

	for _, u := range state.Upstreams {
		if id {
			u.ID = nil
		}
		if ts {
			u.CreatedAt = nil
		}
	}

	for _, t := range state.Targets {
		if id {
			t.ID = nil
		}
		if ts {
			t.CreatedAt = nil
		}
	}

	for _, c := range state.Consumers {
		if id {
			c.ID = nil
		}
		if ts {
			c.CreatedAt = nil
		}
	}
	return nil
}

func outputToFile(state *dump.KongRawState) error {
	c, err := yaml.Marshal(state)
	err = ioutil.WriteFile("out", c, 0644)
	if err != nil {
		return err
	}

	return nil
}
