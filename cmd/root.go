/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile     string
	providerUrl string
	logger      *zap.SugaredLogger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "eth-contracts-go",
	Short: "CLI application ",
	Long:  ``,
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.eth-contracts-go.yaml)")
	rootCmd.PersistentFlags().StringVar(&providerUrl, "provider-url", "", "Ethereum provider url")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("env", "e", "dev", "Environment to use. Not actually does anything")
	rootCmd.Flags().Bool("log-json", false, "Enable json logger output")
	rootCmd.Flags().Bool("log-stacktrace", false, "Enable logger error stacktrace output")

	viper.BindPFlag("provider_url", rootCmd.PersistentFlags().Lookup("provider-url"))
	viper.BindPFlag("env", rootCmd.Flags().Lookup("env"))
	viper.BindPFlag("log_json", rootCmd.Flags().Lookup("log-json"))
	viper.BindPFlag("log_stacktrace", rootCmd.Flags().Lookup("log-stacktrace"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".eth-contracts-go" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".eth-contracts-go")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}

	if err := initLogger(); err != nil {
		log.Fatal(err)
	}
}

func initLogger() error {
	cfg := zap.NewDevelopmentConfig()
	cfg.Development = viper.GetString("env") != "main"
	cfg.DisableStacktrace = !viper.GetBool("log_stacktrace")

	if viper.GetBool("log_json") {
		cfg.Encoding = "json"
	} else {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	l, err := cfg.Build()
	if err != nil {
		return err
	}
	logger = l.Sugar()

	return nil
}
