package main

import (
	"fmt"
	"time"
)

// RequestMetadata represents request metadata...
type RequestMetadata struct {
	JobType    string `json:"jobType"`
	WorkerName string `json:"workerName"`
}

// ResponseMetadata represents response metadata...
type ResponseMetadata struct {
	Success bool `json:"success"`
	Code    int  `json:"code"`
}

// Request represents a request...
type Request struct {
	Metadata RequestMetadata `json:"metadata"`
}

// Response represents a response...
type Response struct {
	Metadata ResponseMetadata `json:"metadata"`
}

// Body represents the body...
type Body struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

// LogEntry represents a log entry...
type LogEntry struct {
	Index string `json:"index"`
	Type  string `json:"type"`
	Body  Body   `json:"body"`
}

// NewLogEntry constructs a new LogEntry
func NewLogEntry(JobType, WorkerName string, Code int) *LogEntry {
	now := time.Now()
	index := fmt.Sprintf("metric-tattle-%04d-%02d-%02d", now.Year(), now.Month(), now.Day())

	requestMetadata := RequestMetadata{JobType, WorkerName}
	request := Request{Metadata: requestMetadata}

	responseMetadata := ResponseMetadata{Code: Code, Success: false}
	response := Response{Metadata: responseMetadata}

	body := Body{Request: request, Response: response}
	return &LogEntry{Index: index, Type: "tattle", Body: body}
}
