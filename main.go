/*
 * These codes are licensed under CC0.
 * http://creativecommons.org/publicdomain/zero/1.0/deed.ja
 *
 */

package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/nlopes/slack"
	yaml "gopkg.in/yaml.v2"
)

type Data struct {
	Responses Response `yaml:"responses"`
}

type Response struct {
	Init    []string `yaml:"init"`
	Message []string `yaml:"message"`
}

func readConfig() (d Data, err error) {
	buf, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
		return d, err
	}
	log.Printf("buf: %+v\n", string(buf))

	// structにUnmasrshal
	err = yaml.Unmarshal(buf, &d)
	if err != nil {
		log.Fatal(err)
		return d, err
	}
	return d, err
}

func run(api *slack.Client) int {
	d, err := readConfig()
	if err != nil {
		log.Fatal(err)
		return 1
	}

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				log.Print("goslackbot Ready.")
				mes := d.Responses.Init[rand.Intn(len(d.Responses.Init))]
				rtm.SendMessage(rtm.NewOutgoingMessage(mes, "CCXPDQY13"))

			case *slack.MessageEvent:
				log.Printf("Message: %v\n", ev)
				log.Printf(ev.Channel)
				mes := d.Responses.Message[rand.Intn(len(d.Responses.Message))]
				rtm.SendMessage(rtm.NewOutgoingMessage(mes, ev.Channel))

			case *slack.InvalidAuthEvent:
				log.Print("Invalid credentials")
				return 1

			}
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	api := slack.New(os.Getenv("GOSLACKBOT_APITOKEN"))
	os.Exit(run(api))
}
