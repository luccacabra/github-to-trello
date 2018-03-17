package main

import (
	"fmt"

	"github.com/luccacabra/aws-github-to-trello/github"
	"github.com/luccacabra/aws-github-to-trello/syncer"
	"github.com/luccacabra/aws-github-to-trello/syncer/open"
	"github.com/luccacabra/aws-github-to-trello/trello"

	"github.com/spf13/viper"
)

func main() {
	// pls don't store secrets in config
	ghAPIToken := viper.GetString("GH_APITOKEN")
	trelloKey := viper.GetString("TRELLO_KEY")
	trelloToken := viper.GetString("TRELLO_TOKEN")

	viper.SetConfigName("local")
	viper.AddConfigPath("config/")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	trelloClient := trello.NewClient(
		trelloKey,
		trelloToken,
		trello.Config{
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

	openIssueSyncer := open.NewIssueSyncer(ghClient, trelloClient.Board, conf.Open.Types.Issue)
	openIssueSyncer.Sync()
}
