// Package server provides functionality to store and retrieve votes.
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/copilot-example-voting-app/api/vote"

	"github.com/gorilla/mux"
)

// Server is an API server.
type Server struct {
	Router *mux.Router
	DB     vote.DB
}

// ServeHTTP delegates to the mux router.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.HandleFunc("/_healthcheck", s.handleHealthCheck())
	s.Router.HandleFunc("/votes", s.handleStoreVote()).Methods(http.MethodPost)
	s.Router.HandleFunc("/votes/{voterID}", s.handleGetVote()).Methods(http.MethodGet)
	s.Router.HandleFunc("/results", s.handleGetResults()).Methods(http.MethodGet)

	s.Router.ServeHTTP(w, r)
}

func (s *Server) handleHealthCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) handleStoreVote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			VoterID string `json:"voter_id"`
			Vote    string `json:"vote"`
		}{}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&data); err != nil {
			log.Printf("ERROR: server: decode payload: %v\n", err)
			http.Error(w, "decode JSON payload", http.StatusBadRequest)
			return
		}
		if err := s.DB.Store(data.VoterID, data.Vote); err != nil {
			log.Printf("ERROR: server: store vote %+v: %v\n", data, err)
			http.Error(w, fmt.Sprintf("store vote for voter ID %s", data.VoterID), http.StatusInternalServerError)
			return
		}
		log.Printf("INFO: server: registered vote for voter ID %s\n", data.VoterID)
		w.WriteHeader(http.StatusOK)
	}
}

func (s *Server) handleGetVote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		result, err := s.DB.Result(vars["voterID"])
		if err != nil {
			if errors.Is(err, vote.ErrNoVote{VoterID: vars["voterID"]}) {
				log.Printf("WARN: server: vote for voterID does not exist: %s\n", vars["voterID"])
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			log.Printf("ERROR: server: get vote for voterID %s: %v\n", vars["voterID"], err)
			http.Error(w, fmt.Sprintf("get vote for voter ID %s", vars["voterID"]), http.StatusInternalServerError)
			return
		}

		dat, err := json.Marshal(&struct {
			Result string `json:"vote"`
		}{
			Result: result,
		})
		if err != nil {
			log.Printf("ERROR: server: encode get vote payload: %v", err)
			http.Error(w, "encode JSON payload", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
	}
}

func (s *Server) handleGetResults() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results, err := s.DB.Results()
		if err != nil {
			log.Printf("ERROR: server: get all vote results: %v\n", err)
			http.Error(w, "get results", http.StatusInternalServerError)
			return
		}

		dat, err := json.Marshal(&struct {
			Results []vote.ResultCount `json:"results"`
		}{
			Results: results,
		})
		if err != nil {
			log.Printf("ERROR: encode get results payload: %v", err)
			http.Error(w, "encode JSON payload", http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(dat)
	}
}
