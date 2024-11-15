package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"go-cli/sshutils"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	user      string
	password  string
	host      string
	port      int
	remoteDir string
	localDir  string
)

type Config struct {
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	RemoteDir string `yaml:"remoteDir"`
	LocalDir  string `yaml:"localDir"`
}

type Params struct {
	User      string
	Password  string
	Host      string
	Port      int
	RemoteDir string
	LocalDir  string
}

var rootCmd = &cobra.Command{
	Use:   "transfercli",
	Short: "A CLI tool for SSH operations",
	Long:  `A CLI tool for listing and transferring files over SSH.`,
}

func checkParams(Params Params) Params {
	config := loadConfig()
	if Params.User == "" {
		Params.User = config.User
		if user == "" {
			log.Fatalf("\033[31mMissing required configuration: user.\nPlease provide the user flag or set the user using set command\033[0m")
		}
	}
	if Params.Password == "" {
		Params.Password = config.Password
		if password == "" {
			log.Fatalf("\033[31mMissing required configuration: password.\nPlease provide the password flag or set the password using set command\033[0m")
		}
	}
	if Params.Host == "" {
		Params.Host = config.Host
		if host == "" {
			log.Fatalf("\033[31mMissing required configuration: host.\nPlease provide the host flag or set the host using set command\033[0m")
		}
	}
	if Params.Port == 0 {
		Params.Port = config.Port
	}
	if Params.RemoteDir == "" {
		Params.RemoteDir = config.RemoteDir
		if remoteDir == "" {
			log.Fatalf("\033[31mMissing required configuration: remoteDir.\nPlease provide the remoteDir flag or set the remoteDir using set command\033[0m")
		}
	}
	if Params.LocalDir == "" {
		Params.LocalDir = config.LocalDir
		if localDir == "" {
			log.Fatalf("\033[31mMissing required configuration: localDir.\nPlease provide the localDir flag or set the localDir using set command\033[0m")
		}
	}
	return Params
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List files in a remote directory",
	Run: func(cmd *cobra.Command, args []string) {
		config := loadConfig()
		if user == "" {
			user = config.User
			if user == "" {
				log.Fatalf("\033[31mMissing required configuration: user.\nPlease provide the user flag or set the user using set command\033[0m")
			}
		}
		if password == "" {
			password = config.Password
			if password == "" {
				log.Fatalf("\033[31mMissing required configuration: password.\nPlease provide the password flag or set the password using set command\033[0m")
			}
		}
		if host == "" {
			host = config.Host
			if host == "" {
				log.Fatalf("\033[31mMissing required configuration: host.\nPlease provide the host flag or set the host using set command\033[0m")
			}
		}
		if port == 0 {
			port = config.Port
		}
		if remoteDir == "" {
			remoteDir = config.RemoteDir
			if remoteDir == "" {
				log.Fatalf("\033[31mMissing required configuration: remoteDir.\nPlease provide the remoteDir flag or set the remoteDir using set command\033[0m")
			}
		}

		client, err := sshutils.SshConnect(user, password, host, port)
		if err != nil {
			log.Fatalf("\033[31mSSH connection failed: %v\033[0m", err)
		}
		defer client.Close()

		err = sshutils.ListFilesInRemoteDir(client, remoteDir)
		if err != nil {
			log.Fatalf("\033[31mFailed to list files in remote directory: %v\033[0m", err)
		}
		fmt.Println("\033[32mSuccessfully listed files in remote directory\033[0m")
	},
}

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit files to staging area. Fetches all files in subdirectories.",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		commitFile := args
		var committedFiles []string

		for _, file := range commitFile {
			// Check if the file exists in the current directory
			if _, err := os.Stat(file); err == nil {
				fmt.Println("\033[32m ---> Committing file:", file, "\033[0m")
				committedFiles = append(committedFiles, file)
				continue
			} else if os.IsNotExist(err) {
				// If the file does not exist in the current directory, check in subdirectories
				found := false
				err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() && filepath.Base(path) == file {
						found = true
						fmt.Println("\033[32m ---> Committing file:", path, "\033[0m")
						committedFiles = append(committedFiles, path)
						return filepath.SkipDir
					}
					return nil
				})
				if err != nil {
					fmt.Println("\033[31m ---> Error while walking through directories: ", err, "\033[0m")
				}

				if !found {
					fmt.Println("\033[31m ---> File not found:", file, "\033[0m")
				}
			} else {
				fmt.Println("\033[31m ---> Error checking file:", file, "\033[0m")
			}
		}

		// Create or open up Os file
		file, err := os.OpenFile(".committed_files", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Println("\033[31m ---> Failed to open .committed_files \033[0m")
		}
		defer file.Close()

		for _, committedFile := range committedFiles {
			if _, err := file.WriteString(committedFile + "\n"); err != nil {
				fmt.Println("\033[31m ---> Failed to add committed files \033[0m")
			}
		}
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync files to remote directory using commit function",
	Run: func(cmd *cobra.Command, args []string) {
		params := checkParams(Params{user, password, host, port, remoteDir, localDir})

		data, err := os.ReadFile(".committed_files")
		if err != nil {
			fmt.Println("\033[31m ---> Failed to read .committed_files \033[0m")
		}

		client, err := sshutils.SshConnect(user, password, host, port)
		if err != nil {
			log.Fatalf("SSH connection failed: %v", err)
		}
		defer client.Close()

		files := strings.Split((string(data)), "\n")
		if len(files) == 0 || (len(files) == 1 && files[0] == "") {
			fmt.Println("\033[31m ---> No files to sync \033[0m")
			fmt.Println("\033[31m ---> Please add files to commit stage using `commit` command \033[0m")
			return
		}

		for _, file := range files {
			if file != "" {
				err := sshutils.TransferFiles(client, file, params.RemoteDir)
				if err != nil {
					fmt.Printf("\033[31m ---> Failed to synced file: %s \033[0m\n", file)
				} else {
					fmt.Printf("\033[32m ---> Successfully synced file: %s \033[0m\n", file)
				}
			}
		}

		err = os.WriteFile(".committed_files", []byte(""), 0644)
		if err != nil {
			fmt.Println("\033[31m ---> Failed to clear .committed_files \033[0m")
		} else {
			// fmt.Printf("\033[32m ---> Successfully synced all files over to %s in %s \033[0m\n", params.Host, params.RemoteDir)
		}
	},
}

var transferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Directly transfer files from a remote directory to a local directory",
	Run: func(cmd *cobra.Command, args []string) {
		params := checkParams(Params{user, password, host, port, remoteDir, localDir})

		client, err := sshutils.SshConnect(params.User, params.Password, params.Host, params.Port)
		if err != nil {
			log.Fatalf("\033[31mSSH connection failed: %v\033[0m", err)
		}
		defer client.Close()

		err = sshutils.TransferFiles(client, params.RemoteDir, params.LocalDir)
		if err != nil {
			log.Fatalf("\033[31mFailed to transfer files: %v\033[0m", err)
		}
		fmt.Println("\033[32mSuccessfully transferred files\033[0m")
	},
}

var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set SSH configuration",
	Run: func(cmd *cobra.Command, args []string) {
		config := loadConfig()

		if user != "" {
			config.User = user
		}
		if password != "" {
			config.Password = password
		}
		if host != "" {
			config.Host = host
		}
		if port != 0 {
			config.Port = port
		}
		if remoteDir != "" {
			config.RemoteDir = remoteDir
		}
		if localDir != "" {
			config.LocalDir = localDir
		}

		data, err := yaml.Marshal(&config)
		if err != nil {
			log.Fatalf("\033[31mFailed to marshal config: %v\033[0m", err)
		}
		err = os.WriteFile("config.yaml", data, 0644)
		if err != nil {
			log.Fatalf("\033[31mFailed to write config file: %v\033[0m", err)
		}
		fmt.Println("\033[32mConfiguration updated in config.yaml\033[0m")
	},
}

func loadConfig() Config {
	var config Config
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("\033[31mFailed to read config file: %v\033[0m", err)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("\033[31mFailed to unmarshal config: %v\033[0m", err)
	}
	return config
}

func initCommittedFile() error {
	file, err := os.OpenFile(".committed_files", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "SSH username")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "SSH password")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "", "SSH host")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "P", 22, "SSH port")
	rootCmd.PersistentFlags().StringVarP(&remoteDir, "remoteDir", "r", "", "Remote directory")
	rootCmd.PersistentFlags().StringVarP(&localDir, "localDir", "l", "", "Local directory")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(transferCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(commitCmd)
	rootCmd.AddCommand(syncCmd)

	if err := initCommittedFile(); err != nil {
		log.Fatalf("\033[31mFailed to initialize committed file tracker: %v\033[0m", err)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
