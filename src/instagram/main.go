package main

import(
	"net/http"
	"encoding/json"
	"sync"
	"io/ioutil"
	"time"
	"fmt"
	"strings"
)

type user struct {
	Name string `json:"name"`
	ID string `json: "id"`
	Email string `json: "email"`
	Password string `json: "password"`
}

type userHandler struct {
	sync.Mutex
	store map[string]user
}

func (h *userHandler) users(w http.ResponseWriter, r *http.Request){
	switch r.Method{
		case "GET":			//take action according to get or post
			h.get(w, r)
			return
		case "POST":
			h.post(w, r)
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)	//if not a get or post request
			return
	}
}

func (h *userHandler) post(w http.ResponseWriter, r *http.Request){
	body, err := ioutil.ReadAll(r.Body)		//contains user sent json
	defer r.Body.Close()
	
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	
	var lst user
	err = json.Unmarshal(body, &lst)				//handle bad data by user
	
	if err!=nil {
		w.WriteHeader(http.StatusBadRequest)			//400
		w.Write([]byte(err.Error()))
		return
	}
	
	lst.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	
	h.Lock()				//deal with concurrency
	h.store[lst.ID] = lst
	defer h.Unlock()
}

func (h *userHandler) get(w http.ResponseWriter, r *http.Request){
	lst := make([]user, len(h.store))		//creating list
	
	h.Lock()					//deal with concurrency
	i := 0
	for _, temp := range h.store {		//add components
		lst[i] = temp
		i++
	}
	h.Unlock()
	
	jsonBytes, err := json.Marshal(lst)		//turn list to json
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)		//handle error if bad data
		w.Write([]byte(err.Error()))
		return
	}
	
	w.WriteHeader(http.StatusOK)			//code 200
	w.Write(jsonBytes)				//send the data
}

func (h *userHandler) getUser(w http.ResponseWriter, r *http.Request){
	parts := strings.Split(r.URL.String(), "/")
	if len(parts)!=3 {
		w.WriteHeader(http.StatusNotFound)			//can't have 3 parts
	}
	
	h.Lock()					//deal with concurrency
	lst, _ := h.store[parts[2]]
	h.Unlock()
	
	jsonBytes, err := json.Marshal(lst)		//turn list to json
	if err!=nil {
		w.WriteHeader(http.StatusInternalServerError)		//handle error if bad data
		w.Write([]byte(err.Error()))
		return
	}
	
	w.WriteHeader(http.StatusOK)			//code 200
	w.Write(jsonBytes)				//send the data
}

func newUserHandler() *userHandler {
	return &userHandler{
		store: map[string]user{},
	}
}

func main() {
	userHandler := newUserHandler()
	http.HandleFunc("/users", userHandler.users)	//handling get and post with users handler
	http.HandleFunc("/users/", userHandler.getUser)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err) 
	}
}
