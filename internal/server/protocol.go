package server

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	Command string
	Key     string
	Value   string
	TTL     time.Duration
}

type Response struct {
	Status string
	Value  string
}

func ParseRequest(input string) (Request, error) {
	parts := strings.Fields(input)

	if len(parts) == 0 {
		return Request{}, errors.New("empty command")
	}

	switch parts[0] {

	case "SET":
		if len(parts) == 3 {
			return Request{
				Command: "SET",
				Key:     parts[1],
				Value:   parts[2],
			}, nil
		}
		if len(parts) == 4 {
			ttlSecs, err := strconv.Atoi(parts[3])
			if err != nil || ttlSecs <= 0 {
				return Request{}, errors.New("invalid TTL value")
			}
			return Request{
				Command: "SET",
				Key:     parts[1],
				Value:   parts[2],
				TTL:     time.Duration(ttlSecs) * time.Second,
			}, nil
		}
		return Request{}, errors.New("invalid SET command")

	case "GET", "DELETE":
		if len(parts) != 2 {
			return Request{}, errors.New("invalid command")
		}

		return Request{
			Command: parts[0],
			Key:     parts[1],
		}, nil

	case "DUMP":
		if len(parts) != 1 {
			return Request{}, errors.New("invalid DUMP command")
		}

		return Request{
			Command: "DUMP",
		}, nil

	default:
		return Request{}, errors.New("unknown command")
	}
}
