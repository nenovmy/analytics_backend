package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Server struct {
	db         *sql.DB
	maxConn    int
	dsn        string
	httpServer *http.Server
}

type Application struct {
	AppKey       string `json:"app_key"`
	Name         string `json:"name"`
	CreationTime int64  `json:"creation_time"`
}

type Event struct {
	ID        int    `json:"id"`
	ClientKey string `json:"client_key"`
	AppKey    string `json:"app_key"`
	Time      int64  `json:"time"`
	Platform  string `json:"platform"`
	IP        string `json:"ip"`
	Country   string `json:"country"`
	Version   string `json:"version"`
	Name      string `json:"name"`
	Data      string `json:"data"`
}

func NewServer(dsn string, maxConn int, addr string) *Server {
	return &Server{
		db:      nil,
		maxConn: maxConn,
		dsn:     dsn,
		httpServer: &http.Server{
			Addr: addr,
		},
	}
}

func (s *Server) StartServer() error {
	err := s.ConnectDB()
	if err != nil {
		return err
	}

	s.StartRouter()

	return nil
}

func (s *Server) ConnectDB() error {
	db, err := sql.Open("postgres", s.dsn)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(s.maxConn)

	if err = db.Ping(); err != nil {
		return err
	}

	s.db = db

	return nil
}

func (s *Server) StartRouter() {
	router := mux.NewRouter()

	router.HandleFunc("/applications", s.GetApplicationsHandler).Methods("GET")
	router.HandleFunc("/application", s.GetApplicationByAppKeyHandler).Methods("GET")
	router.HandleFunc("/events", s.GetEventsHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":12345", handlers.CORS(handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD"}), handlers.AllowedOrigins([]string{"*"}))(router)))
}

func (s *Server) GetApplicationByAppKey(apiKey string) (*Application, error) {
	rows, err := s.db.Query("SELECT app_key, name, creation_time FROM application WHERE app_key = $1", apiKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var apiKey string
		var name string
		var creationTime int64
		if err := rows.Scan(&apiKey, &name, &creationTime); err != nil {
			return nil, err
		}
		return &Application{AppKey: apiKey, Name: name, CreationTime: creationTime}, nil
	}

	return nil, err
}

func (s *Server) GetApplicationByAppKeyHandler(w http.ResponseWriter, r *http.Request) {
	appKey := r.URL.Query().Get("app_key")
	if appKey == "" {
		http.Error(w, "app_key parameter is required", http.StatusBadRequest)
		return
	}

	application, err := s.GetApplicationByAppKey(appKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(application)
}

func (s *Server) GetApplications() ([]Application, error) {
	rows, err := s.db.Query("SELECT app_key, name, creation_time FROM application")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	apps := []Application{}
	for rows.Next() {
		var apiKey string
		var name string
		var creationTime int64
		if err := rows.Scan(&apiKey, &name, &creationTime); err != nil {
			return nil, err
		}
		apps = append(apps, Application{AppKey: apiKey, Name: name, CreationTime: creationTime})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return apps, nil
}

func (s *Server) GetApplicationsHandler(w http.ResponseWriter, r *http.Request) {
	applications, err := s.GetApplications()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonBytes, err := json.Marshal(applications)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func (s *Server) GetEventsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse application ID from request
	appKey := r.URL.Query().Get("app_key") //vars["app_key"]

	// Parse time duration from request
	timeStr := r.URL.Query().Get("time")
	timeInt, err := strconv.Atoi(timeStr)
	if err != nil {
		http.Error(w, "Invalid time duration", http.StatusBadRequest)
		return
	}
	timeDuration := /*time.Duration(*/ timeInt //) * time.Minute

	println("time duration", timeDuration)

	// Get events for application and time duration
	events, err := s.GetEventsForApplication(appKey, int(timeDuration /*.Minutes()*/))
	if err != nil {
		http.Error(w, "Error retrieving events", http.StatusInternalServerError)
		return
	}

	// Convert events to JSON and return as response
	jsonResponse, err := json.Marshal(events)
	if err != nil {
		http.Error(w, "Error converting events to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (s *Server) GetEventsForApplication(apiKey string, minutesAgo int) ([]*Event, error) {
	// Calculate the cutoff time
	cutoffTime := time.Now().Unix() - int64(minutesAgo*60)

	// Query the database for the events for the given application and time range
	rows, err := s.db.Query("SELECT * FROM event WHERE app_key = $1 AND time >= $2 ORDER BY time DESC", apiKey, cutoffTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the rows and create Event structs
	events := []*Event{}
	for rows.Next() {
		var e Event
		err := rows.Scan(&e.ID, &e.ClientKey, &e.AppKey, &e.Time, &e.Platform, &e.IP, &e.Country, &e.Version, &e.Name, &e.Data)
		if err != nil {
			return nil, err
		}
		events = append(events, &e)
	}

	// Check for any errors during row iteration
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return events, nil
}
