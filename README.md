# Gator

An RSS feed aggre*gator*

## Install

Install dependencies

```
apt install postgresql
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Create the database

`sudo -u postgres psql`

Then, in postgres

`CREATE DATABASE gator;`

Set the password if you wish

`ALTER USE postgres PASSWORD '<password>'`

To exit postgres

`\q`

Install the gator binary

`go install github.com/quanchobi/gator`

Clone the repository (you have to do this for goose)

`git clone github.com/quanchobi/gator`

`cd gator/sql/schema`

Create the necessary tables with the postgres connection string

`goose postgres "postgres://<postgres user>:@localhost:<port (default 5432)>/gator up`

Create `~/.gatorconfig.json`, with the following contents, where `connection string` is the value you used to connect with goose.

`{"db_url":<connection string>?sslmode=disable}`

## Usage

To reset the database

`gator reset`

To add a user (also logs in as that user)

`gator register <username>`

To login as a different user

`gator login <username>`

To list all available users

`gator users`

To add a feed (automatically follows the added feed for the logged-in user)

`gator addfeed <name> <url>`

To show all available feeds

`gator feeds`

To follow a feed

`gator follow <url>`

To unfollow a feed

`gator unfollow <url>`

To browse posts from followed feeds

`gator browse <limit> # limit is optional, default is 2`
