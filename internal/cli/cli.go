package cli

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/quanchobi/gator/internal/config"
	"github.com/quanchobi/gator/internal/database"
	"github.com/quanchobi/gator/internal/parser"
)

type State struct {
	Cfg *config.Config
	Db  *database.Queries
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Cmds map[string]func(*State, Command) error
}

func GetFunctions() map[string]func(*State, Command) error {
	return map[string]func(*State, Command) error{
		"login":     HandlerLogin,
		"register":  HandlerRegister,
		"reset":     HandlerReset,
		"users":     HandlerUsers,
		"agg":       HandlerAggregate,
		"feeds":     HandlerPrintFeeds,
		"addfeed":   MiddlewareLoggedIn(HandlerAddFeed),
		"follow":    MiddlewareLoggedIn(HandlerFollow),
		"following": MiddlewareLoggedIn(HandlerFollowing),
		"unfollow":  MiddlewareLoggedIn(HandlerUnfollow),
	}
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		fmt.Println("login expects one argument, the username.")
		os.Exit(1)
	}
	username := cmd.Args[0]
	if username == "" {
		fmt.Println("login expects one argument, the username.")
		os.Exit(1)
	}
	// check if user is in database
	user, err := s.Db.GetUser(context.Background(), username)
	if err != nil {
		return err
	}
	err = s.Cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("%s has logged in successfully.\n", username)
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		fmt.Println("register expects one argument, the username")
		os.Exit(1)
	}
	username := cmd.Args[0]
	if username == "" {
		// should probably move all of these errors into main for proper error prop
		fmt.Println("register expects one argument, the username.")
		os.Exit(1)
	}

	user, err := s.Db.CreateUser(context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      username,
		},
	)
	if err != nil {
		return err
	}

	err = s.Cfg.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("%s has registered successfully.\n", user.Name)
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	if len(cmd.Args) != 0 {
		fmt.Println("reset takes no arguments")
		os.Exit(1)
	}
	err := s.Db.Reset(context.Background())
	if err != nil {
		fmt.Printf("table was not reset\n")
		return err
	}
	fmt.Printf("table was reset successfully\n")
	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	users, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user.Name == s.Cfg.CurrentUserName {
			fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}
	}
	return nil
}

func HandlerAggregate(s *State, cmd Command) error {
	if len(cmd.Args) != 1 {
		fmt.Printf("agg expects 1 argument, the interval at which to aggregate")
	}

	fmt.Println("Fetching every ", cmd.Args[0])
	interval, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	for ; ; <-ticker.C {
		err = scrapeFeeds(s)
		if err != nil {
			return err
		}
	}
}

func scrapeFeeds(s *State) error {
	nextFeedSlice, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}
	nextFeed := nextFeedSlice[0]
	feedURL := nextFeed.Url

	fetchedFeed, err := parser.FetchFeed(context.Background(), feedURL)
	if err != nil {
		return err
	}

	s.Db.MarkFeedFetched(context.Background(),
		database.MarkFeedFetchedParams{
			ID: nextFeed.ID,
			LastFetchedAt: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
			UpdatedAt: time.Now(),
		},
	)

	fmt.Println(fetchedFeed.Channel.Title)
	fmt.Println(fetchedFeed.Channel.Link)
	fmt.Println(fetchedFeed.Channel.Description)
	fmt.Println(fetchedFeed.Channel.Item)

	return nil
}

func HandlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 2 {
		fmt.Println("addfeed takes two arguments: name of the feed and url")
		os.Exit(1)
	}
	feedName := cmd.Args[0]
	url := cmd.Args[1]

	feed, err := s.Db.CreateFeed(context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      feedName,
			Url:       url,
			LastFetchedAt: sql.NullTime{
				Time:  time.Now(),
				Valid: false,
			},
			UserID: user.ID,
		},
	)
	if err != nil {
		fmt.Printf("Error in creating feed %v \n", err)
		return err
	}

	feedFollow, err := s.Db.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{
			ID:     uuid.New(),
			UserID: user.ID,
			FeedID: feed.ID,
		},
	)
	if err != nil {
		fmt.Println("Error in creating feed follows")
		return err
	}

	fmt.Printf("Created feed record: %v, at: %v, %v (%v), for user %v\n", feed.ID, feed.CreatedAt, feed.Name, feed.Url, feed.UserID)
	fmt.Printf("Created following record: %v, for user %v and feed %v\n", feedFollow.ID, feedFollow.UserID, feedFollow.FeedID)

	return nil
}

func HandlerPrintFeeds(s *State, cmd Command) error {
	if len(cmd.Args) != 0 {
		fmt.Println("feeds takes no arguments")
		os.Exit(1)
	}
	feeds, err := s.Db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Printf("%v, %v: %v\n", feed.Username, feed.Name, feed.Url)
	}
	return nil
}

func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		fmt.Println("follow takes one argument: the URL")
		os.Exit(1)
	}

	feed, err := s.Db.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}

	_, err = s.Db.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{
			ID:     uuid.New(),
			UserID: user.ID,
			FeedID: feed.ID,
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("feed name: %v, user name: %v\n", feed.Name, user.Name)
	return nil
}

func HandlerFollowing(s *State, cmd Command, user database.User) error {
	follows, err := s.Db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	fmt.Printf("%s is following:\n", user.Name)
	for _, follow := range follows {
		fmt.Printf("%s (%s)\n", follow.Feedname, follow.Url)
	}
	return nil
}

func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) != 1 {
		fmt.Printf("unfollow expects 1 argument, the URL of the feed that is being unfollowed")
	}

	feed, err := s.Db.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}

	s.Db.DeleteFeedFollow(context.Background(),
		database.DeleteFeedFollowParams{
			UserID: user.ID,
			FeedID: feed.ID,
		},
	)
	return nil
}

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		user, err := s.Db.GetUser(context.Background(), s.Cfg.CurrentUserName)
		if err != nil {
			return err
		}

		handler(s, cmd, user)
		return nil
	}
}

func (c *Commands) Run(s *State, cmd Command) error {
	fn, ok := c.Cmds[cmd.Name]
	if !ok {
		fmt.Printf("command %s not found\n", cmd.Name)
	}
	err := fn(s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *Commands) Register(name string, f func(*State, Command) error) error {
	_, exists := c.Cmds[name]
	if exists {
		return fmt.Errorf("attempted re-register of function signature %T", f) // prolly shouldnt happen
	}
	c.Cmds[name] = f
	return nil
}
