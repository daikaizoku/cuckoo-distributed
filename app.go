package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func (a *App) Initialize(user, password, dbname string) {
	connectionString :=
		fmt.Sprintf("user=%s password=%s dbname=%s", user, password, dbname)
	var err error
	a.DB, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(":8000", a.Router))
}

func (a *App) getNode(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	node_name := vars["name"]
	n := Node{Name: node_name}
	if err := n.getNode(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Node alias not found.")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
	}
	respondWithJSON(w, http.StatusOK, n)
}

func (a *App) getNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := getNodes(a.DB)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, nodes)
}

func (a *App) createNode(w http.ResponseWriter, r *http.Request) {
	var n Node
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&n); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()
	if err := n.createNode(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, n)
}

func (a *App) deleteNode(w http.ResponseWriter, r *http.Request) {
	var n Node
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&n); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	defer r.Body.Close()
	if err := n.deleteNode(a.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) getTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := getTasks(a.DB)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, tasks)
}

func (a *App) createTask(w http.ResponseWriter, r *http.Request) {
	file, handler, err := r.FormFile("filename")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
	}

	defer file.Close()

	f, err := os.OpenFile("/tmp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()
	if _, err := io.Copy(f, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	err = a.submitTask("/tmp/"+handler.Filename, w)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
	}
	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

func (a *App) getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	task_sha256 := vars["sha256"]
	/*	if val, ok := vars["md5"]; ok {

		}
	*/
	t := Task{sha256: task_sha256}
	if err := t.getTask(a.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "Task sha256 not found.")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
	}
	respondWithJSON(w, http.StatusOK, t)
}

// TODO: change this whole function done with curl for speedness
func (a *App) submitTask(filename string, w http.ResponseWriter) error {
	nodes, err := getNodes(a.DB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return err
	}
	for _, node := range nodes {
		if node.Status == "active" {
			cmd := fmt.Sprintf("curl -F file=@%s http://%s:8090/tasks/create/file", filename, node.Host)
			parts := strings.Fields(cmd)
			out, err := exec.Command(parts[0], parts[1], parts[2], parts[3]).Output()
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, err.Error())
				return err
			}
			// Little hack i did, this shouldn't really be like this
			var parsed_output map[string]interface{}
			json.Unmarshal(out, &parsed_output)
			respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
		}
	}
	respondWithJSON(w, http.StatusServiceUnavailable, map[string]string{"result": "All machines are busy. Queued."})
	return nil
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/node/list", a.getNodes).Methods("GET")
	a.Router.HandleFunc("/node/{id:[0-9]+}", a.getNode).Methods("GET")
	a.Router.HandleFunc("/node/create", a.createNode).Methods("POST")
	a.Router.HandleFunc("/node/delete", a.deleteNode).Methods("POST")
	a.Router.HandleFunc("/task/list", a.getTasks).Methods("GET")
	a.Router.HandleFunc("/task/", a.getTask).Methods("POST")
	a.Router.HandleFunc("/task/create", a.createTask).Methods("POST")
}
