package users

import (
	"log"
	"net/http"

	"github.com/JeffreyOmoakah/AUTH.git/internal/json"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Signup(w http.ResponseWriter, r *http.Request){
	var tempSignupReq createSignupReq
	if err := json.Read(r, &tempSignupReq); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	newUser, err := h.service.Signup(r.Context(), tempSignupReq)
    if err != nil {
    	log.Println(err)
     
    	if err == ErrCredentialsRequired {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
   
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
    }
	
	
	json.Write(w, http.StatusCreated, newUser)
}

