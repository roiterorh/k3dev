package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
)

func init() {
	rootCmd.AddCommand(versionsCmd)
}



func versions_from_dockerhub() []string {
	url := "https://registry.hub.docker.com/v2/repositories/rancher/k3s/tags?page_size=1024"
	resp, err := http.Get(url)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("No response from request")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	result := gjson.GetBytes(body, "results.#.name")
	re := regexp.MustCompile("^(v\\d.\\d+.\\d+-k3s\\d+)$")
	var arr []string
	result.ForEach(func(key, value gjson.Result) bool {
		if re.Match([]byte(value.String())) {
			arr = append(arr, value.String())
		}
		return true
	})
	sort.Slice(arr, func(i, j int) bool { return arr[i] > arr[j] })
	return arr
}

var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Print the version number of Hugo",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		arr := versions_from_dockerhub()
		if viper.GetBool("json") {
			log.SetFormatter(&log.JSONFormatter{})
		}
		log.WithFields(log.Fields{
			"versions": arr,
		}).Info("Available kubernetes versions")
	},
}
