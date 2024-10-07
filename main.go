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

func main() {
    name := flag.String("name", "world", "a name to say hello to")
    verbose := flag.Bool("verbose", false, "enable verbose mode")
    uploadDir := flag.String("uploadDir", "", "directory to upload files from")
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