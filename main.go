/*
 * These codes are licensed under CC0.
 * http://creativecommons.org/publicdomain/zero/1.0/deed.ja
 *
 */

// package
package main

// import
import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nlopes/slack"
	yaml "gopkg.in/yaml.v2"
)

// structs
type Data struct {
	Channels  ChannelINFO `yaml:"channels"`
	Responses Response    `yaml:"responses"`
}
type ChannelINFO struct {
	Sandbox    string `yaml:"sandbox"`
	Production string `yaml:"production"`
}
type Response struct {
	Init    []string `yaml:"init"`
	Message []string `yaml:"message"`
	Summary []string `yaml:"summary"`
}

// config manager
func readConfig() (d Data, err error) {
	buf, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
		return d, err
	}
	log.Printf("buf: %+v\n", string(buf))
	err = yaml.Unmarshal(buf, &d)
	if err != nil {
		log.Fatal(err)
		return d, err
	}
	return d, err
}

// counter and memory

var memoryFilename = "slackbot_memory.tmp"
var mesCounter = 0

func writeMemory() {
	log.Printf("Writing Memory : %v\n", strconv.Itoa(mesCounter))
	file, err := os.OpenFile(memoryFilename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	fmt.Fprint(file, strconv.Itoa(mesCounter))
}

func readMemory() {
	log.Print("Reading Memory ...")
	data, err := ioutil.ReadFile(memoryFilename)
	if err != nil {
		log.Print(err)
		mesCounter = 0
		return
	}
	ret, err := strconv.Atoi(string(data))
	if err != nil {
		log.Print(err)
		mesCounter = 0
		return
	}
	mesCounter = ret
	return
}

// catch os signal
func catchSig(sig os.Signal) {
	switch sig {
	case syscall.SIGHUP:
		log.Print("SIGHUP Happend! ", sig)
	case syscall.SIGTERM:
		log.Print("SIGTERM Happend! ", sig)
	case syscall.SIGKILL:
		log.Print("SIGKILL Happend! ", sig)
	default:
		log.Print("Other singal Happend! ", sig)
	}
}

// main
func run(api *slack.Client) int {
	d, err := readConfig()
	if err != nil {
		log.Fatal(err)
		return 1
	}

	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)

	readMemory()

	rtm := api.NewRTM()
	go rtm.ManageConnection()
	summaryPosted := false
	for {
		// summary
		h, m, _ := time.Now().Clock()
		if h == 0 && m == 0 && !summaryPosted {
			log.Printf("Daily summary : %v\n", strconv.Itoa(mesCounter))
			// TODO : exclude message (in macro) to config file
			mes := "今日は **" + strconv.Itoa(mesCounter) + "回** 発言できてたよ。\n" + d.Responses.Summary[rand.Intn(len(d.Responses.Summary))]
			nMes := rtm.NewOutgoingMessage(mes, d.Channels.Production)
			rtm.SendMessage(nMes)

			mesCounter = 0
			summaryPosted = true
		}
		if h == 0 && m == 5 && summaryPosted {
			summaryPosted = false
		}

		// RTM Event handling
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				log.Print("goslackbot Ready.")
				mes := d.Responses.Init[rand.Intn(len(d.Responses.Init))]
				rtm.SendMessage(rtm.NewOutgoingMessage(mes, d.Channels.Sandbox))

			case *slack.MessageEvent:
				// TODO : add action
				if ev.Channel != d.Channels.Production {
					continue
				}
				log.Printf("Message: %v\n", ev)
				mes := d.Responses.Message[rand.Intn(len(d.Responses.Message))]
				nMes := rtm.NewOutgoingMessage(mes, d.Channels.Production)
				rtm.SendMessage(nMes)
				mesCounter++

			case *slack.InvalidAuthEvent:
				log.Print("Invalid credentials")
				return 1

			}
		// catch os signal
		case ch := <-signalCh:
			catchSig(ch)
			return 0
		}
	}
}

// enrty
func main() {
	rand.Seed(time.Now().UnixNano())
	api := slack.New(os.Getenv("GOSLACKBOT_APITOKEN"))
	ret := run(api)
	writeMemory()
	os.Exit(ret)
}
