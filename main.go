package main

import (
    "fmt"
    "time"
    "net/http"

    "appengine"
    "appengine/datastore"
)


type Employee struct {
    Name     string
    Role     string
    HireDate time.Time
}


func handler(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)

    e1 := Employee{
        Name:     "Joe Citizen",
        Role:     "Manager",
        HireDate: time.Now(),
    }

    key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "employee", nil), &e1)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var e2 Employee
    if err = datastore.Get(c, key, &e2); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Stored and retrieved the Employee named %q", e2.Name)
}

func init() {
    http.HandleFunc("/", handler)
}
