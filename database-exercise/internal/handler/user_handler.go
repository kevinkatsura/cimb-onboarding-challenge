package handler

import (
	"database-exercise/internal/response"
	"database-exercise/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string
		Email string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error(),
			"invalid request body")
		return
	}

	user, err := h.service.CreateUser(req.Name, req.Email)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(),
			"failed to create user")
		return
	}
	response.Success(w, user, "user created")
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		response.Error(w, http.StatusBadRequest, nil,
			"Missing user id")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err,
			"Invalid id format")
		return
	}

	user, err := h.service.GetUserByID(int64(id))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err,
			"failed to fetch user with id:"+idStr)
		return
	}
	response.Success(w, user, "")
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers()
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error(),
			"failed to fetch users")
		return
	}
	response.Success(w, users, "")
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		response.Error(w, http.StatusBadRequest, nil,
			"Missing user id")
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, nil,
			"Invalid id format")
		return
	}

	err = h.service.DeleteUser(int64(id))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err,
			"failed to delete user with id:"+idStr)
		return
	}
	response.Success(w, nil, "user id:"+idStr+" deleted successfully")
}
