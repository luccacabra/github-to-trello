package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/luccacabra/github-to-trello/github"
	"github.com/luccacabra/github-to-trello/syncer"
	githubSync "github.com/luccacabra/github-to-trello/syncer/github"
	"github.com/luccacabra/github-to-trello/trello"

	"github.com/spf13/viper"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile = kingpin.Flag("config.file", "github-to-trello configuration file.").Default("github-to-trello.yaml").String()
)

func main() {
	// pls don't store secrets in config
	ghAPIToken := viper.GetString("GH_APITOKEN")
	trelloKey := viper.GetString("TRELLO_KEY")
	trelloToken := viper.GetString("TRELLO_TOKEN")

	// load config file
	kingpin.Parse()
	configFileBaseName := filepath.Base(*configFile)

	viper.SetConfigName(strings.TrimSuffix(configFileBaseName, filepath.Ext(configFileBaseName)))
	viper.AddConfigPath(filepath.Dir(*configFile))

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	trelloClient := trello.NewClient(
		trelloKey,
		trelloToken,
		trello.ClientConfig{
			BoardName:     viper.GetString("trello_board_name"),
			LabelCardName: viper.GetString("trello_label_card_name"),
			LabelMap:      viper.GetStringMapString("trello_label_map"),
		},
	)

	ghClient := github.NewClient(
		ghAPIToken,
		github.Config{
			OrgName:  viper.GetString("github_org_name"),
			UserName: viper.GetString("github_user_name"),
		},
	)

	conf := &syncer.Config{}
	err = viper.UnmarshalKey("config", conf)

	issueSyncer := githubSync.NewIssueSyncer(ghClient, trelloClient, conf.Issue)
	if err = issueSyncer.Sync(); err != nil {
		log.Fatal(err)
	}
}
