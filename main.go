package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

const prefix string = "go play "

var (
	token  string
	buffer = make([][]byte, 0)
	vc     *discordgo.VoiceConnection
)

func main() {
	token = Token()
	if token == "" {
		fmt.Println("No token provided. Please place a token followed by a ';' at ../token.txt")
		return
	}

	// Load the sound file.
	err := loadSound("./resource/music.dca")
	if err != nil {
		fmt.Println("Error loading sound: ", err)
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Register guildCreate as a callback for the guildCreate events.
	dg.AddHandler(guildCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("The bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func Token() string {
	b, err := ioutil.ReadFile("../token.txt")
	if err != nil {
		fmt.Println(err)
	}
	c := string(b)
	return strings.Split(c, ";")[0]
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	s.UpdateStatus(0, "go <youtube-url>")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.Contains(m.Content, "go") {
		// Find the channel that the message came from.
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			// Could not find channel.
			return
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			// Could not find guild.
			return
		}

		// url := strings.Split(m.Content, " ")[1]
		// fmt.Println(url)
		dsu := true // downloadSound(url)
		if !dsu {
			fmt.Println("Error downloading")
		}

		// Look for the message sender in that guild's current voice states.
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				if strings.Contains(m.Content, "join") {
					vc, _ = s.ChannelVoiceJoin(g.ID, vs.ChannelID, false, false)
					return
				} else if strings.Contains(m.Content, "leave") {
					vc.Disconnect()
					return
				}
				err = playSound(s, g.ID, vs.ChannelID)
				if err != nil {
					fmt.Println("Error playing sound:", err)
				}

				return
			}
		}
	}
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {

	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = s.ChannelMessageSend(channel.ID, "Gobot the Gopher has *ARRIVED!*")
			return
		}
	}
}

func downloadSound(url string) bool {
	if out, err := exec.Command("./download.sh", url).Output(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Println(out)
		return false
	} else {
		fmt.Println(out)
		return true
	}
}

// loadSound attempts to load an encoded sound file from disk.
func loadSound(sound string) error {

	buffer = make([][]byte, 0)
	file, err := os.Open(sound)
	if err != nil {
		fmt.Println("Error opening dca file:", err)
		return err
	}

	var opuslen int16
	binary.Read(file, binary.LittleEndian, &_)

	for {
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}

		if err != nil {
			fmt.Println("Error reading from dca file:", err)
			return err
		}

		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		if err != nil {
			fmt.Println("Error reading from dca file:", err)
			return err
		}

		buffer = append(buffer, InBuf)
	}
}

// playSound plays the current buffer to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID string) (err error) {

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
	for _, buff := range buffer {
		vc.OpusSend <- buff
	}

	// Stop speaking
	vc.Speaking(false)

	return nil
}
