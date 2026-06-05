package handlers

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"discord-tmdb-bot/tmdb"

	"github.com/bwmarrin/discordgo"
)

var (
	validCustomID     = regexp.MustCompile(`^(page|tpage|view|watch|trailer|cast|back)_(movie|tv)_[a-z0-9_]{1,20}$`)
	userRateLimiter   = NewRateLimiter(5, 10) // 5 req/s, burst 10
	buttonRateLimiter = NewRateLimiter(2, 5)  // 2 req/s, burst 5
)

type CommandHandler struct {
	TMDBClient *tmdb.Client
}

func (h *CommandHandler) StartCleanup(interval time.Duration) {
	go userRateLimiter.Cleanup(interval)
	go buttonRateLimiter.Cleanup(interval)
}

var slashCommands = []*discordgo.ApplicationCommand{
	{
		Name:        "popular",
		Description: "Get popular movies or TV shows",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "Media type",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Movies", Value: "movie"},
					{Name: "TV Shows", Value: "tv"},
				},
			},
		},
	},
	{
		Name:        "trending",
		Description: "Get trending movies or TV shows",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "Media type",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Movies", Value: "movie"},
					{Name: "TV Shows", Value: "tv"},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "window",
				Description: "Time window",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Today", Value: "day"},
					{Name: "This Week", Value: "week"},
				},
			},
		},
	},
	{
		Name:        "search",
		Description: "Search for movies or TV shows",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "type",
				Description: "Media type",
				Required:    true,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{Name: "Movie", Value: "movie"},
					{Name: "TV Show", Value: "tv"},
				},
			},
			{
				Type:         discordgo.ApplicationCommandOptionString,
				Name:         "query",
				Description:  "Search query — type to get live suggestions",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
}


func (h *CommandHandler) RegisterCommands(s *discordgo.Session) {
	s.AddHandler(h.handleSlashCommand)
	s.AddHandler(h.handleInteraction)

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as %s#%s — registering slash commands...", r.User.Username, r.User.Discriminator)
		for _, cmd := range slashCommands {
			if _, err := s.ApplicationCommandCreate(r.User.ID, "", cmd); err != nil {
				log.Printf("Failed to register command %q: %v", cmd.Name, err)
			}
		}
		log.Println("Slash commands registered. Bot is ready!")
	})
}

func (h *CommandHandler) handleSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		userID := interactionUserID(i)
		if userID == "" {
			return
		}
		cmdName := i.ApplicationCommandData().Name
		log.Printf("[cmd] /%s invoked by userID=%s", cmdName, userID)
		if !userRateLimiter.Allow(userID) {
			respondError(s, i, "⏳ You're doing that too fast. Please slow down.")
			return
		}
		switch cmdName {
		case "popular":
			h.handlePopularCommand(s, i)
		case "trending":
			h.handleTrendingCommand(s, i)
		case "search":
			h.handleSearchCommand(s, i)
		}
	case discordgo.InteractionApplicationCommandAutocomplete:
		h.handleAutocomplete(s, i)
	}
}

func (h *CommandHandler) handlePopularCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	mediaType := i.ApplicationCommandData().Options[0].StringValue()

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		log.Printf("[popular] defer failed: %v", err)
		return
	}

	var embed *discordgo.MessageEmbed
	var components []discordgo.MessageComponent
	if mediaType == "movie" {
		embed, components = h.createMovieListEmbed(1)
	} else {
		embed, components = h.createTVListEmbed(1)
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	}); err != nil {
		log.Printf("[popular] edit failed: %v", err)
	}
}

func (h *CommandHandler) handleTrendingCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options
	mediaType := opts[0].StringValue()
	timeWindow := opts[1].StringValue()

	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	}); err != nil {
		log.Printf("[trending] defer failed: %v", err)
		return
	}

	embed, components := h.createTrendingEmbed(mediaType, timeWindow, 1)

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	}); err != nil {
		log.Printf("[trending] edit failed: %v", err)
	}
}

func (h *CommandHandler) handleSearchCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options
	mediaType := opts[0].StringValue()
	query := strings.TrimSpace(opts[1].StringValue())

	if query == "" {
		respondError(s, i, "Search query cannot be empty.")
		return
	}
	if len(query) > 200 {
		query = query[:200]
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	var embed *discordgo.MessageEmbed
	var components []discordgo.MessageComponent

	if mediaType == "movie" {
		resp, err := h.TMDBClient.SearchMovies(query, 1)
		if err != nil {
			log.Printf("SearchMovies error: %v", err)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("Failed to search movies.")})
			return
		}
		if len(resp.Results) == 0 {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("No results found.")})
			return
		}
		embed, components = h.createSearchResultsEmbed(resp, "movie")
	} else {
		resp, err := h.TMDBClient.SearchTVShows(query, 1)
		if err != nil {
			log.Printf("SearchTVShows error: %v", err)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("Failed to search TV shows.")})
			return
		}
		if len(resp.Results) == 0 {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("No results found.")})
			return
		}
		embed, components = h.createTVSearchResultsEmbed(resp)
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	}); err != nil {
		log.Printf("[search] edit failed: %v", err)
	}
}

func (h *CommandHandler) handleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if data.Name != "search" {
		return
	}

	var mediaType, query string
	for _, opt := range data.Options {
		switch opt.Name {
		case "type":
			mediaType = opt.StringValue()
		case "query":
			if opt.Focused {
				query = opt.StringValue()
			}
		}
	}

	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, 8)

	if len(strings.TrimSpace(query)) >= 2 {
		if mediaType == "movie" {
			if resp, err := h.TMDBClient.SearchMovies(query, 1); err == nil {
				for _, m := range resp.Results[:min(8, len(resp.Results))] {
					year := ""
					if len(m.ReleaseDate) >= 4 {
						year = " (" + m.ReleaseDate[:4] + ")"
					}
					choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
						Name:  truncateString(m.Title+year, 100),
						Value: m.Title,
					})
				}
			}
		} else {
			if resp, err := h.TMDBClient.SearchTVShows(query, 1); err == nil {
				for _, show := range resp.Results[:min(8, len(resp.Results))] {
					year := ""
					if len(show.FirstAirDate) >= 4 {
						year = " (" + show.FirstAirDate[:4] + ")"
					}
					choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
						Name:  truncateString(show.Name+year, 100),
						Value: show.Name,
					})
				}
			}
		}
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{Choices: choices},
	})
}


func (h *CommandHandler) createMovieListEmbed(page int) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	resp, err := h.TMDBClient.GetPopularMovies(page)
	if err != nil {
		return errorEmbed("Failed to fetch popular movies."), nil
	}
	embed := &discordgo.MessageEmbed{
		Title:       "🎬 Popular Movies",
		Description: fmt.Sprintf("Page %d of %d", page, resp.TotalPages),
		Color:       0x00ff00,
		Thumbnail:   tmdbThumbnail(),
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}
	var btns []discordgo.MessageComponent
	for idx, movie := range resp.Results[:min(5, len(resp.Results))] {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s", (page-1)*5+idx+1, movie.Title),
			Value:  fmt.Sprintf("⭐ **%.1f** | 📅 %s", movie.VoteAverage, movie.ReleaseDate),
			Inline: false,
		})
		btns = append(btns, mkBtn(fmt.Sprintf("%d. %s", idx+1, truncateString(movie.Title, 20)),
			discordgo.PrimaryButton, fmt.Sprintf("view_movie_%d", movie.ID)))
	}
	return embed, paginatedComponents(btns, "page_movie", page, resp.TotalPages)
}

func (h *CommandHandler) createTVListEmbed(page int) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	resp, err := h.TMDBClient.GetPopularTVShows(page)
	if err != nil {
		return errorEmbed("Failed to fetch popular TV shows."), nil
	}
	embed := &discordgo.MessageEmbed{
		Title:       "📺 Popular TV Shows",
		Description: fmt.Sprintf("Page %d of %d", page, resp.TotalPages),
		Color:       0x0099ff,
		Thumbnail:   tmdbThumbnail(),
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}
	var btns []discordgo.MessageComponent
	for idx, show := range resp.Results[:min(5, len(resp.Results))] {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s", (page-1)*5+idx+1, show.Name),
			Value:  fmt.Sprintf("⭐ **%.1f** | 📅 %s", show.VoteAverage, show.FirstAirDate),
			Inline: false,
		})
		btns = append(btns, mkBtn(fmt.Sprintf("%d. %s", idx+1, truncateString(show.Name, 20)),
			discordgo.PrimaryButton, fmt.Sprintf("view_tv_%d", show.ID)))
	}
	return embed, paginatedComponents(btns, "page_tv", page, resp.TotalPages)
}

func (h *CommandHandler) createTrendingEmbed(mediaType, timeWindow string, page int) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	label := map[string]string{"day": "Today", "week": "This Week"}[timeWindow]

	if mediaType == "movie" {
		resp, err := h.TMDBClient.GetTrendingMovies(timeWindow, page)
		if err != nil {
			return errorEmbed("Failed to fetch trending movies."), nil
		}
		embed := &discordgo.MessageEmbed{
			Title:       fmt.Sprintf("🔥 Trending Movies — %s", label),
			Description: fmt.Sprintf("Page %d of %d", page, resp.TotalPages),
			Color:       0xff6600,
			Thumbnail:   tmdbThumbnail(),
			Fields:      make([]*discordgo.MessageEmbedField, 0),
		}
		var btns []discordgo.MessageComponent
		for idx, movie := range resp.Results[:min(5, len(resp.Results))] {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   fmt.Sprintf("%d. %s", (page-1)*5+idx+1, movie.Title),
				Value:  fmt.Sprintf("⭐ **%.1f** | 📅 %s", movie.VoteAverage, movie.ReleaseDate),
				Inline: false,
			})
			btns = append(btns, mkBtn(fmt.Sprintf("%d. %s", idx+1, truncateString(movie.Title, 20)),
				discordgo.PrimaryButton, fmt.Sprintf("view_movie_%d", movie.ID)))
		}
		return embed, paginatedComponents(btns, fmt.Sprintf("tpage_movie_%s", timeWindow), page, resp.TotalPages)
	}

	resp, err := h.TMDBClient.GetTrendingTVShows(timeWindow, page)
	if err != nil {
		return errorEmbed("Failed to fetch trending TV shows."), nil
	}
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🔥 Trending TV Shows — %s", label),
		Description: fmt.Sprintf("Page %d of %d", page, resp.TotalPages),
		Color:       0xff6600,
		Thumbnail:   tmdbThumbnail(),
		Fields:      make([]*discordgo.MessageEmbedField, 0),
	}
	var btns []discordgo.MessageComponent
	for idx, show := range resp.Results[:min(5, len(resp.Results))] {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s", (page-1)*5+idx+1, show.Name),
			Value:  fmt.Sprintf("⭐ **%.1f** | 📅 %s", show.VoteAverage, show.FirstAirDate),
			Inline: false,
		})
		btns = append(btns, mkBtn(fmt.Sprintf("%d. %s", idx+1, truncateString(show.Name, 20)),
			discordgo.PrimaryButton, fmt.Sprintf("view_tv_%d", show.ID)))
	}
	return embed, paginatedComponents(btns, fmt.Sprintf("tpage_tv_%s", timeWindow), page, resp.TotalPages)
}

func (h *CommandHandler) createSearchResultsEmbed(resp *tmdb.MovieResponse, mediaType string) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	embed := &discordgo.MessageEmbed{Title: "🔍 Movie Search Results", Color: 0xffcc00, Fields: make([]*discordgo.MessageEmbedField, 0)}
	var btns []discordgo.MessageComponent
	for idx, movie := range resp.Results[:min(5, len(resp.Results))] {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s", idx+1, movie.Title),
			Value:  fmt.Sprintf("⭐ **%.1f** | 📅 %s", movie.VoteAverage, movie.ReleaseDate),
			Inline: false,
		})
		btns = append(btns, mkBtn(fmt.Sprintf("%d. %s", idx+1, truncateString(movie.Title, 20)),
			discordgo.PrimaryButton, fmt.Sprintf("view_%s_%d", mediaType, movie.ID)))
	}
	var components []discordgo.MessageComponent
	if len(btns) > 0 {
		components = append(components, discordgo.ActionsRow{Components: btns})
	}
	return embed, components
}

func (h *CommandHandler) createTVSearchResultsEmbed(resp *tmdb.TVResponse) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	embed := &discordgo.MessageEmbed{Title: "🔍 TV Show Search Results", Color: 0xffcc00, Fields: make([]*discordgo.MessageEmbedField, 0)}
	var btns []discordgo.MessageComponent
	for idx, show := range resp.Results[:min(5, len(resp.Results))] {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s", idx+1, show.Name),
			Value:  fmt.Sprintf("⭐ **%.1f** | 📅 %s", show.VoteAverage, show.FirstAirDate),
			Inline: false,
		})
		btns = append(btns, mkBtn(fmt.Sprintf("%d. %s", idx+1, truncateString(show.Name, 20)),
			discordgo.PrimaryButton, fmt.Sprintf("view_tv_%d", show.ID)))
	}
	var components []discordgo.MessageComponent
	if len(btns) > 0 {
		components = append(components, discordgo.ActionsRow{Components: btns})
	}
	return embed, components
}


func paginatedComponents(itemBtns []discordgo.MessageComponent, prefix string, page, totalPages int) []discordgo.MessageComponent {
	var nav []discordgo.MessageComponent
	if page > 1 {
		nav = append(nav, mkBtn("◀️ Previous", discordgo.SecondaryButton, fmt.Sprintf("%s_%d", prefix, page-1)))
	}
	if page < totalPages {
		nav = append(nav, mkBtn("Next ▶️", discordgo.SecondaryButton, fmt.Sprintf("%s_%d", prefix, page+1)))
	}
	var out []discordgo.MessageComponent
	if len(itemBtns) > 0 {
		out = append(out, discordgo.ActionsRow{Components: itemBtns})
	}
	if len(nav) > 0 {
		out = append(out, discordgo.ActionsRow{Components: nav})
	}
	return out
}

func errorEmbed(msg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{Title: "Error", Description: msg, Color: 0xff0000}
}

func tmdbThumbnail() *discordgo.MessageEmbedThumbnail {
	return &discordgo.MessageEmbedThumbnail{
		URL: "https://www.themoviedb.org/assets/2/v4/logos/v2/blue_square_2-d537fb228cf3ded904ef09b136fe3fec72548ebc1fea3fbbd1ad9e36364db38b.svg",
	}
}

func respondError(s *discordgo.Session, i *discordgo.InteractionCreate, msg string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: msg, Flags: 64},
	})
}

func stringPtr(s string) *string { return &s }

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func interactionUserID(i *discordgo.InteractionCreate) string {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}
