package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	Token      string // API Read Access Token (Bearer token)
	BaseURL    string
	ImageURL   string
	HTTPClient *http.Client
	Cache      *Cache
}

func NewClient(token string) *Client {
	return &Client{
		Token:    token,
		BaseURL:  "https://api.themoviedb.org/3",
		ImageURL: "https://image.tmdb.org/t/p",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Cache: NewCache(15 * time.Minute),
	}
}

func (c *Client) buildURL(path string, params map[string]string) string {
	u, _ := url.Parse(c.BaseURL + path)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func (c *Client) do(rawURL string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("TMDB API returned %s: %s", resp.Status, string(body))
	}
	return resp, nil
}

func (c *Client) GetPopularMovies(page int) (*MovieResponse, error) {
	cacheKey := fmt.Sprintf("popular_movies_%d", page)
	var resp MovieResponse
	if c.Cache.GetAndUnmarshal(cacheKey, &resp) {
		return &resp, nil
	}

	rawURL := c.buildURL("/movie/popular", map[string]string{
		"language": "en-US",
		"page":     strconv.Itoa(page),
	})

	httpResp, err := c.do(rawURL)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	c.Cache.MarshalAndSet(cacheKey, resp)
	return &resp, nil
}

func (c *Client) GetPopularTVShows(page int) (*TVResponse, error) {
	cacheKey := fmt.Sprintf("popular_tv_%d", page)
	var resp TVResponse
	if c.Cache.GetAndUnmarshal(cacheKey, &resp) {
		return &resp, nil
	}

	rawURL := c.buildURL("/tv/popular", map[string]string{
		"language": "en-US",
		"page":     strconv.Itoa(page),
	})

	httpResp, err := c.do(rawURL)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	c.Cache.MarshalAndSet(cacheKey, resp)
	return &resp, nil
}

func (c *Client) GetMovieDetails(movieID int) (*MovieDetails, error) {
	cacheKey := fmt.Sprintf("movie_details_%d", movieID)
	var details MovieDetails
	if c.Cache.GetAndUnmarshal(cacheKey, &details) {
		return &details, nil
	}

	rawURL := c.buildURL(fmt.Sprintf("/movie/%d", movieID), map[string]string{
		"language":           "en-US",
		"append_to_response": "credits,videos,similar",
	})

	resp, err := c.do(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	c.Cache.MarshalAndSet(cacheKey, details)
	return &details, nil
}

func (c *Client) GetTVDetails(tvID int) (*TVDetails, error) {
	cacheKey := fmt.Sprintf("tv_details_%d", tvID)
	var details TVDetails
	if c.Cache.GetAndUnmarshal(cacheKey, &details) {
		return &details, nil
	}

	rawURL := c.buildURL(fmt.Sprintf("/tv/%d", tvID), map[string]string{
		"language":           "en-US",
		"append_to_response": "credits,videos",
	})

	resp, err := c.do(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	c.Cache.MarshalAndSet(cacheKey, details)
	return &details, nil
}

func (c *Client) SearchMovies(query string, page int) (*MovieResponse, error) {
	cacheKey := fmt.Sprintf("search_movie_%s_%d", query, page)
	var resp MovieResponse
	if c.Cache.GetAndUnmarshal(cacheKey, &resp) {
		return &resp, nil
	}

	rawURL := c.buildURL("/search/movie", map[string]string{
		"query":    query,
		"language": "en-US",
		"page":     strconv.Itoa(page),
	})

	httpResp, err := c.do(rawURL)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	c.Cache.MarshalAndSet(cacheKey, resp)
	return &resp, nil
}

func (c *Client) SearchTVShows(query string, page int) (*TVResponse, error) {
	cacheKey := fmt.Sprintf("search_tv_%s_%d", query, page)
	var resp TVResponse
	if c.Cache.GetAndUnmarshal(cacheKey, &resp) {
		return &resp, nil
	}

	rawURL := c.buildURL("/search/tv", map[string]string{
		"query":    query,
		"language": "en-US",
		"page":     strconv.Itoa(page),
	})

	httpResp, err := c.do(rawURL)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}

	c.Cache.MarshalAndSet(cacheKey, resp)
	return &resp, nil
}

func (c *Client) GetTrendingMovies(timeWindow string, page int) (*MovieResponse, error) {
	cacheKey := fmt.Sprintf("trending_movies_%s_%d", timeWindow, page)
	var resp MovieResponse
	if c.Cache.GetAndUnmarshal(cacheKey, &resp) {
		return &resp, nil
	}
	rawURL := c.buildURL(fmt.Sprintf("/trending/movie/%s", timeWindow), map[string]string{
		"language": "en-US",
		"page":     strconv.Itoa(page),
	})
	httpResp, err := c.do(rawURL)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}
	c.Cache.MarshalAndSet(cacheKey, resp)
	return &resp, nil
}

func (c *Client) GetTrendingTVShows(timeWindow string, page int) (*TVResponse, error) {
	cacheKey := fmt.Sprintf("trending_tv_%s_%d", timeWindow, page)
	var resp TVResponse
	if c.Cache.GetAndUnmarshal(cacheKey, &resp) {
		return &resp, nil
	}
	rawURL := c.buildURL(fmt.Sprintf("/trending/tv/%s", timeWindow), map[string]string{
		"language": "en-US",
		"page":     strconv.Itoa(page),
	})
	httpResp, err := c.do(rawURL)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}
	c.Cache.MarshalAndSet(cacheKey, resp)
	return &resp, nil
}

func (c *Client) GetMovieWatchProviders(movieID int) (*WatchProvidersResponse, error) {
	cacheKey := fmt.Sprintf("watch_providers_movie_%d", movieID)
	var resp WatchProvidersResponse
	if c.Cache.GetAndUnmarshal(cacheKey, &resp) {
		return &resp, nil
	}
	rawURL := c.buildURL(fmt.Sprintf("/movie/%d/watch/providers", movieID), nil)
	httpResp, err := c.do(rawURL)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}
	c.Cache.MarshalAndSet(cacheKey, resp)
	return &resp, nil
}

func (c *Client) GetTVWatchProviders(tvID int) (*WatchProvidersResponse, error) {
	cacheKey := fmt.Sprintf("watch_providers_tv_%d", tvID)
	var resp WatchProvidersResponse
	if c.Cache.GetAndUnmarshal(cacheKey, &resp) {
		return &resp, nil
	}
	rawURL := c.buildURL(fmt.Sprintf("/tv/%d/watch/providers", tvID), nil)
	httpResp, err := c.do(rawURL)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, err
	}
	c.Cache.MarshalAndSet(cacheKey, resp)
	return &resp, nil
}


func (c *Client) GetImageURL(path string, size string) string {
	if path == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s%s", c.ImageURL, size, path)
}
