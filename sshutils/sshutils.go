package sshutils

import (
    "bytes"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"

    "golang.org/x/crypto/ssh"
)

func SshConnect(user, password, host string, port int) (*ssh.Client, error) {
    config := &ssh.ClientConfig{
        User: user,
        Auth: []ssh.AuthMethod{
            ssh.Password(password),
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    }
    address := fmt.Sprintf("%s:%d", host, port)
    client, err := ssh.Dial("tcp", address, config)
    if err != nil {
        log.Printf("Failed to connect to SSH server at %s: %v", address, err)
        return nil, err
    }
    log.Printf("Successfully connected to SSH server at %s", address)
    return client, nil
}

func ListFilesInRemoteDir(client *ssh.Client, remoteDir string) error {
    session, err := client.NewSession()
    if err != nil {
        return fmt.Errorf("failed to create SSH session: %v", err)
    }
    defer session.Close()

    var b bytes.Buffer
    session.Stdout = &b
    cmd := fmt.Sprintf("ls -l %s", remoteDir)
    if err := session.Run(cmd); err != nil {
        return fmt.Errorf("failed to list files in directory %s: %v", remoteDir, err)
    }

    fmt.Println("Files in remote directory:")
    fmt.Println(b.String())
    return nil
}

func TransferFiles(client *ssh.Client, remoteDir, localDir string) error {
    session, err := client.NewSession()
    if err != nil {
        return fmt.Errorf("failed to create SSH session: %v", err)
    }
    defer session.Close()

    var b bytes.Buffer
    session.Stdout = &b
    cmd := fmt.Sprintf("find %s -type f", remoteDir)
    if err := session.Run(cmd); err != nil {
        return fmt.Errorf("failed to list files in directory %s: %v", remoteDir, err)
    }

    files := strings.Split(b.String(), "\n")
    for _, file := range files {
        if file == "" {
            continue
        }
        localFilePath := filepath.Join(localDir, strings.TrimPrefix(file, remoteDir))

        log.Printf("Transferring file from %s to %s", file, localFilePath)
        err := CopyFileFromRemote(client, file, localFilePath)
        if err != nil {
            return fmt.Errorf("failed to transfer file %s: %v", file, err)
        }

        fmt.Printf("Successfully transferred %s to %s\n", file, localFilePath)
    }

    return nil
}

func CopyFileFromRemote(client *ssh.Client, remoteFilePath, localFilePath string) error {
    session, err := client.NewSession()
    if err != nil {
        return fmt.Errorf("failed to create SSH session: %v", err)
    }
    defer session.Close()

    // Ensure the local directory exists
    localDir := filepath.Dir(localFilePath)
    if err := os.MkdirAll(localDir, os.ModePerm); err != nil {
        return fmt.Errorf("failed to create local directory %s: %v", localDir, err)
    }

    outFile, err := os.Create(localFilePath)
    if err != nil {
        return fmt.Errorf("failed to create local file %s: %v", localFilePath, err)
    }
    defer outFile.Close()

    session.Stdout = outFile
    if err := session.Run(fmt.Sprintf("cat %s", remoteFilePath)); err != nil {
        return fmt.Errorf("failed to copy file %s: %v", remoteFilePath, err)
    }

    return nil
}