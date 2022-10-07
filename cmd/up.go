package cmd

import (
	"fmt"
    // "path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// "errors"

	// "os"
)

func init() {
	rootCmd.AddCommand(upCmd)
	
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Print the version number of Hugo",
	Long:  `All software has versions. This is Hugo's`,
	Run: func(cmd *cobra.Command, args []string) {
		kubeconfig := NormalizePath(viper.GetString("kubeconfig"))
		// certificate := viper.GetString("certificate")
		version := viper.GetString("kubernetes-version")
		domain := viper.GetString("domain")
		name := viper.GetString("cluster-name")
		fmt.Println(kubeconfig)
		fmt.Println(version)




		clusterUp(name, kubeconfig, version)

		createCertificates(domain)

		healthAPI(kubeconfig)
		tlsSecret(kubeconfig)
		serverReady(kubeconfig)

			applyManifests(kubeconfig,domain)
	},
}
