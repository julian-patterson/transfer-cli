package cmd

import (
    "fmt"
    "log"
    "os"

    "github.com/spf13/cobra"
    "../sshutils"
)

var (
    user      string
    password  string
    host      string
    port      int
    remoteDir string
    localDir  string
)

var rootCmd = &cobra.Command{
    Use:   "yourproject",
    Short: "A CLI tool for SSH operations",
    Long:  `A CLI tool for listing and transferring files over SSH.`,
}

var listCmd = &cobra.Command{
    Use:   "list",
    Short: "List files in a remote directory",
    Run: func(cmd *cobra.Command, args []string) {
        client, err := sshutils.SshConnect(user, password, host, port)
        if err != nil {
            log.Fatalf("SSH connection failed: %v", err)
        }
        defer client.Close()

        err = sshutils.ListFilesInRemoteDir(client, remoteDir)
        if err != nil {
            log.Fatalf("Failed to list files in remote directory: %v", err)
        }
    },
}

var transferCmd = &cobra.Command{
    Use:   "transfer",
    Short: "Transfer files from a remote directory to a local directory",
    Run: func(cmd *cobra.Command, args []string) {
        client, err := sshutils.SshConnect(user, password, host, port)
        if err != nil {
            log.Fatalf("SSH connection failed: %v", err)
        }
        defer client.Close()

        err = sshutils.TransferFiles(client, remoteDir, localDir)
        if err != nil {
            log.Fatalf("Failed to transfer files: %v", err)
        }
    },
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
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}