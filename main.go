package main

import (
    "encoding/gob"
    "io/ioutil"
    "os/exec"
    "fmt"
    "net/http"
    "time"
    "encoding/json"
    "github.com/gorilla/mux"
    "github.com/gorilla/securecookie"
    "github.com/gorilla/sessions"
    "strconv"
)

var store = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))

type Diary struct {
    Date string
    Content string
}

type Jsontext struct {
    User_name string
    Date string
    Content string
}


func write_json(data Jsontext) {
    jsondata, _ := json.Marshal(data)
    s := fmt.Sprintf("echo '%s' >> test.json", string(jsondata))
    _ = exec.Command("bash", "-c", s).Run()
}

func write_DateContent(date string) {
    s := "'. | select(.Date == " + strconv.Quote(date) + ") | select(.User_name == " + strconv.Quote("tomoya") + ") | .Content'"
    s = "jq " + s + " test.json"
    jq := exec.Command("bash", "-c", s)
    out, _ := jq.Output()
    _ = ioutil.WriteFile(date + "_tomoya", out, 0777)
}

func createSession(w http.ResponseWriter, r *http.Request) {
    gob.Register([]Diary{})

    vars := mux.Vars(r)

    now := time.Now()
    nowUTC := now.UTC()
    jst := time.FixedZone("Japan", 9*60*60)
    nowJst := nowUTC.In(jst)
    s := fmt.Sprintf("%d年%d月%d日", nowJst.Year(), int(nowJst.Month()), nowJst.Day())

    var diarys []Diary
    session, _ := store.Get(r, "getDiary")
    if session.Values["diary"] != nil {
	diarys = session.Values["diary"].([]Diary)
        diarys = append(diarys, Diary{s, vars["content"]})
    } else {
        diarys = append(diarys, Diary{s, vars["content"]})
    }
    session.Values["diarys"] = diarys
    session.Save(r, w)

    ss := fmt.Sprintf("cookieに%vを保存しました。", session.Values["diarys"])
    fmt.Fprint(w, ss)

    write_json(Jsontext{ vars["name"], s, vars["content"] } )
    write_DateContent(s)
}

func getSession(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "getDiarys")
    fmt.Fprint(w, session.Values["diarys"])
}

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/diary/{name}/{content}", createSession)
    r.HandleFunc("/index/{name}", getSession)
    http.ListenAndServe("0.0.0.0:8000", r)
}
