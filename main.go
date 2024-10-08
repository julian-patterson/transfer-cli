package main

import (
    "bytes"
    "flag"
    "fmt"
    "html/template"
    "io"
    "log"
    "mime/multipart"
    "net/http"
    "os"
    "path/filepath"
    "time"
    "golang.org/x/crypto/ssh"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
    // render the index page template
    t, err := template.ParseFiles("index.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    t.Execute(w, nil)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        t, err := template.ParseFiles("index.html")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        t.Execute(w, nil)
    } else if r.Method == "POST" {
        // parse input, type multipart/form-data
        err := r.ParseMultipartForm(32 << 20) // limit your max input length!
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        // retrieve the file from form data
        file, header, err := r.FormFile("file")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer file.Close()

        // create a new file in the uploads directory
        err = os.MkdirAll("./uploads", os.ModePerm)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        dst, err := os.Create("./uploads/" + header.Filename)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        defer dst.Close()

        // copy the uploaded file to the destination file
        if _, err := io.Copy(dst, file); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        fmt.Fprintf(w, "File uploaded successfully!")
    } else {
        http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
    }
}

func uploadFilesFromDir(dir string, url string) error {
    log.Printf("Uploading file %s to %s\n", dir, url)

    files, err := os.ReadDir(dir)
    if err != nil {
        return err
    }

    for _, file := range files {
        if !file.IsDir() {
            err := uploadFile(filepath.Join(dir, file.Name()), url)
            if err != nil {
                return err
            }
        }
    }
    return nil
}

func uploadFile(filePath string, url string) error {
    log.Printf("Uploading file %s to %s\n", filePath, url)
    file, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    part, err := writer.CreateFormFile("file", filepath.Base(filePath))
    if err != nil {
        return err
    }
    _, err = io.Copy(part, file)
    if err != nil {
        return err
    }
    writer.Close()

    req, err := http.NewRequest("POST", url, body)
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to upload file: %s", resp.Status)
    }

    return nil
}

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

func listFiles(client *ssh.Client, dir string) (map[string]time.Time, error) {
    session, err := client.NewSession()
    if err != nil {
        return nil, fmt.Errorf("failed to create SSH session: %v", err)
    }
    defer session.Close()

    var b bytes.Buffer
    session.Stdout = &b
    cmd := fmt.Sprintf("find %s -type f -printf '%%T@ %%p\\n'", dir)
    if err := session.Run(cmd); err != nil {
        return nil, fmt.Errorf("failed to list files in directory %s: %v", dir, err)
    }

    files := make(map[string]time.Time)
    for _, line := range bytes.Split(b.Bytes(), []byte("\n")) {
        if len(line) == 0 {
            continue
        }
        var timestamp float64
        var name string
        _, err := fmt.Sscanf(string(line), "%f %s", &timestamp, &name)
        if err == nil {
            modTime := time.Unix(int64(timestamp), 0)
            files[name] = modTime
        }
    }
    return files, nil
}

func syncDirectories(localDir, remoteDir, user, password, host string, port int) error {
    log.Printf("Syncing local directory %s with remote directory %s\n", localDir, remoteDir)
    client, err := sshConnect(user, password, host, port)
    if err != nil {
        return fmt.Errorf("failed to connect to SSH: %v", err)
    }
    defer client.Close()
    log.Println("Connected to SSH server")
    localFiles, err := listLocalFiles(localDir)
    if err != nil {
        return fmt.Errorf("failed to list local files: %v", err)
    }

    remoteFiles, err := listFiles(client, remoteDir)
    if err != nil {
        return fmt.Errorf("failed to list remote files: %v", err)
    }

    log.Println("Comparing files and syncing...")

    for name, localModTime := range localFiles {
        if remoteModTime, exists := remoteFiles[name]; !exists || localModTime.After(remoteModTime) {
            log.Printf("Uploading local file %s to remote", name)
            err := uploadFile(filepath.Join(localDir, name), "http://localhost:8080/upload")
            if err != nil {
                return fmt.Errorf("failed to upload file %s: %v", name, err)
            }
        }
    }

    for name, remoteModTime := range remoteFiles {
        if localModTime, exists := localFiles[name]; !exists || remoteModTime.After(localModTime) {
            log.Printf("Downloading remote file %s to local", name)
            err := downloadFile(client, filepath.Join(remoteDir, name), filepath.Join(localDir, name))
            if err != nil {
                return fmt.Errorf("failed to download file %s: %v", name, err)
            }
        }
    }

    log.Println("Sync complete")
    return nil
}

func listLocalFiles(dir string) (map[string]time.Time, error) {
    log.Printf("Listing files in directory %s\n", dir)
    files := make(map[string]time.Time)

    entries, err := os.ReadDir(dir)
    if err != nil {
        return nil, fmt.Errorf("failed to read directory %s: %v", dir, err)
    }

    for _, entry := range entries {
        if !entry.IsDir() {
            info, err := entry.Info()
            if err != nil {
                return nil, fmt.Errorf("failed to get info for file %s: %v", entry.Name(), err)
            }
            files[entry.Name()] = info.ModTime()
            log.Printf("Found file: %s, modified at: %s\n", entry.Name(), info.ModTime().Format(time.RFC3339))
        }
    }
    return files, nil
}

func downloadFile(client *ssh.Client, remotePath, localPath string) error {
    session, err := client.NewSession()
    if err != nil {
        return err
    }
    defer session.Close()

    outFile, err := os.Create(localPath)
    if err != nil {
        return err
    }
    defer outFile.Close()

    session.Stdout = outFile
    if err := session.Run(fmt.Sprintf("cat %s", remotePath)); err != nil {
        return err
    }

    return nil
}

func main() {
    name := flag.String("name", "world", "a name to say hello to")
    verbose := flag.Bool("verbose", false, "enable verbose mode")
    uploadDir := flag.String("uploadDir", "", "directory to upload files from")
    remoteDir := flag.String("remoteDir", "", "remote directory to sync with")
    user := flag.String("user", "", "SSH username")
    password := flag.String("password", "", "SSH password")
    host := flag.String("host", "", "SSH host")
    port := flag.Int("port", 22, "SSH port")
    flag.Parse()

    fmt.Printf("Hello, %s!\n", *name)
    if *verbose {
        fmt.Println("Verbose mode on")
    }

    if *uploadDir != "" {
        err := uploadFilesFromDir(*uploadDir, "http://localhost:8080/upload")
        if err != nil {
            log.Fatalf("failed to upload files: %v", err)
        }
        fmt.Println("Files uploaded successfully!")
        return
    }
    log.Println("Starting server...")
    if *remoteDir != "" && *user != "" && *password != "" && *host != "" {
        err := syncDirectories("./updatedDir", *remoteDir, *user, *password, *host, *port)
        if err != nil {
            log.Fatalf("failed to sync directories: %v", err)
        }
        fmt.Println("Directories synchronized successfully!")
        return
    }

    err := os.MkdirAll("./uploads", os.ModePerm)
    if err != nil {
        log.Fatalf("failed to create uploads directory: %v", err)
    }

    fs := http.FileServer(http.Dir("uploads"))
    http.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

    http.HandleFunc("/", indexHandler)
    http.HandleFunc("/upload", uploadHandler)

    log.Println("Server started on: http://localhost:8080")
    log.Println("Press Ctrl + C to stop the server")

    log.Printf("Serving files from %s on http://0.0.0.0:8080\n", "/uploads")
    log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}