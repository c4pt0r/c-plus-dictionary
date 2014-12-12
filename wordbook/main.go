package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"strings"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	KeySeperator    = "_"
	KeyPrefixUser   = "user"
	KeyPrefixRecord = "record"
	KeyPrefixToken  = "token"
)

type User struct {
	Username string `json:"username"`
	Pwd      string `json:"password"`
	Token    string `json:"token"`
}

type Record struct {
	Word     string    `json:"word"`
	Phonetic string    `json:"phonetic"`
	Explain  string    `json:"explain"`
	Username string    `json:"username"`
	CreateAt time.Time `json:"create_at"`
}

var dbPath = flag.String("db", "./.wordbook.db", "db path")
var addr = flag.String("addr", ":8088", "addr")
var db *leveldb.DB

func buildKey(args ...string) []byte {
	return []byte(strings.Join(args, KeySeperator))
}
func md5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

var ErrUserExists = errors.New("username already exists")

func UserExists(username string) (bool, error) {
	k := buildKey(KeyPrefixUser, username)
	b, err := db.Has(k, nil)
	if err != nil {
		return false, err
	}
	return b, nil
}

func GetUserFromName(username string) (*User, error) {
	k := buildKey(KeyPrefixUser, username)

	b, err := db.Get(k, nil)
	if err != nil {
		return nil, err
	}

	var u User
	json.Unmarshal(b, &u)
	return &u, nil
}

func CheckUserToken(username string, token string) (bool, error) {
	u, err := GetUserFromName(username)
	if err != nil {
		return false, err
	}
	if u == nil {
		return false, nil
	}
	if u.Token != token {
		return false, nil
	}
	return true, nil
}

func CheckUser(username string, pwd string) (bool, error) {
	u, err := GetUserFromName(username)
	if err != nil {
		return false, err
	}
	if u.Username == username && u.Pwd == md5Hash(pwd) {
		return true, nil
	}
	return false, nil
}

func GetUserFromToken(token string) (*User, error) {
	k := buildKey(KeyPrefixToken, token)
	uname, err := db.Get(k, nil)
	if err != nil {
		return nil, err
	}
	return GetUserFromName(string(uname))
}

func Register(username string, pwd string) (*User, error) {
	exists, err := UserExists(username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}
	k := buildKey(KeyPrefixUser, username)

	u4, _ := uuid.NewV4()

	u := &User{
		Username: username,
		Pwd:      md5Hash(pwd),
		Token:    md5Hash(u4.String()),
	}

	b, _ := json.Marshal(u)
	err = db.Put(k, b, nil)
	if err != nil {
		return nil, err
	}

	// update token list
	k = buildKey(KeyPrefixToken, u.Token)
	err = db.Put(k, []byte(u.Username), nil)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func CreateRecord(username string, rec *Record) error {
	rec.CreateAt = time.Now()
	date := rec.CreateAt.Format("20060102")
	k := buildKey(KeyPrefixRecord, username, date, rec.Word)
	b, _ := json.Marshal(rec)
	err := db.Put(k, b, nil)
	if err != nil {
		return err
	}
	return nil
}

func iterRecords(keyPrefix []byte) ([]*Record, error) {
	records := make([]*Record, 0)
	iter := db.NewIterator(util.BytesPrefix(keyPrefix), nil)
	for iter.Next() {
		var rec Record
		b := iter.Value()
		err := json.Unmarshal(b, &rec)
		if err != nil {
			return nil, err
		}
		records = append(records, &rec)
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		return nil, err
	}
	return records, nil
}

func GetRecordsByDate(username string, dateStr string) ([]*Record, error) {
	k := buildKey(KeyPrefixRecord, username, dateStr)
	return iterRecords(k)
}

func GetRecords(username string) ([]*Record, error) {
	k := buildKey(KeyPrefixRecord, username)
	return iterRecords(k)
}

/* handlers */

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	form, err := jason.NewObjectFromReader(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	username, err := form.GetString("username")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	password, err := form.GetString("password")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	u, err := Register(username, password)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	b, _ := json.Marshal(map[string]interface{}{
		"token": u.Token,
	})
	w.Write(b)
}

func GetTokenFromHeader(r *http.Request) string {
	_, ok := r.Header["Authorization"]
	if !ok {
		return ""
	}
	auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)
	if len(auth) != 2 {
		return ""
	}
	method, payload := auth[0], auth[1]
	if method != "Token" {
		return ""
	}
	return payload
}

func AuthFilter(w http.ResponseWriter, r *http.Request) bool {
	_, ok := r.Header["Authorization"]
	if !ok {
		http.Error(w, "Need Authorization Header", http.StatusBadRequest)
		return false
	}
	auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)

	if len(auth) != 2 {
		http.Error(w, "Bad Syntax", http.StatusBadRequest)
		return false
	}

	method, payload := auth[0], auth[1]

	switch method {
	case "Basic":
		p, _ := base64.StdEncoding.DecodeString(payload)
		pair := strings.SplitN(string(p), ":", 2)
		if len(pair) != 2 {
			break
		}
		b, err := CheckUser(pair[0], pair[1])
		if err != nil {
			break
		}
		return b
	case "Token":
		u, err := GetUserFromToken(payload)
		if err != nil || u == nil {
			break
		}
		return true
	}
	http.Error(w, "Authorization Failed", http.StatusUnauthorized)
	return false
}

func CreateRecordHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	dec := json.NewDecoder(r.Body)
	var rec Record
	err := dec.Decode(&rec)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	rec.Username = username
	err = CreateRecord(username, &rec)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	b, _ := json.Marshal(rec)
	w.Write(b)
}

func GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	if ok, _ := CheckUserToken(username, GetTokenFromHeader(r)); !ok {
		http.Error(w, "Authorization Failed", http.StatusUnauthorized)
		return
	}

	k := buildKey(KeyPrefixUser, username)
	b, err := db.Get(k, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var u User
	json.Unmarshal(b, &u)

	b, _ = json.Marshal(map[string]interface{}{
		"token": u.Token,
	})
	w.Write(b)
}

func GetRecordsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	if ok, _ := CheckUserToken(username, GetTokenFromHeader(r)); !ok {
		http.Error(w, "Authorization Failed", http.StatusUnauthorized)
		return
	}

	recs, err := GetRecords(username)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	b, _ := json.Marshal(recs)
	w.Write(b)
}

func GetRecordsByDateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]
	dateStr := vars["dateStr"]
	if ok, _ := CheckUserToken(username, GetTokenFromHeader(r)); !ok {
		http.Error(w, "Authorization Failed", http.StatusUnauthorized)
		return
	}

	recs, err := GetRecordsByDate(username, dateStr)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	b, _ := json.Marshal(recs)
	w.Write(b)
}

func main() {
	flag.Parse()

	var err error
	db, err = leveldb.OpenFile(*dbPath, nil)
	defer db.Close()
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/register", RegisterHandler).Methods("POST")
	r.Handle("/{username}/token", NewRouteFilter().AddFilter(AuthFilter).Handler(GetTokenHandler)).Methods("GET")
	r.Handle("/{username}", NewRouteFilter().AddFilter(AuthFilter).Handler(CreateRecordHandler)).Methods("POST")
	r.Handle("/{username}", NewRouteFilter().AddFilter(AuthFilter).Handler(GetRecordsHandler)).Methods("GET")
	r.Handle("/{username}/{dateStr}", NewRouteFilter().AddFilter(AuthFilter).Handler(GetRecordsByDateHandler)).Methods("GET")

	http.ListenAndServe(*addr, r)
}
