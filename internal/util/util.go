package util

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

var Validdate = validator.New()

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}
	return json.NewDecoder(r.Body).Decode(payload)
}

func WriteJSON(w http.ResponseWriter, status int, value any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(value)
}

func WriteError(w http.ResponseWriter, status int, err error) {
	errResp := WriteJSON(w, status, map[string]string{"error": err.Error()})
	if errResp != nil {
		log.Fatal(errResp)
	}
}
func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func ParseFlightIDIntoBinary(flightID string) ([]byte, int, error) {
	fligtId, err := uuid.Parse(flightID)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("failed to format FlightID to uuid %w", err)
	}
	binaryFlightID, err := fligtId.MarshalBinary()
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to format FlightID to uuid binary %w", err)
	}
	return binaryFlightID, 0, nil
}
