package users 

type createSignupReq struct {
       Email    string `json:"email"`
       Password string `json:"password"`
}

type User struct {
    ID       int
    Email    string
    Password string
}

