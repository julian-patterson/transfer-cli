package cmd

import (
    "fmt"
    "log"
    "os"

    "github.com/spf13/cobra"
    "go-cli/sshutils"
    "gopkg.in/yaml.v2"
    "io/ioutil"
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

var rootCmd = &cobra.Command{
    Use:   "transfercli",
    Short: "A CLI tool for SSH operations",
    Long:  `A CLI tool for listing and transferring files over SSH.`,
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

var transferCmd = &cobra.Command{
    Use:   "transfer",
    Short: "Transfer files from a remote directory to a local directory",
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
        if localDir == "" {
            localDir = config.LocalDir
            if localDir == "" {
                log.Fatalf("\033[31mMissing required configuration: localDir.\nPlease provide the localDir flag or set the localDir using set command\033[0m")
            }
        }

        client, err := sshutils.SshConnect(user, password, host, port)
        if err != nil {
            log.Fatalf("\033[31mSSH connection failed: %v\033[0m", err)
        }
        defer client.Close()

        err = sshutils.TransferFiles(client, remoteDir, localDir)
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
        err = ioutil.WriteFile("config.yaml", data, 0644)
        if err != nil {
            log.Fatalf("\033[31mFailed to write config file: %v\033[0m", err)
        }
        fmt.Println("\033[32mConfiguration updated in config.yaml\033[0m")
    },
}

func loadConfig() Config {
    var config Config
    data, err := ioutil.ReadFile("config.yaml")
    if err != nil {
        log.Fatalf("\033[31mFailed to read config file: %v\033[0m", err)
    }
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        log.Fatalf("\033[31mFailed to unmarshal config: %v\033[0m", err)
    }
    return config
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
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}