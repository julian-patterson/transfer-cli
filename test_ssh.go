package main

import (
    "flag"
    "fmt"
    "log"
    "golang.org/x/crypto/ssh"
	"bytes"
)

func sshConnect(user, password, host string, port int) (*ssh.Client, error) {
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

func listFilesInRemoteDir(client *ssh.Client, remoteDir string) error {
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



func main() {
    user := flag.String("user", "", "SSH username")
    password := flag.String("password", "", "SSH password")
    host := flag.String("host", "", "SSH host")
    port := flag.Int("port", 22, "SSH port")
	remoteDir := flag.String("remoteDir", "", "Remote directory to list files")
    flag.Parse()

    if *user == "" || *password == "" || *host == "" {
        log.Fatal("Please provide SSH username, password, and host")
    }

    client, err := sshConnect(*user, *password, *host, *port)
    if err != nil {
        log.Fatalf("SSH connection failed: %v", err)
    }
    defer client.Close()

    log.Println("SSH connection established successfully")

	log.Println("SSH connection established successfully")

    err = listFilesInRemoteDir(client, *remoteDir)
    if err != nil {
        log.Fatalf("Failed to list files in remote directory: %v", err)
    }
}