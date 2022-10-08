package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/crossedbot/common/golang/logger"
	"github.com/crossedbot/common/golang/server"
	middleware "github.com/crossedbot/simplemiddleware"

	"github.com/crossedbot/axis/pkg/pins/models"
)

const (
	MaxPinLimit  = 1000
	MaxNameLimit = 255
)

func FindPins(w http.ResponseWriter, r *http.Request, p server.Parameters) {
	var err error
	uid, ok := r.Context().Value(middleware.ClaimUserId).(string)
	if !ok {
		logger.Info("no user id")
	}
	query := r.URL.Query()
	// get cid parameter(s)
	cids, present := query["cid"]
	if present {
		if len(cids) > MaxPinLimit {
			server.JsonResponse(w, models.NewFailure(
				models.ErrMaxPinLimitCode,
				fmt.Sprintf(
					"max cid limit exceeded [1 .. %d]",
					MaxPinLimit,
				),
			), http.StatusBadRequest)
			return
		}
	}
	// partial or fullmatch name of pins
	name := query.Get("name")
	if len(name) > MaxNameLimit {
		server.JsonResponse(w, models.NewFailure(
			models.ErrMaxPinLimitCode,
			fmt.Sprintf(
				"max name limit exceeded %d characters",
				MaxNameLimit,
			),
		), http.StatusBadRequest)
		return
	}
	// get text matching strategy
	matchStr := query.Get("match")
	match := models.TextMatchExact
	if matchStr != "" {
		match, err = models.ToTextMatch(matchStr)
		if err != nil {
			// unknown text matching
			server.JsonResponse(w, models.NewFailure(
				models.ErrUnknownTextMatchStringCode,
				fmt.Sprintf("%s \"%s\"", err.Error(), matchStr),
			), http.StatusBadRequest)
			return
		}
	}
	// get status parameter(s)
	statuses := []models.Status{}
	statusStrings, present := query["status"]
	if present {
		// create status map to remove duplicates, and append them to the status
		// list
		statusMap := make(map[models.Status]struct{})
		for _, s := range statusStrings {
			status, err := models.ToStatus(s)
			if err != nil {
				// unknown status
				server.JsonResponse(w, models.NewFailure(
					models.ErrUnknownStatusStringCode,
					fmt.Sprintf("%s \"%s\"", err.Error(), s),
				), http.StatusBadRequest)
				return
			}
			statusMap[status] = struct{}{}
		}
		for s, _ := range statusMap {
			statuses = append(statuses, s)
		}
	}
	// get time range parameters
	before := query.Get("before")
	after := query.Get("after")
	// get limit parameter
	limit := 10
	if v := query.Get("limit"); v != "" {
		var err error
		limit, err = strconv.Atoi(v)
		if err != nil {
			server.JsonResponse(w, models.NewFailure(
				models.ErrFailedConversionCode,
				"limit is not an integer",
			), http.StatusBadRequest)
			return
		}
		if limit > MaxPinLimit {
			server.JsonResponse(w, models.NewFailure(
				models.ErrMaxPinLimitCode,
				fmt.Sprintf(
					"max limit exceeded [1 .. %d]",
					MaxPinLimit,
				),
			), http.StatusBadRequest)
			return
		}
	}
	// get offset data parameter
	offset := 0
	if v := query.Get("offset"); v != "" {
		var err error
		offset, err = strconv.Atoi(v)
		if err != nil {
			server.JsonResponse(w, models.NewFailure(
				models.ErrFailedConversionCode,
				"offset is not an integer",
			), http.StatusBadRequest)
			return
		}
		if offset > MaxPinLimit {
			server.JsonResponse(w, models.NewFailure(
				models.ErrMaxPinLimitCode,
				fmt.Sprintf(
					"offset exceeds max limit [1 .. %d]",
					MaxPinLimit,
				),
			), http.StatusBadRequest)
			return
		}
	}
	// get sort data parameter
	sortBy := "created"
	if v := query.Get("sort"); v != "" {
		switch v {
		case "cid":
		case "created":
		case "name":
		case "status":
		default:
			// unknown sorting string
			server.JsonResponse(w, models.NewFailure(
				models.ErrUnknownSortStringCode,
				fmt.Sprintf("Unknown sorting string: \"%s\"", v),
			), http.StatusBadRequest)
			return
		}
		sortBy = v
	}
	// get meta data parameter
	meta := make(map[string]string)
	if metaString := query.Get("meta"); metaString != "" {
		if err := json.Unmarshal([]byte(metaString), &meta); err != nil {
			server.JsonResponse(w, models.NewFailure(
				models.ErrFailedConversionCode,
				fmt.Sprintf("failed to parse meta data; %s", err),
			), http.StatusBadRequest)
			return
		}
	}
	pins, err := Ctrl().FindPins(
		uid, cids, name,
		before, after, match,
		statuses, limit, offset,
		sortBy, meta,
	)
	if err != nil {
		server.JsonResponse(w, models.NewFailure(
			models.ErrProcessingRequestCode,
			fmt.Sprintf("failed to find pins; %s", err),
		), http.StatusInternalServerError)
		return
	}
	server.JsonResponse(w, &pins, http.StatusOK)
}

func GetPin(w http.ResponseWriter, r *http.Request, p server.Parameters) {
	uid, ok := r.Context().Value(middleware.ClaimUserId).(string)
	if !ok {
		logger.Info("no user id")
	}
	id := p.Get("id")
	if id == "" {
		server.JsonResponse(w, models.NewFailure(
			models.ErrRequiredParamCode,
			"path parameter 'id' is required",
		), http.StatusBadRequest)
		return
	}
	pinStatus, err := Ctrl().GetPin(uid, id)
	if err == ErrorPinNotFound {
		server.JsonResponse(w, models.NewFailure(
			models.ErrProcessingRequestCode,
			fmt.Sprintf("failed to get pin; %s", err),
		), http.StatusNotFound)
		return
	} else if err != nil {
		server.JsonResponse(w, models.NewFailure(
			models.ErrProcessingRequestCode,
			fmt.Sprintf("failed to get pin; %s", err),
		), http.StatusInternalServerError)
		return
	}
	server.JsonResponse(w, &pinStatus, http.StatusOK)
}

func CreatePin(w http.ResponseWriter, r *http.Request, p server.Parameters) {
	uid, ok := r.Context().Value(middleware.ClaimUserId).(string)
	if !ok {
		logger.Info("no user id")
	}
	var pin models.Pin
	if err := json.NewDecoder(r.Body).Decode(&pin); err != nil {
		logger.Error(err)
		server.JsonResponse(w, models.NewFailure(
			models.ErrFailedConversionCode,
			fmt.Sprintf("failed to parse request body; %s", err),
		), http.StatusBadRequest)
		return
	}
	pinStatus, err := Ctrl().CreatePin(uid, pin)
	if err != nil {
		logger.Error(err)
		server.JsonResponse(w, models.NewFailure(
			models.ErrProcessingRequestCode,
			fmt.Sprintf("failed to create pin; %s", err),
		), http.StatusInternalServerError)
		return
	}
	server.JsonResponse(w, &pinStatus, http.StatusAccepted)
}

func UpdatePin(w http.ResponseWriter, r *http.Request, p server.Parameters) {
	uid, ok := r.Context().Value(middleware.ClaimUserId).(string)
	if !ok {
		logger.Info("no user id")
	}
	id := p.Get("id")
	if id == "" {
		server.JsonResponse(w, models.NewFailure(
			models.ErrRequiredParamCode,
			"path parameter 'id' is required",
		), http.StatusBadRequest)
		return
	}
	var pin models.Pin
	if err := json.NewDecoder(r.Body).Decode(&pin); err != nil {
		server.JsonResponse(w, models.NewFailure(
			models.ErrFailedConversionCode,
			fmt.Sprintf("failed to parse request body; %s", err),
		), http.StatusBadRequest)
		return
	}
	pinStatus, err := Ctrl().UpdatePin(uid, id, pin)
	if err != nil {
		server.JsonResponse(w, models.NewFailure(
			models.ErrProcessingRequestCode,
			fmt.Sprintf("failed to update pin; %s", err),
		), http.StatusInternalServerError)
		return
	}
	server.JsonResponse(w, &pinStatus, http.StatusAccepted)
}

func PatchPin(w http.ResponseWriter, r *http.Request, p server.Parameters) {
	uid, ok := r.Context().Value(middleware.ClaimUserId).(string)
	if !ok {
		logger.Info("no user id")
	}
	id := p.Get("id")
	if id == "" {
		server.JsonResponse(w, models.NewFailure(
			models.ErrRequiredParamCode,
			"path parameter 'id' is required",
		), http.StatusBadRequest)
		return
	}
	var pin models.Pin
	if err := json.NewDecoder(r.Body).Decode(&pin); err != nil {
		server.JsonResponse(w, models.NewFailure(
			models.ErrFailedConversionCode,
			fmt.Sprintf("failed to parse request body; %s", err),
		), http.StatusBadRequest)
		return
	}
	pinStatus, err := Ctrl().PatchPin(uid, id, pin)
	if err != nil {
		server.JsonResponse(w, models.NewFailure(
			models.ErrProcessingRequestCode,
			fmt.Sprintf("failed to update pin; %s", err),
		), http.StatusInternalServerError)
		return
	}
	server.JsonResponse(w, &pinStatus, http.StatusAccepted)
}

func RemovePin(w http.ResponseWriter, r *http.Request, p server.Parameters) {
	uid, ok := r.Context().Value(middleware.ClaimUserId).(string)
	if !ok {
		logger.Info("no user id")
	}
	id := p.Get("id")
	if id == "" {
		server.JsonResponse(w, models.NewFailure(
			models.ErrRequiredParamCode,
			"path parameter 'id' is required",
		), http.StatusBadRequest)
		return
	}
	err := Ctrl().RemovePin(uid, id)
	if err != nil {
		server.JsonResponse(w, models.NewFailure(
			models.ErrProcessingRequestCode,
			fmt.Sprintf("failed to remove pin; %s", err),
		), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
