package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

var router *mux.Router

type HttpFunc func(w http.ResponseWriter, r *http.Request) bool

type RouteFilter struct {
	filters []HttpFunc
	hdlr    http.HandlerFunc
}

func NewRouteFilter() *RouteFilter {
	return &RouteFilter{}
}

func (f *RouteFilter) AddFilter(h HttpFunc) *RouteFilter {
	f.filters = append(f.filters, h)
	return f
}

func (f *RouteFilter) Handler(h http.HandlerFunc) *RouteFilter {
	f.hdlr = h
	return f
}

func (f *RouteFilter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, hdlr := range f.filters {
		if !hdlr(w, req) {
			return
		}
	}
	f.hdlr(w, req)
}
