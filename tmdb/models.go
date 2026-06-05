package tmdb

type MovieDetails struct {
	ID                   int           `json:"id"`
	Title                string        `json:"title"`
	Overview             string        `json:"overview"`
	ReleaseDate          string        `json:"release_date"`
	Runtime              int           `json:"runtime"`
	VoteAverage          float64       `json:"vote_average"`
	VoteCount            int           `json:"vote_count"`
	PosterPath           string        `json:"poster_path"`
	BackdropPath         string        `json:"backdrop_path"`
	Budget               int64         `json:"budget"`
	Revenue              int64         `json:"revenue"`
	Tagline              string        `json:"tagline"`
	Status               string        `json:"status"`
	Genres               []Genre       `json:"genres"`
	Credits              Credits       `json:"credits"`
	Videos               VideoResponse `json:"videos"`
	Similar              MovieResponse `json:"similar"`
	ProductionCompanies []Company     `json:"production_companies"`
}

type TVDetails struct {
	ID               int           `json:"id"`
	Name             string        `json:"name"`
	Overview         string        `json:"overview"`
	FirstAirDate     string        `json:"first_air_date"`
	LastAirDate      string        `json:"last_air_date"`
	NumberOfSeasons  int           `json:"number_of_seasons"`
	NumberOfEpisodes int           `json:"number_of_episodes"`
	EpisodeRuntime   []int         `json:"episode_run_time"`
	VoteAverage      float64       `json:"vote_average"`
	VoteCount        int           `json:"vote_count"`
	PosterPath       string        `json:"poster_path"`
	BackdropPath     string        `json:"backdrop_path"`
	Status           string        `json:"status"`
	Genres           []Genre       `json:"genres"`
	Credits          Credits       `json:"credits"`
	Videos           VideoResponse `json:"videos"`
	Seasons          []Season      `json:"seasons"`
}

type Movie struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float64 `json:"vote_average"`
	PosterPath  string  `json:"poster_path"`
	GenreIDs    []int   `json:"genre_ids"`
}

type TVShow struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview"`
	FirstAirDate string  `json:"first_air_date"`
	VoteAverage  float64 `json:"vote_average"`
	PosterPath   string  `json:"poster_path"`
	GenreIDs     []int   `json:"genre_ids"`
}

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Credits struct {
	Cast []Cast `json:"cast"`
	Crew []Crew `json:"crew"`
}

type Cast struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
	Order       int    `json:"order"`
}

type Crew struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Job        string `json:"job"`
	Department string `json:"department"`
}

type VideoResponse struct {
	Results []Video `json:"results"`
}

type Video struct {
	Key      string `json:"key"`
	Name     string `json:"name"`
	Site     string `json:"site"`
	Type     string `json:"type"`
	Official bool   `json:"official"`
}

type Season struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	SeasonNumber int    `json:"season_number"`
	EpisodeCount int    `json:"episode_count"`
	PosterPath   string `json:"poster_path"`
	AirDate      string `json:"air_date"`
}

type Company struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	LogoPath      string `json:"logo_path"`
	OriginCountry string `json:"origin_country"`
}

type MovieResponse struct {
	Page         int     `json:"page"`
	Results      []Movie `json:"results"`
	TotalPages   int     `json:"total_pages"`
	TotalResults int     `json:"total_results"`
}

type TVResponse struct {
	Page         int      `json:"page"`
	Results      []TVShow `json:"results"`
	TotalPages   int      `json:"total_pages"`
	TotalResults int      `json:"total_results"`
}

type WatchProvidersResponse struct {
	ID      int                            `json:"id"`
	Results map[string]WatchProviderRegion `json:"results"`
}

type WatchProviderRegion struct {
	Link     string          `json:"link"`
	FlatRate []WatchProvider `json:"flatrate"`
	Rent     []WatchProvider `json:"rent"`
	Buy      []WatchProvider `json:"buy"`
}

type WatchProvider struct {
	LogoPath        string `json:"logo_path"`
	ProviderID      int    `json:"provider_id"`
	ProviderName    string `json:"provider_name"`
	DisplayPriority int    `json:"display_priority"`
}
