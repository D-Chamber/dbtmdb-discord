package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"discord-tmdb-bot/config"
	"discord-tmdb-bot/handlers"
	"discord-tmdb-bot/tmdb"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg := config.Load()

	tmdbClient := tmdb.NewClient(cfg.TMDBToken)
	tmdbClient.Cache = tmdb.NewCache(cfg.CacheDuration)

	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	cmdHandler := &handlers.CommandHandler{
		TMDBClient: tmdbClient,
	}
	cmdHandler.RegisterCommands(dg)

	// Start background goroutines to evict stale rate-limiter entries.
	cmdHandler.StartCleanup(10 * time.Minute)

	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening connection:", err)
	}
	defer dg.Close()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
