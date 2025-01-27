package main

import (
	"Gator/internal/config"
	"Gator/internal/database"
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"html"
	"io"
	"net/http"
	"os"
	"time"
)

type state struct {
	db *database.Queries
	*config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (c *commands) register(name string, f func(*state, command) error) {
	if c.handlers == nil {
		c.handlers = make(map[string]func(*state, command) error)
	}
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	if c.handlers == nil {
		return fmt.Errorf("No handlers registered")
	}
	handler, exists := c.handlers[cmd.name]
	if !exists {
		fmt.Printf("command %s not found", cmd.name)
		return fmt.Errorf("command %s not found", cmd.name)
	}
	return handler(s, cmd)
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		if s.CurrentUser == "" {
			return fmt.Errorf("no user is logged in")
		}

		ctx := context.Background()
		user, err := s.db.GetUser(ctx, s.CurrentUser)
		if err != nil {
			return fmt.Errorf("unable to get user from DB: %v", err)
		}
		return handler(s, cmd, user)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("invalid login command")
	}

	userName := cmd.args[0]
	ctx := context.Background()
	_, err := s.db.GetUser(ctx, userName)
	if err != nil {
		return fmt.Errorf("user %s not found", userName)
	}

	err = s.SetUser(userName)
	if err != nil {
		return fmt.Errorf("failed to set user")
	}
	fmt.Printf("%s logged in\n", userName)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("invalid register command")
	}
	userName := cmd.args[0]
	ctx := context.Background()
	_, err := s.db.GetUser(ctx, userName)
	if err == nil {
		fmt.Printf("user %s already exists\n", userName)
		return fmt.Errorf("user %s already exists", userName)
	} else if err != sql.ErrNoRows {
		fmt.Printf("failed to check user %s: %v\n", userName, err)
		return fmt.Errorf("failed to check user %s", userName)
	}
	newUUID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	userParams := database.CreateUserParams{
		ID:        newUUID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Name:      userName,
	}

	user, err := s.db.CreateUser(ctx, userParams)
	if err != nil {
		return err
	}
	fmt.Printf("User %s created successfully! \n", userName)
	err = s.SetUser(user.Name)
	if err != nil {
		fmt.Printf("failed to set user")
		return fmt.Errorf("failed to set user")
	}
	return nil
}

func handlerReset(s *state, cmd command) error {
	if len(cmd.name) == 0 {
		fmt.Printf("invalid reset command")
		return fmt.Errorf("invalid reset command")
	}
	ctx := context.Background()
	err := s.db.DeleteUsers(ctx)
	if err != nil {
		fmt.Printf("Failed to delete users: %+v\n", err)
		return err
	}
	err = s.db.DeleteFeeds(ctx)
	if err != nil {
		fmt.Printf("Failed to delete feeds: %+v\n", err)
		return err
	}
	err = s.db.DeleteFollows(ctx)
	if err != nil {
		fmt.Printf("Failed to delete follows: %+v\n", err)
		return err
	}
	fmt.Println("All tables have been reset")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	if len(cmd.name) == 0 {
		fmt.Printf("invalid get users command")
		return fmt.Errorf("invalid get users command")
	}

	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		fmt.Printf("Failed to get users: %+v\n", err)
		return err
	}
	for _, user := range users {
		var userName string
		if user.Name == s.CurrentUser {
			userName = fmt.Sprintf("* %s (current)", user.Name)
		} else {
			userName = fmt.Sprintf("* %s", user.Name)
		}
		fmt.Println(userName)
	}
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	if feedURL == "" {
		return nil, fmt.Errorf("invalid feed URL")
	}
	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")
	res, err := client.Do(req)
	if err != nil || res.Body == nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var feed RSSFeed
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		return nil, err
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i, item := range feed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		feed.Channel.Item[i] = item
	}

	return &feed, nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.name) == 0 {
		fmt.Printf("invalid aggregation command")
		return fmt.Errorf("invalid aggregation command")
	}
	ctx := context.Background()
	feedURL := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(ctx, feedURL)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", feed)
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		fmt.Printf("invalid add feed command")
		return fmt.Errorf("invalid add feed command")
	}
	ctx := context.Background()
	user, err := s.db.GetUser(ctx, s.CurrentUser)
	if err != nil {
		fmt.Printf("Failed to get user")
		return err
	}
	userID := user.ID
	feedName := cmd.args[0]
	feedLink := cmd.args[1]
	_, err = fetchFeed(ctx, feedLink)
	if err != nil {
		fmt.Printf("Failed to fetch feed")
		return err
	}
	feedID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	feedParams := database.AddFeedParams{
		ID:        feedID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Name:      feedName,
		Url:       feedLink,
		UserID:    userID,
	}
	addedFeed, err := s.db.AddFeed(ctx, feedParams)
	if err != nil {
		fmt.Printf("Failed to add feed: %+v\n", err)
		return err
	}
	fmt.Printf("%s added by user %s\n", feedName, user.Name)
	followParams := database.CreateFeedFollowParams{
		UserID: userID,
		FeedID: addedFeed.ID,
	}
	following, err := s.db.CreateFeedFollow(ctx, followParams)
	fmt.Printf("%s is now following: %s\n", user.Name, following.FeedName)
	return nil
}

func handlerGetFeeds(s *state, cmd command) error {
	if len(cmd.name) == 0 {
		fmt.Printf("invalid get feed command")
	}
	ctx := context.Background()
	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		userID := feed.UserID
		userName, err := s.db.GetUserName(ctx, userID)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n%s\n%s\n", feed.Name, feed.Url, userName)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		fmt.Printf("invalid follow command")
		return fmt.Errorf("invalid follow command")
	}
	ctx := context.Background()
	user, err := s.db.GetUser(ctx, s.CurrentUser)

	if err != nil {
		fmt.Printf("Failed to get user ID: %+v\n", err)
		return err
	}
	feedURL := cmd.args[0]

	feed, err := s.db.GetFeedByURL(ctx, feedURL)
	if err != nil {
		fmt.Printf("Failed to get feed ID: %+v\n", err)
		return err
	}
	followParams := database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	follow, err := s.db.CreateFeedFollow(ctx, followParams)
	if err != nil {
		fmt.Printf("Failed to create follow: %+v\n", err)
		return err
	}
	fmt.Printf("%+v\n", follow)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.name) == 0 {
		fmt.Printf("invalid following command")
	}
	ctx := context.Background()
	user, err := s.db.GetUser(ctx, s.CurrentUser)
	if err != nil {
		fmt.Printf("Failed to get user ID: %+v\n", err)
		return err
	}
	userID := user.ID
	following, err := s.db.GetFeedFollowsForUser(ctx, userID)
	if err != nil {
		fmt.Printf("Failed to get feed follows for user: %+v\n", err)
		return err
	}
	for _, follow := range following {
		fmt.Printf("%s\n", follow.FeedName)
	}
	//	fmt.Printf("%+v\n", following)
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		fmt.Printf("invalid unfollow command")
		return fmt.Errorf("invalid unfollow command")
	}
	ctx := context.Background()
	user, err := s.db.GetUser(ctx, s.CurrentUser)
	if err != nil {
		fmt.Printf("Failed to get user ID: %+v\n", err)
	}
	userID := user.ID
	feedURL := cmd.args[0]
	feed, err := s.db.GetFeedByURL(ctx, feedURL)
	if err != nil {
		fmt.Printf("Failed to get feed ID: %+v\n", err)
	}
	feedID := feed.ID
	unFollowParams := database.UnFollowParams{
		UserID: userID,
		FeedID: feedID,
	}
	err = s.db.UnFollow(ctx, unFollowParams)
	if err != nil {
		fmt.Printf("Failed to unfollow: %+v\n", err)
	}
	fmt.Printf("%s unfollowed %s\n", user.Name, feed.Name)
	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
		os.Exit(1)
	}
	dbURL := cfg.DbUrl
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println("Error opening database:", err)
	}
	dbQueries := database.New(db)
	appState := &state{dbQueries, &cfg}

	appCommands := &commands{}
	appCommands.register("login", handlerLogin)
	appCommands.register("register", handlerRegister)
	appCommands.register("reset", handlerReset)
	appCommands.register("users", handlerGetUsers)
	appCommands.register("agg", handlerAgg)
	appCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	appCommands.register("feeds", handlerGetFeeds)
	appCommands.register("follow", middlewareLoggedIn(handlerFollow))
	appCommands.register("following", middlewareLoggedIn(handlerFollowing))
	appCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	args := os.Args[1:]
	if len(args) < 2 && args[0] == "login" {
		fmt.Println("Usage: Gator login {username}")
		os.Exit(1)
	}
	cmd := args[0]
	cmdArgs := args[1:]

	cmdName := command{name: cmd, args: cmdArgs}
	err = appCommands.run(appState, cmdName)
	if err != nil {
		os.Exit(1)
	}
}

/* Database Connect String:
"postgres://postgres:postgres@localhost:5432/gator"
*/
