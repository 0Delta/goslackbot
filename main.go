package main

import (
    "log"
    "os"

    "github.com/nlopes/slack"
)

func run(api *slack.Client) int {
    rtm := api.NewRTM()
    go rtm.ManageConnection()

    for {
        select {
        case msg := <-rtm.IncomingEvents:
            switch ev := msg.Data.(type) {
            case *slack.HelloEvent:
                log.Print("goslackbot Ready.")
                rtm.SendMessage(rtm.NewOutgoingMessage("再起動が完了しましたよ、マスター", "CCXPDQY13"))

            case *slack.MessageEvent:
                log.Printf("Message: %v\n", ev)
                log.Printf(ev.Channel)
                rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", ev.Channel))

            case *slack.InvalidAuthEvent:
                log.Print("Invalid credentials")
                return 1

            }
        }
    }
}

func main() {
    api := slack.New(os.Getenv("GOSLACKBOT_APITOKEN"))
    os.Exit(run(api))
}
