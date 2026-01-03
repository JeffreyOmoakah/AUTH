package users

import (
	"log"
	"log/slog"
	"net/http"

	"github.com/JeffreyOmoakah/AUTH.git/internal/json"
)

type handler struct {
	service Service
	logger  *slog.Logger
}

func NewHandler(service Service, logger *slog.Logger) *handler {
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

func (h *handler) Login(w http.ResponseWriter, r *http.Request){
	// 1. Parse the request body (Reuse your types)
    var req loginReq 
    if err := json.Read(r, &req); err != nil {
        h.logger.Error("failed to read login body", "error", err)
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }
    
    // 2. Call the service to verify credentials
      user, err := h.service.Login(r.Context(), req)
      if err != nil {
          // Use a generic 401 Unauthorized for security
          h.logger.Warn("login attempt failed", "email", req.Email, "error", err)
          http.Error(w, "invalid email or password", http.StatusUnauthorized)
          return
      }

      // 3. Generate the JWT 
        token, err := h.service.GenerateToken(user)
          if err != nil {
              h.logger.Error("token generation failed", "error", err)
              http.Error(w, "internal server error", http.StatusInternalServerError)
              return
          }
      
    // 4. Send the response
        json.Write(w, http.StatusOK, map[string]string{
              "token": token,
          })

}