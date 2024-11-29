package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
	Token            string
	voiceConnections = make(map[string]*discordgo.VoiceConnection) // Map to track voice connections per guild
)

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file:", err)
	}

	// Get token from environment
	Token = os.Getenv("BOT_TOKEN")
	if Token == "" {
		log.Fatalln("No token provided. Please set BOT_TOKEN in your .env file")
	}
}

func main() {
	// Create a new Discord session
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	// Register handlers
	dg.AddHandler(messageCreate)

	// Open websocket connection
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}

	fmt.Println("Bot is running. Press CTRL-C to exit.")

	// Wait for CTRL-C
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	// Close all voice connections and the session
	for _, vc := range voiceConnections {
		vc.Disconnect()
	}
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Example commands
	if m.Content == "!ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	if m.Content == "!join" {
		joinVoiceChannel(s, m)
	}

	if m.Content == "!leave" {
		leaveVoiceChannel(s, m)
	}
}

func joinVoiceChannel(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Get the guild (server) the message was sent in
	guild, err := s.State.Guild(m.GuildID)
	if err != nil {
		fmt.Println("Error finding guild:", err)
		return
	}

	// Get the author’s voice state
	var voiceChannelID string
	for _, vs := range guild.VoiceStates {
		if vs.UserID == m.Author.ID {
			voiceChannelID = vs.ChannelID
			break
		}
	}

	// If the user is not in a voice channel, notify them
	if voiceChannelID == "" {
		s.ChannelMessageSend(m.ChannelID, "You need to be in a voice channel for me to join!")
		return
	}

	// Check if the bot is already in a voice channel in this guild
	if _, exists := voiceConnections[m.GuildID]; exists {
		s.ChannelMessageSend(m.ChannelID, "I'm already in a voice channel!")
		return
	}

	// Join the user's voice channel
	vc, err := s.ChannelVoiceJoin(m.GuildID, voiceChannelID, false, false)
	if err != nil {
		fmt.Println("Error joining voice channel:", err)
		s.ChannelMessageSend(m.ChannelID, "I couldn't join the voice channel.")
		return
	}

	// Store the voice connection
	voiceConnections[m.GuildID] = vc
	s.ChannelMessageSend(m.ChannelID, "Joined your voice channel!")
}

func leaveVoiceChannel(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Check if the bot is connected in this guild
	if vc, exists := voiceConnections[m.GuildID]; exists {
		vc.Disconnect()
		delete(voiceConnections, m.GuildID)
		s.ChannelMessageSend(m.ChannelID, "I have left the voice channel.")
	} else {
		s.ChannelMessageSend(m.ChannelID, "I'm not in a voice channel!")
	}
}
