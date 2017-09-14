package main

import (
	"github.com/bwmarrin/discordgo"
	"fmt"
	"os"
	"syscall"
	"os/signal"
	"strings"
	"github.com/zmb3/spotify"
	"net/http"
	"unicode"
	"strconv"
	"github.com/ahl5esoft/golang-underscore"
)
import (
	_ "github.com/joho/godotenv/autoload"
	"log"
)


var (
	trackBuffer []spotify.FullTrack
)


func isNumeric(s string) bool {
	for _, c := range s {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func getCommandArgs(command string, input string) string {
	return strings.TrimSpace(strings.Replace(input, command, "", 1))
}

func resetTrackBuffer() {
	trackBuffer = trackBuffer[:0]
}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	if strings.HasPrefix(message.Content, "!") {
		command := strings.TrimLeft(message.Content, "!")
		commandArgs := strings.Split(command, " ")

		if len(commandArgs) == 0 {
			return
		}

		commandName := commandArgs[0]

		switch commandName {
			case "listen":
				embed := discordgo.MessageEmbed{
					Title: "LoreyBot - Spotify DJ",
					Color: 0x1ED760,
					Description: "[Click here to start listening]("+ GetAuthoriseUrl() +")",
				}

				session.ChannelMessageSendEmbed(message.ChannelID, &embed)
			case "search":
				query := getCommandArgs(commandName, command)

				result, err := Clients[0].Search(query, spotify.SearchTypeTrack)

				if err != nil {
					fmt.Println(err)
					return
				}

				resetTrackBuffer()

				var fields []*discordgo.MessageEmbedField;

				for i := 0; i < len(result.Tracks.Tracks); i++  {
					track :=  result.Tracks.Tracks[i]

					artistNames := underscore.MapBy(track.Artists, "Name").([]string)
					artistNamesString := strings.Join(artistNames, ", ")

					fields = append(fields, &discordgo.MessageEmbedField{
						Name: strconv.Itoa(i) + ". " + artistNamesString,
						Value: track.Name,
						Inline: true,
					})

					trackBuffer = append(trackBuffer, result.Tracks.Tracks[i])
				}

				embed := discordgo.MessageEmbed{
					Title: "Found the following tracks:",
					Color: 0x1ED760,
					Fields: fields,
					Description: "Type !play {number} to play the track\n\n",
				}

				session.ChannelMessageSendEmbed(message.ChannelID, &embed)
			case "play":
				spotifyUri := getCommandArgs(commandName, command)

				if len(Clients) == 0 {
					session.ChannelMessageSend(message.ChannelID, "Sorry, currently no listeners!")
					return
				}

				playOptions := spotify.PlayOptions {}
				var contextUri spotify.URI = spotify.URI(spotifyUri)

				if len(trackBuffer) > 0 && len(spotifyUri) == 1 && isNumeric(spotifyUri) {
					i, err := strconv.Atoi(spotifyUri)

					if err != nil {
						fmt.Println(err)
						return
					}

					track := trackBuffer[i]
					playOptions.URIs = []spotify.URI{track.URI}
				} else {
					playOptions.PlaybackContext = &contextUri
				}

				for i := 0; i < len(Clients); i++  {
					err := Clients[i].PlayOpt(&playOptions)

					if err != nil {
						fmt.Println(err)
					}
				}
		}
	}
}

func startHttpServer() *http.Server {
	srv := &http.Server{Addr: ":"+os.Getenv("HTTP_PORT")}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			// cannot panic, because this probably is an intentional close
			log.Printf("Httpserver: ListenAndServe() error: %s", err)
		}
	}()

	// returning reference so caller can call Shutdown()
	return srv
}

func main() {
	httpServer := startHttpServer()
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))

	if err != nil {
		fmt.Println(err)
	}

	discord.AddHandler(messageCreate)

	err = discord.Open()

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	httpServer.Close()
	StoreTokens()
	discord.Close()
}
