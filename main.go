package blobstore_example

import (
        "html/template"
        "io"
        "net/http"
        "time"

        "golang.org/x/net/context"

        "google.golang.org/appengine"
        "google.golang.org/appengine/blobstore"
        "google.golang.org/appengine/log"

        "google.golang.org/appengine/datastore"
        "google.golang.org/appengine/user"
)

type Entry struct {
        Author  string
        Content string
        ImgKey  string
        Date    time.Time
}

func entryKey(ctx context.Context) *datastore.Key {
        return datastore.NewKey(ctx, "Entry", "default_entry", 0, nil)
}

func serveError(ctx context.Context, w http.ResponseWriter, err error) {
        w.WriteHeader(http.StatusInternalServerError)
        w.Header().Set("Content-Type", "text/plain")
        io.WriteString(w, "Internal Server Error")
        log.Errorf(ctx, "%v", err)
}


var rootTemplate = template.Must(template.New("root").Parse(rootTemplateHTML))

const rootTemplateHTML = `
<html><body>
<form action="{{.}}" method="POST" enctype="multipart/form-data">
<input type="text" name="author" placeholder="author"><br>
<input type="text" name="content" placeholder="content"><br>
Upload File: <input type="file" name="file"><br>
<input type="submit" name="submit" value="Submit">
</form></body></html>
`

func handleRoot(w http.ResponseWriter, r *http.Request) {
        // entrys := make([]Entry, 0, 10)

        ctx := appengine.NewContext(r)
        uploadURL, err := blobstore.UploadURL(ctx, "/upload", nil)
        if err != nil {
                serveError(ctx, w, err)
                return
        }
        w.Header().Set("Content-Type", "text/html")
        err = rootTemplate.Execute(w, uploadURL)
        if err != nil {
                log.Errorf(ctx, "%v", err)
        }
}

func handleView(w http.ResponseWriter, r *http.Request) {
        ctx := appengine.NewContext(r)
        q := datastore.NewQuery("Entry").Ancestor(entryKey(ctx)).Order("-Date").Limit(100)
        entrys := make([]Entry, 0, 10)
        if _, err := q.GetAll(ctx, &entrys); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        if err := entryTemplate.Execute(w, entrys); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
        }
}

var entryTemplate = template.Must(template.New("book").Parse(`
<html>
  <head>
    <title></title>
    <link rel="stylesheet" href="https://storage.googleapis.com/code.getmdl.io/1.0.6/material.indigo-pink.min.css">
    <script src="https://storage.googleapis.com/code.getmdl.io/1.0.6/material.min.js"></script>
    <link rel="stylesheet" href="https://fonts.googleapis.com/icon?family=Material+Icons">
  </head>
  <body>
    {{range .}}
      {{with .Author}}
        <p>{{.}}:
        {{.Date}}</p>
      {{else}}
        <p>名無しさん:
        {{.Date.Format "2006-01-02 15:04:05"}}</p>
      {{end}}
      <pre>{{.Content}}</pre>
      <pre><img src="/serve/?blobKey={{.ImgKey}}"></pre>
    {{end}}
  </body>
</html>
`))


func handleServe(w http.ResponseWriter, r *http.Request) {
        blobstore.Send(w, appengine.BlobKey(r.FormValue("blobKey")))
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
        ctx := appengine.NewContext(r)
        blobs, _, err := blobstore.ParseUpload(r)
        if err != nil {
                serveError(ctx, w, err)
                return
        }
        file := blobs["file"]
        if len(file) == 0 {
                log.Errorf(ctx, "no file uploaded")
                http.Redirect(w, r, "/", http.StatusFound)
                return
        }

        g := Entry{
                Author:  r.FormValue("author"),
                Content: r.FormValue("content"),
                // Author:  string("aut"),
                // Content: string("con"),
                ImgKey:  string(file[0].BlobKey),
                Date:    time.Now(),
        }
        if u := user.Current(ctx); u != nil {
                g.Author = u.String()
        }
        key := datastore.NewIncompleteKey(ctx, "Entry", entryKey(ctx))
        _, err2 := datastore.Put(ctx, key, &g)
        if err2 != nil {
                http.Error(w, err2.Error(), http.StatusInternalServerError)
                return
        }
        http.Redirect(w, r, "/", http.StatusFound)
}

func init() {
        http.HandleFunc("/", handleRoot)
        http.HandleFunc("/serve/", handleServe)
        http.HandleFunc("/upload", handleUpload)
        http.HandleFunc("/view", handleView)
}
