package app

import (
	"net/http"
)

var TESTS = []handler{
	handler{
		Action: "Create",
		Method: "POST",
		URI:    "/tests",
		// Auth:   <true, false>,
		// Params: []string{...},
		Do: func(w http.ResponseWriter, r *http.Request) {
		
		}},
	handler{
		Action: "Show",
		Method: "GET",
		URI:    "/tests/:id",
		// Auth:   <true, false>,
		// Params: []string{...},
		Do: func(w http.ResponseWriter, r *http.Request) {
		
		}},
	handler{
		Action: "Update",
		Method: "PUT",
		URI:    "/tests/:id",
		// Auth:   <true, false>,
		// Params: []string{...},
		Do: func(w http.ResponseWriter, r *http.Request) {
		
		}},
	handler{
		Action: "Delete",
		Method: "DELETE",
		URI:    "/tests/:id",
		// Auth:   <true, false>,
		// Params: []string{...},
		Do: func(w http.ResponseWriter, r *http.Request) {
		
		}}}
		