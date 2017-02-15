package main

import (
  "fmt"
  "net/http"
  "os"
  "plugin"
)

const (
  CODE_PATH = "/userfunc/user"
)

var userFunc http.HandlerFunc

func loadPlugin() http.HandlerFunc {
  p, err := plugin.Open(CODE_PATH)
  if err != nil {
    panic(err)
  }
  sym, err := p.Lookup("HandlerFunc")
  if err != nil {
    panic("Entry point missing: func HandlerFunc(w http.ResponseWriter, r *http.Request)")
  }
  return sym.(func(http.ResponseWriter, *http.Request))
}

func specializeHandler(w http.ResponseWriter, r *http.Request) {
  if userFunc != nil {
    w.WriteHeader(400)
    w.Write([]byte("Not a generic container"))
    return
  }

  _, err := os.Stat(CODE_PATH)
  if err != nil {
    if os.IsNotExist(err) {
      w.WriteHeader(http.StatusNotFound)
      w.Write([]byte(CODE_PATH + ": not found"))
      return
    } else {
      panic(err)
    }
  }

  fmt.Printf("Specializing... ")
  userFunc = loadPlugin()
  fmt.Println("Done")
}

func main() {
  http.HandleFunc("/specialize", specializeHandler)

  // Generic route -- all http requests go to the user function.
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    if userFunc == nil {
      w.WriteHeader(http.StatusInternalServerError)
      w.Write([]byte("Generic container: no requests supported"))
      return
    }
    userFunc(w, r)
  })

  fmt.Println("Listening on :8888 ...")
  http.ListenAndServe(":8888", nil)
}
