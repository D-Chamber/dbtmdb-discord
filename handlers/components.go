package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"discord-tmdb-bot/tmdb"

	"github.com/bwmarrin/discordgo"
)

func (h *CommandHandler) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	data := i.MessageComponentData()
	if !validCustomID.MatchString(data.CustomID) {
		respondError(s, i, "Invalid interaction.")
		return
	}

	userID := interactionUserID(i)
	if userID == "" {
		return
	}
	if !buttonRateLimiter.Allow(userID) {
		respondError(s, i, "Too many interactions. Slow down!")
		return
	}

	parts := strings.Split(data.CustomID, "_")
	if len(parts) < 3 {
		return
	}
	action := parts[0]
	mediaType := parts[1]

	if action == "tpage" {
		if len(parts) < 4 {
			return
		}
		timeWindow := parts[2]
		page, err := strconv.Atoi(parts[3])
		if err != nil || page < 1 {
			return
		}
		h.handleTrendingNavigation(s, i, mediaType, timeWindow, page)
		return
	}

	id, err := strconv.Atoi(parts[2])
	if err != nil || id < 0 {
		return
	}
	switch action {
	case "page":
		h.handlePageNavigation(s, i, mediaType, id)
	case "view":
		h.handleViewDetails(s, i, mediaType, id)
	case "watch":
		h.handleWatchProviders(s, i, mediaType, id)
	case "trailer":
		h.handleTrailer(s, i, mediaType, id)
	case "cast":
		h.handleCast(s, i, mediaType, id)
	case "back":
		h.handleBackToList(s, i, mediaType)
	}
}

func (h *CommandHandler) handlePageNavigation(s *discordgo.Session, i *discordgo.InteractionCreate, mediaType string, page int) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("[page] defer failed: %v", err)
		return
	}

	var embed *discordgo.MessageEmbed
	var components []discordgo.MessageComponent

	if mediaType == "movie" {
		resp, err := h.TMDBClient.GetPopularMovies(page)
		if err != nil {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("❌ Error fetching page.")})
			return
		}
		if page < 1 || page > resp.TotalPages {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("❌ Invalid page number.")})
			return
		}
		embed, components = h.createMovieListEmbed(page)
	} else {
		resp, err := h.TMDBClient.GetPopularTVShows(page)
		if err != nil {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("❌ Error fetching page.")})
			return
		}
		if page < 1 || page > resp.TotalPages {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("❌ Invalid page number.")})
			return
		}
		embed, components = h.createTVListEmbed(page)
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	}); err != nil {
		log.Printf("[page] edit failed: %v", err)
	}
}

func (h *CommandHandler) handleTrendingNavigation(s *discordgo.Session, i *discordgo.InteractionCreate, mediaType, timeWindow string, page int) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("[trending-nav] defer failed: %v", err)
		return
	}
	embed, components := h.createTrendingEmbed(mediaType, timeWindow, page)
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	}); err != nil {
		log.Printf("[trending-nav] edit failed: %v", err)
	}
}


func (h *CommandHandler) handleViewDetails(s *discordgo.Session, i *discordgo.InteractionCreate, mediaType string, id int) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("[view] defer failed: %v", err)
		return
	}

	if mediaType == "movie" {
		details, err := h.TMDBClient.GetMovieDetails(id)
		if err != nil {
			log.Printf("[view] GetMovieDetails error: %v", err)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("❌ Failed to load movie details.")})
			return
		}
		embed, components := h.createMovieDetailEmbed(details)
		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed}, Components: &components,
		}); err != nil {
			log.Printf("[view] movie edit failed: %v", err)
		}
	} else {
		details, err := h.TMDBClient.GetTVDetails(id)
		if err != nil {
			log.Printf("[view] GetTVDetails error: %v", err)
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("❌ Failed to load TV show details.")})
			return
		}
		embed, components := h.createTVDetailEmbed(details)
		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Embeds: &[]*discordgo.MessageEmbed{embed}, Components: &components,
		}); err != nil {
			log.Printf("[view] tv edit failed: %v", err)
		}
	}
}

func (h *CommandHandler) createMovieDetailEmbed(details *tmdb.MovieDetails) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	genres := genreNames(details.Genres)
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎬 %s (%s)", details.Title, yearOf(details.ReleaseDate)),
		Description: details.Overview,
		Color:       0x00ff00,
		Image:       &discordgo.MessageEmbedImage{URL: h.TMDBClient.GetImageURL(details.BackdropPath, "w1280")},
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: h.TMDBClient.GetImageURL(details.PosterPath, "w500")},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "⭐ Rating", Value: fmt.Sprintf("%.1f/10 (%d votes)", details.VoteAverage, details.VoteCount), Inline: true},
			{Name: "⏱️ Runtime", Value: fmt.Sprintf("%d minutes", details.Runtime), Inline: true},
			{Name: "🎭 Genres", Value: strings.Join(genres, ", "), Inline: true},
			{Name: "💰 Budget", Value: fmt.Sprintf("$%d", details.Budget), Inline: true},
			{Name: "💵 Revenue", Value: fmt.Sprintf("$%d", details.Revenue), Inline: true},
			{Name: "🏢 Studios", Value: formatCompanies(details.ProductionCompanies), Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Status: %s | %s", details.Status, details.Tagline)},
	}
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			mkBtn("🎬 Trailer", discordgo.DangerButton, fmt.Sprintf("trailer_movie_%d", details.ID)),
			mkBtn("👥 Cast", discordgo.SuccessButton, fmt.Sprintf("cast_movie_%d", details.ID)),
			mkBtn("📺 Watch", discordgo.PrimaryButton, fmt.Sprintf("watch_movie_%d", details.ID)),
			mkBtn("🔙 Back", discordgo.SecondaryButton, fmt.Sprintf("back_movie_%d", details.ID)),
		}},
	}
	return embed, components
}

func (h *CommandHandler) createTVDetailEmbed(details *tmdb.TVDetails) (*discordgo.MessageEmbed, []discordgo.MessageComponent) {
	genres := genreNames(details.Genres)
	runtime := "N/A"
	if len(details.EpisodeRuntime) > 0 {
		runtime = fmt.Sprintf("%d min/ep", details.EpisodeRuntime[0])
	}
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📺 %s (%s)", details.Name, yearOf(details.FirstAirDate)),
		Description: details.Overview,
		Color:       0x0099ff,
		Image:       &discordgo.MessageEmbedImage{URL: h.TMDBClient.GetImageURL(details.BackdropPath, "w1280")},
		Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: h.TMDBClient.GetImageURL(details.PosterPath, "w500")},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "⭐ Rating", Value: fmt.Sprintf("%.1f/10 (%d votes)", details.VoteAverage, details.VoteCount), Inline: true},
			{Name: "⏱️ Runtime", Value: runtime, Inline: true},
			{Name: "🎭 Genres", Value: strings.Join(genres, ", "), Inline: true},
			{Name: "📺 Seasons", Value: fmt.Sprintf("%d seasons, %d episodes", details.NumberOfSeasons, details.NumberOfEpisodes), Inline: true},
			{Name: "📅 Aired", Value: fmt.Sprintf("%s → %s", details.FirstAirDate, details.LastAirDate), Inline: true},
			{Name: "📊 Status", Value: details.Status, Inline: true},
		},
		Footer: &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Status: %s", details.Status)},
	}
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{Components: []discordgo.MessageComponent{
			mkBtn("🎬 Trailer", discordgo.DangerButton, fmt.Sprintf("trailer_tv_%d", details.ID)),
			mkBtn("👥 Cast", discordgo.SuccessButton, fmt.Sprintf("cast_tv_%d", details.ID)),
			mkBtn("📺 Watch", discordgo.PrimaryButton, fmt.Sprintf("watch_tv_%d", details.ID)),
			mkBtn("🔙 Back", discordgo.SecondaryButton, fmt.Sprintf("back_tv_%d", details.ID)),
		}},
	}
	return embed, components
}

// ── where to watch ────────────────────────────────────────────────────────────

func (h *CommandHandler) handleWatchProviders(s *discordgo.Session, i *discordgo.InteractionCreate, mediaType string, id int) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	}); err != nil {
		log.Printf("[watch] defer failed: %v", err)
		return
	}

	var title string
	var providers *tmdb.WatchProvidersResponse
	var err error

	if mediaType == "movie" {
		d, e := h.TMDBClient.GetMovieDetails(id)
		if e != nil {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("❌ Failed to fetch watch providers.")})
			return
		}
		title = d.Title
		providers, err = h.TMDBClient.GetMovieWatchProviders(id)
	} else {
		d, e := h.TMDBClient.GetTVDetails(id)
		if e != nil {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("❌ Failed to fetch watch providers.")})
			return
		}
		title = d.Name
		providers, err = h.TMDBClient.GetTVWatchProviders(id)
	}

	if err != nil {
		log.Printf("[watch] provider fetch error: %v", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("❌ Failed to fetch watch providers.")})
		return
	}

	region, ok := providers.Results["US"]
	if !ok || (len(region.FlatRate) == 0 && len(region.Rent) == 0 && len(region.Buy) == 0) {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: stringPtr("No streaming data available for this title in the US."),
		})
		return
	}

	embed := buildWatchEmbed(title, region)
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	}); err != nil {
		log.Printf("[watch] edit failed: %v", err)
	}
}

func buildWatchEmbed(title string, r tmdb.WatchProviderRegion) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("📺 Where to Watch — %s", title),
		Color:  0x1db954,
		Fields: make([]*discordgo.MessageEmbedField, 0),
		Footer: &discordgo.MessageEmbedFooter{Text: "Streaming data powered by JustWatch • Region: 🇺🇸 US"},
	}
	if len(r.FlatRate) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: "🎬 Stream", Value: providerNames(r.FlatRate), Inline: true,
		})
	}
	if len(r.Rent) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: "🛒 Rent", Value: providerNames(r.Rent), Inline: true,
		})
	}
	if len(r.Buy) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: "💳 Buy", Value: providerNames(r.Buy), Inline: true,
		})
	}
	return embed
}

func providerNames(ps []tmdb.WatchProvider) string {
	names := make([]string, 0, len(ps))
	for _, p := range ps {
		names = append(names, p.ProviderName)
	}
	return strings.Join(names, "\n")
}


func (h *CommandHandler) handleTrailer(s *discordgo.Session, i *discordgo.InteractionCreate, mediaType string, id int) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	}); err != nil {
		log.Printf("[trailer] defer failed: %v", err)
		return
	}

	var videos []tmdb.Video
	if mediaType == "movie" {
		d, err := h.TMDBClient.GetMovieDetails(id)
		if err != nil {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("❌ Failed to fetch trailer.")})
			return
		}
		videos = d.Videos.Results
	} else {
		d, err := h.TMDBClient.GetTVDetails(id)
		if err != nil {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("❌ Failed to fetch trailer.")})
			return
		}
		videos = d.Videos.Results
	}

	url := ""
	for _, v := range videos {
		if v.Site == "YouTube" && v.Type == "Trailer" && v.Official {
			url = fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.Key)
			break
		}
	}
	if url == "" {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("No official trailer available.")})
		return
	}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: stringPtr(fmt.Sprintf("🎬 Trailer: %s", url)),
	}); err != nil {
		log.Printf("[trailer] edit failed: %v", err)
	}
}

func (h *CommandHandler) handleCast(s *discordgo.Session, i *discordgo.InteractionCreate, mediaType string, id int) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	}); err != nil {
		log.Printf("[cast] defer failed: %v", err)
		return
	}

	var credits tmdb.Credits
	var title string
	if mediaType == "movie" {
		d, err := h.TMDBClient.GetMovieDetails(id)
		if err != nil {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("❌ Failed to fetch cast.")})
			return
		}
		credits, title = d.Credits, d.Title
	} else {
		d, err := h.TMDBClient.GetTVDetails(id)
		if err != nil {
			s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: stringPtr("❌ Failed to fetch cast.")})
			return
		}
		credits, title = d.Credits, d.Name
	}

	embed := buildCastEmbed(title, credits)
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Embeds: &[]*discordgo.MessageEmbed{embed},
	}); err != nil {
		log.Printf("[cast] edit failed: %v", err)
	}
}

func buildCastEmbed(title string, credits tmdb.Credits) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("👥 Cast & Crew — %s", title),
		Color:  0xffa500,
		Fields: make([]*discordgo.MessageEmbedField, 0),
	}
	castText := ""
	for idx, c := range credits.Cast[:min(10, len(credits.Cast))] {
		castText += fmt.Sprintf("**%s** as *%s*\n", c.Name, c.Character)
		if idx == 4 && len(credits.Cast) > 5 {
			castText += fmt.Sprintf("*...and %d more*\n", len(credits.Cast)-5)
			break
		}
	}
	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "🎭 Main Cast", Value: castText, Inline: false})

	dirs := filterCrew(credits.Crew, "Directing")
	writers := filterCrew(credits.Crew, "Writing")
	crew := ""
	if len(dirs) > 0 {
		crew += "**Director:** " + dirs[0].Name + "\n"
	}
	if len(writers) > 0 {
		crew += "**Writer:** " + writers[0].Name + "\n"
	}
	if crew != "" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "🎬 Key Crew", Value: crew, Inline: false})
	}
	return embed
}


func (h *CommandHandler) handleBackToList(s *discordgo.Session, i *discordgo.InteractionCreate, mediaType string) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	}); err != nil {
		log.Printf("[back] defer failed: %v", err)
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
		Embeds: &[]*discordgo.MessageEmbed{embed}, Components: &components,
	}); err != nil {
		log.Printf("[back] edit failed: %v", err)
	}
}


func filterCrew(crew []tmdb.Crew, department string) []tmdb.Crew {
	var out []tmdb.Crew
	for _, c := range crew {
		if c.Department == department {
			out = append(out, c)
		}
	}
	return out
}

func formatCompanies(companies []tmdb.Company) string {
	if len(companies) == 0 {
		return "N/A"
	}
	names := make([]string, 0, 3)
	for _, c := range companies[:min(3, len(companies))] {
		names = append(names, c.Name)
	}
	return strings.Join(names, ", ")
}

func genreNames(genres []tmdb.Genre) []string {
	names := make([]string, len(genres))
	for i, g := range genres {
		names[i] = g.Name
	}
	return names
}

func yearOf(date string) string {
	if len(date) >= 4 {
		return date[:4]
	}
	return "N/A"
}
