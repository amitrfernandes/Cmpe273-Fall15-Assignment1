package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Request struct {
	Name string `json:"name"`
}

type Response struct {
	Greet string `json:"greeting"`
}

func (res *Response) UnmarshalJSON(data []byte) error {
	var req Request

	// unmarshal the given data into req structure
	if err := json.Unmarshal(data, &req); err != nil {
		return err
	}

	// get the request present in 'req' and pass it in the Response structure
	res.Greet = "Hello, " + req.Name + "!"

	return nil
}

func post_h(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Stub an user to be populated from the body
	u := Request{}

	// Populate the request data
	json.NewDecoder(r.Body).Decode(&u)

	// Marshal provided interface into JSON structure
	uj, _ := json.Marshal(u)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)

	// Save the collected JSON data in another object
	b := uj

	// Create a new Response struct for storing the unmarshaled JSON struct
	var m Response
	if err := json.Unmarshal(b, &m); err != nil {
		fmt.Println(err)
		return
	}

	// Marshal provided request into JSON Response
	un, _ := json.Marshal(m)
	fmt.Fprintf(w, "%s", un)

}

func getuser(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {

	fmt.Fprintf(rw, "Hello, %s!\n", p.ByName("name"))
}

func main() {
	// Instantiate a new router
	r := httprouter.New()

	// Add a handler on /hello
	r.GET("/hello/:name", getuser)
	r.POST("/hello", post_h)

	// Fire up the server
	http.ListenAndServe("localhost:8080", r)
}

/***********************************************OUTPUT*****************


C:\Users\Amit\golang\Go\src\golang\httpjson
go run post_demo.go

C:\Users\Amit\golang\Go\src\golang\httpjson>
curl -H "Content-Type: application/json" -d '{"name":"Rock"}' http://localhost:8080/hello
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100    46  100    29  100    17   1812   1062 --:--:-- --:--:-- --:--:-- 29000
{"greeting":"Hello, Rock!"}

**********************************************************************/
