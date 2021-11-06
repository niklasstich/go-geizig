package main

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	Token string
	Channel string
	Watchlists []string
)

func init() {
	var ok bool
	Token, ok = os.LookupEnv("TOKEN")
	if !ok || Token == "" {
		log.Fatalln("No valid token in environment variable TOKEN")
	}

	Channel, ok = os.LookupEnv("CHANNEL")
	if !ok || Channel == "" {
		log.Fatalln("No valid channel in environment variable TOKEN")
	}

	temp, ok := os.LookupEnv("WATCHLISTS")
	if !ok || temp == "" {
		log.Fatalln("No valid watchlists in environment variable WATCHLISTS")
	}
	Watchlists = strings.Split(temp, ",")
}

func main() {
	log.Info("Connecting to Discord...")
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalln("Failed to open discord session, ", err.Error())
	}

	dg.AddHandler(func(_ *discordgo.Session, _ *discordgo.Ready){
		log.Println("Connected.")
	})

	defer dg.Close()

	go func() {
		for {
			for _, URL := range Watchlists {
				go FetchUpdateAndPost(dg, Channel, URL)
				log.Infof("Updated %s, next update at %v", URL, time.Now().Add(time.Hour*6).Format(time.RFC1123))
			}
			//sleep 6 hours
			time.Sleep(time.Hour*6)
		}
	}()

	log.Info("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	log.Info("Shutting down gracefully")
}

func FetchUpdateAndPost(dg *discordgo.Session, channel, URL string) {
	resp, err := http.Get(URL)
	if err != nil {
		log.Errorf("Failed to get %s, %v", URL, err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Failed to read body for %s, %v", URL, err)
		return
	}

	doc, err := htmlquery.Parse(strings.NewReader(string(body)))
	if err != nil {
		log.Errorf("Failed to parse Document for %s, %v", URL, err)
		return
	}

	node := htmlquery.FindOne(doc, "/html/body/div/div[1]/main/div[2]/div[2]/form/div[18]/div[3]/div/span/span/span/span")
	price := node.FirstChild.Data

	node = htmlquery.FindOne(doc, "/html/body/div/div[1]/main/div[2]/div[2]/div/h1/span")
	name := strings.ReplaceAll(node.FirstChild.Data, "\n", "")


	msg := fmt.Sprintf("Wishlist %s (%s) costs %s at %v", name, URL, price, time.Now().Format(time.RFC1123))
	_, err = dg.ChannelMessageSend(channel, msg)
	if err != nil {
		log.Errorf("Failed to send message to channel %s, %v", channel, err)
		return
	}
}

