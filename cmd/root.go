package cmd

import (
	"fmt"
	"github.com/mbndr/figlet4go"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	// Used for flags.
	cfgFile     string
	userLicense string
	rootCmd     = &cobra.Command{
		Use:   "cobra",
		Short: "A generator for Cobra based Applications",
		Long: `Cobra is a CLI library for Go that empowers applications.
  This application is a tool to generate the needed files
  to quickly create a Cobra application.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default to config.yml)")
	rootCmd.PersistentFlags().StringP("kubeconfig", "k", "~/.kube/config", "path for binded kubernetes config")
	rootCmd.PersistentFlags().StringP("cluster-name", "n", "k3dev", "kubernetes cluster name")
	rootCmd.PersistentFlags().StringP("kubernetes-version", "v", "latest", "kubernetes version to run")
	rootCmd.PersistentFlags().String("domain", "localtest.me", "wildcard domain to add to cert")
	rootCmd.PersistentFlags().Bool("json", false, "logs in json format ")
	rootCmd.PersistentFlags().Bool("certificate", true, "create and deploy SSL certificate and add it to trust store")
	viper.BindPFlag("cluster-name", rootCmd.PersistentFlags().Lookup("cluster-name"))
	viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
	viper.BindPFlag("kubernetes-version", rootCmd.PersistentFlags().Lookup("kubernetes-version"))
	viper.BindPFlag("json", rootCmd.PersistentFlags().Lookup("json"))
	viper.BindPFlag("certificate", rootCmd.PersistentFlags().Lookup("certificate"))
	viper.BindPFlag("domain", rootCmd.PersistentFlags().Lookup("domain"))
	//   client.containers.run("vishnunair/docker-mkcert",environment=["domain="+args["--wildcard"]+",*."+args["--wildcard"]],name="mkcert-k3dev",detach=True,auto_remove=True,volumes={ script_path +'/certs': {'bind': '/root/.local/share/mkcert', 'mode': 'rw'}})
	//   while not os.path.exists(script_path +'/certs/'+args["--wildcard"]+"-key.pem") or not os.path.exists(script_path +'/certs/'+args["--wildcard"]+".pem"):
	// 	time.sleep(1)
}
func er(msg interface{}) {
	fmt.Println("Error:", msg)
	os.Exit(1)
}
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			er(err)
		}
		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home + "/.k3dev")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		//   fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
	if !(viper.GetBool("json")) {
		ascii := figlet4go.NewAsciiRender()
		options := figlet4go.NewRenderOptions()
		options.FontName = "larry3d"
		options.FontColor = []figlet4go.Color{
			figlet4go.ColorGreen,
			figlet4go.ColorCyan,
		}
		renderStr, _ := ascii.RenderOpts("k3dev", options)
		fmt.Print(renderStr)
	}
}
