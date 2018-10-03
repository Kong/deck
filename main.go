package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hbagdi/go-kong/kong"
	yaml "gopkg.in/yaml.v2"
)

var entities = []string{"key-auth", "hmac-auth", "jwt", "oauth2", "acl"}

func main() {
	fmt.Println("vim-go")
	dumpState()
}

func dumpState() {
	client, err := kong.NewClient(nil, nil)
	if err != nil {
		log.Fatalln(err)
	}

	var b []byte

	services, err := GetAllServices(client)
	if err != nil {
		log.Fatalln(err)
	}
	c, err := yaml.Marshal(services)
	if err != nil {
		log.Fatalln(err)
	}
	b = append(b, c...)

	routes, err := GetAllRoutes(client)
	if err != nil {
		log.Fatalln(err)
	}
	c, err = yaml.Marshal(routes)
	if err != nil {
		log.Fatalln(err)
	}
	b = append(b, c...)

	plugins, err := GetAllPlugins(client)
	if err != nil {
		log.Fatalln(err)
	}
	c, err = yaml.Marshal(plugins)
	if err != nil {
		log.Fatalln(err)
	}
	b = append(b, c...)

	certificates, err := GetAllCertificates(client)
	if err != nil {
		log.Fatalln(err)
	}
	c, err = yaml.Marshal(certificates)
	if err != nil {
		log.Fatalln(err)
	}
	b = append(b, c...)

	snis, err := GetAllSNIs(client)
	if err != nil {
		log.Fatalln(err)
	}
	c, err = yaml.Marshal(snis)
	if err != nil {
		log.Fatalln(err)
	}
	b = append(b, c...)

	consumers, err := GetAllConsumers(client)
	if err != nil {
		log.Fatalln(err)
	}
	c, err = yaml.Marshal(consumers)
	if err != nil {
		log.Fatalln(err)
	}
	b = append(b, c...)

	upstreams, err := GetAllUpstreams(client)
	if err != nil {
		log.Fatalln(err)
	}
	c, err = yaml.Marshal(upstreams)
	if err != nil {
		log.Fatalln(err)
	}
	b = append(b, c...)

	fmt.Println(services)
	fmt.Println(routes)
	fmt.Print(plugins)
	fmt.Println(certificates)
	fmt.Println(snis)
	fmt.Println(consumers)
	fmt.Println(upstreams)

	err = ioutil.WriteFile("test2.out", b, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	// err = ioutil.WriteFile("test.out", b, 0644)

}

func GetAllServices(client *kong.Client) ([]*kong.Service, error) {
	var services []*kong.Service
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.Services.List(nil, opt)
		if err != nil {
			return nil, err
		}
		services = append(services, s...)
		if opt == nil {
			break
		}
	}
	return services, nil
}

func GetAllRoutes(client *kong.Client) ([]*kong.Route, error) {
	var routes []*kong.Route
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.Routes.List(nil, opt)
		if err != nil {
			return nil, err
		}
		routes = append(routes, s...)
		if opt == nil {
			break
		}
	}
	return routes, nil
}

func GetAllPlugins(client *kong.Client) ([]*kong.Plugin, error) {
	var plugins []*kong.Plugin
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.Plugins.List(nil, opt)
		if err != nil {
			return nil, err
		}
		plugins = append(plugins, s...)
		if opt == nil {
			break
		}
	}
	return plugins, nil
}

func GetAllCertificates(client *kong.Client) ([]*kong.Certificate, error) {
	var certificates []*kong.Certificate
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.Certificates.List(nil, opt)
		if err != nil {
			return nil, err
		}
		certificates = append(certificates, s...)
		if opt == nil {
			break
		}
	}
	return certificates, nil
}

func GetAllSNIs(client *kong.Client) ([]*kong.SNI, error) {
	var snis []*kong.SNI
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.SNIs.List(nil, opt)
		if err != nil {
			return nil, err
		}
		snis = append(snis, s...)
		if opt == nil {
			break
		}
	}
	return snis, nil
}

func GetAllConsumers(client *kong.Client) ([]*kong.Consumer, error) {
	var consumers []*kong.Consumer
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.Consumers.List(nil, opt)
		if err != nil {
			return nil, err
		}
		consumers = append(consumers, s...)
		if opt == nil {
			break
		}
	}
	return consumers, nil
}

func GetAllUpstreams(client *kong.Client) ([]*kong.Upstream, error) {
	var upstreams []*kong.Upstream
	opt := new(kong.ListOpt)
	opt.Size = 1000
	for {
		s, opt, err := client.Upstreams.List(nil, opt)
		if err != nil {
			return nil, err
		}
		upstreams = append(upstreams, s...)
		if opt == nil {
			break
		}
	}
	return upstreams, nil
}
