package main

import (
	"database/sql"
)

type Node struct {
	Name   string `json:"name"`
	Host   string `json:"host"`
	Status string `json:"status"`
}

type Task struct {
	Sha256  string  `json:"sha256"`
	Md5     string  `json:"md5"`
	Status  string  `json:"status"`
	Task_id float64 `json:"task_id"`
	Host    string  `json:"host"`
}

type CuckooStruct struct {
	Tasks struct {
		Reported  int `json:"reported"`
		Running   int `json:"running"`
		Total     int `json:"total"`
		Completed int `json:"completed"`
		Pending   int `json:"pending"`
	} `json:"tasks"`
	Diskspace struct {
		Analyses struct {
			Total int64 `json:"total"`
			Free  int64 `json:"free"`
			Used  int64 `json:"used"`
		} `json:"analyses"`
		Binaries struct {
			Total int64 `json:"total"`
			Free  int64 `json:"free"`
			Used  int64 `json:"used"`
		} `json:"binaries"`
		Temporary struct {
			Total int64 `json:"total"`
			Free  int64 `json:"free"`
			Used  int64 `json:"used"`
		} `json:"temporary"`
	} `json:"diskspace"`
	Version         string `json:"version"`
	ProtocolVersion int    `json:"protocol_version"`
	Hostname        string `json:"hostname"`
	Machines        struct {
		Available int `json:"available"`
		Total     int `json:"total"`
	} `json:"machines"`
}

type CuckooTaskStruct struct {
	Task struct {
		AddedOn        string   `json:"added_on"`
		Category       string   `json:"category"`
		Clock          string   `json:"clock"`
		CompletedOn    string   `json:"completed_on"`
		Custom         string   `json:"custom"`
		Duration       int      `json:"duration"`
		EnforceTimeout bool     `json:"enforce_timeout"`
		Errors         []string `json:"errors"`
		Guest          struct {
			ID         int    `json:"id"`
			Label      string `json:"label"`
			Manager    string `json:"manager"`
			Name       string `json:"name"`
			ShutdownOn string `json:"shutdown_on"`
			StartedOn  string `json:"started_on"`
			Status     string `json:"status"`
			TaskID     int    `json:"task_id"`
		} `json:"guest"`
		ID      int    `json:"id"`
		Machine string `json:"machine"`
		Memory  bool   `json:"memory"`
		Options struct {
		} `json:"options"`
		Owner      string      `json:"owner"`
		Package    string      `json:"package"`
		Platform   string      `json:"platform"`
		Priority   int         `json:"priority"`
		Processing interface{} `json:"processing"`
		Route      string      `json:"route"`
		Sample     struct {
			Crc32    string      `json:"crc32"`
			FileSize int         `json:"file_size"`
			FileType string      `json:"file_type"`
			ID       int         `json:"id"`
			Md5      string      `json:"md5"`
			Sha1     string      `json:"sha1"`
			Sha256   string      `json:"sha256"`
			Sha512   string      `json:"sha512"`
			Ssdeep   interface{} `json:"ssdeep"`
		} `json:"sample"`
		SampleID  int           `json:"sample_id"`
		StartedOn string        `json:"started_on"`
		Status    string        `json:"status"`
		SubmitID  interface{}   `json:"submit_id"`
		Tags      []interface{} `json:"tags"`
		Target    string        `json:"target"`
		Timeout   int           `json:"timeout"`
	} `json:"task"`
}

func getNodes(db *sql.DB) ([]Node, error) {
	rows, err := db.Query("SELECT name, node FROM nodes")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := []Node{}

	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.Name, &n.Host); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (n *Node) getNode(db *sql.DB) error {
	return db.QueryRow("SELECT name, node FROM nodes WHERE name=$1", n.Name).Scan(&n.Name)
}

func (n *Node) createNode(db *sql.DB) error {
	err := db.QueryRow("INSERT INTO nodes(name, node) VALUES($1, $2) RETURNING id",
		n.Name, n.Host).Scan(&n.Name)
	if err != nil {
		return err
	}
	return nil
}

func (n *Node) deleteNode(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM nodes WHERE name=$1", n.Name)
	return err
}

func (n *Node) updateNode(db *sql.DB) error {
	_, err := db.Exec("UPDATE nodes SET status=$1 WHERE host=$2", n.Status, n.Host)
	return err
}

func getTasks(db *sql.DB) ([]Task, error) {
	rows, err := db.Query("SELECT sha256, md5, status, task_id, host FROM tasks WHERE status='pending'")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tasks := []Task{}

	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.Sha256, &t.Md5, &t.Status, &t.Task_id, &t.Host); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (t *Task) getTask(db *sql.DB) error {
	if len(t.Md5) > 0 {
		return db.QueryRow("SELECT sha256, md5, status, task_id, host FROM tasks WHERE md5=$1", t.Md5).Scan(&t.Md5)
	}
	return db.QueryRow("SELECT sha256, md5, status, task_id, host FROM tasks WHERE sha256=$1", t.Sha256).Scan(&t.Sha256)
}

func (t *Task) insertTask(db *sql.DB) error {
	err := db.QueryRow("INSERT INTO tasks(sha256, md5, status, task_id, host) VALUES($1,$2,$3,$4,$5)", t.Sha256, t.Md5, t.Status, t.Task_id, t.Host).Scan(
		&t.Sha256, &t.Md5, &t.Status, &t.Task_id, &t.Host)
	if err != nil {
		return err
	}
	return nil
}
