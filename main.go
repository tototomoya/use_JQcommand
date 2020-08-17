// apt -y install jq
package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/uniplaces/carbon"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))
var json_file = "diary.json"

type Diary struct {
	Date    string
	Content string
}

type Jsontext struct {
	User_name string
	Date      string
	Content   string
}

func write_json(data Jsontext) {
	jsondata, _ := json.Marshal(data)
	s := fmt.Sprintf("echo '%s' >> "+json_file, string(jsondata))
	_ = exec.Command("bash", "-c", s).Run()
}

func write_JsonPerUser(name string, date string) {
	cmd := "'. | select(.Date == " + strconv.Quote(date) + ") | select(.User_name == " + strconv.Quote(name) + ") | .Content'"
	cmd = "jq " + cmd + " " + json_file
	jq := exec.Command("bash", "-c", cmd)
	jq_out, _ := jq.Output()
	_ = ioutil.WriteFile(date+name, jq_out, 0777)
}

func createSession(w http.ResponseWriter, r *http.Request) {
	gob.Register([]Diary{})

	vars := mux.Vars(r)
	content := vars["content"]
	now, _ := carbon.NowInLocation("Asia/Tokyo")
	s_now := now.DateTimeString()

	var diarys []Diary
	session, _ := store.Get(r, "getDiary")
	if session.Values["diarys"] != nil {
		diarys = session.Values["diarys"].([]Diary)
		diarys = append(diarys, Diary{s_now, content})
	} else {
		diarys = append(diarys, Diary{s_now, content})
	}
	session.Values["diarys"] = diarys
	session.Values["name"] = vars["name"]
	session.Save(r, w)

	message := fmt.Sprintf("%vさんのcookieに%vを保存しました。", session.Values["name"], session.Values["diarys"])
	fmt.Fprint(w, message)

	go func() {
		write_json(Jsontext{vars["name"], s_now, vars["content"]})
		write_JsonPerUser(vars["name"], s_now)
	}()
}

func getSession(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "getDiary")
	message := fmt.Sprintf("%vさんのcookieのデータは%v ", session.Values["name"], session.Values["diarys"])
	fmt.Fprint(w, message)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/diary/{name}/{content}", createSession)
	r.HandleFunc("/list", getSession)
	http.ListenAndServe("0.0.0.0:8001", r)
}
