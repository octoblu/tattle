package logentry

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
	ElapsedTime int      `json:"elapsedTime"`
	Request     Request  `json:"request"`
	Response    Response `json:"response"`
}

// LogEntry represents a log entry...
type LogEntry struct {
	Index string `json:"index"`
	Type  string `json:"type"`
	Body  Body   `json:"body"`
}

// New constructs a new LogEntry
func New(indexPrefix, typeName, JobType, WorkerName string, Code, ElapsedTime int) *LogEntry {
	now := time.Now()
	index := fmt.Sprintf("%v-%04d-%02d-%02d", indexPrefix, now.Year(), now.Month(), now.Day())

	requestMetadata := RequestMetadata{JobType, WorkerName}
	request := Request{Metadata: requestMetadata}

	Success := (Code < 500)
	responseMetadata := ResponseMetadata{Code: Code, Success: Success}
	response := Response{Metadata: responseMetadata}

	body := Body{Request: request, Response: response, ElapsedTime: ElapsedTime}
	return &LogEntry{Index: index, Type: typeName, Body: body}
}
