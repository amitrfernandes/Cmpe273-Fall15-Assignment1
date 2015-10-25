package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Startresults struct {
	Results []struct {
		AddressComponents []struct {
			LongName  string   `json:"long_name"`
			ShortName string   `json:"short_name"`
			Types     []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry         struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport     struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string   `json:"place_id"`
		Types   []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

type postresp struct {
	Id      bson.ObjectId `json:"id" bson:"_id"`
	Name    string        `json:"name" bson:"name"`
	Address string        `json:"address" bson:"address"`
	City    string        `json:"city" bson:"city"`
	State   string        `json:"state" bson:"state"`
	Zip     string        `json:"zip" bson:"zip"`
	Loc     Cord          `json:"coordinate" bson:"coordinate"`
}

type Cord struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lng float64 `json:"lng" bson:"lng"`
}

// LocNavigator represents the controller for operating on the User resource
type LocNavigator struct {
	session *mgo.Session
}

// NewNavigator provides a reference to a LocNavigator with provided mongo session
func NewNavigator(s *mgo.Session) *LocNavigator {
	return &LocNavigator{s}
}

// GetLoc retrieves an individual location resource
func (ln LocNavigator) GetLoc(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Grab id
	id := p.ByName("id")

	// Verify id is ObjectId, otherwise bail
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	// Grab id
	oid := bson.ObjectIdHex(id)

	// Stub location response
	po := postresp{}

	// Fetch location from mongodb deployment
	if err := ln.session.DB("tripplanner").C("locationsC").FindId(oid).One(&po); err != nil {
		w.WriteHeader(404)
		return
	}

	json.NewDecoder(r.Body).Decode(po)

	// Marshal provided interface into JSON structure
	uj, _ := json.Marshal(po)

	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", uj)
}

func (ln LocNavigator) UpdateLoc(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	// Grab id
	id := p.ByName("id")

	// Verify id is ObjectId, otherwise bail
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	// Grab id
	oid := bson.ObjectIdHex(id)

	// Stub location resource for both the request and response
	po := postresp{}
	ps := postresp{}

	//set the response location id
	ps.Id = oid

	// Fetch location request body
	json.NewDecoder(r.Body).Decode(&ps)

	// retrieve the desired resource
	if err := ln.session.DB("tripplanner").C("locationsC").FindId(oid).One(&po); err != nil {
		w.WriteHeader(404)
		return
	}

	// store the name of the retrieved location resource
	na := po.Name

	// connect to mongolab session
	collections := ln.session.DB("tripplanner").C("locationsC")

	// retrieve the new address coordinates
	po = fetchdata(&ps)
	collections.Update(bson.M{"_id": oid}, bson.M{"$set": bson.M{"address": ps.Address, "city": ps.City, "state": ps.State, "zip": ps.Zip, "coordinate": bson.M{"lat": po.Loc.Lat, "lng": po.Loc.Lng}}})

	po.Name = na

	// Marshal the provided location interface into JSON structure
	uj, _ := json.Marshal(po)

	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)

}

func (ln LocNavigator) RemoveLoc(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	// Grab id
	id := p.ByName("id")

	// Verify id is ObjectId, otherwise bail
	if !bson.IsObjectIdHex(id) {
		w.WriteHeader(404)
		return
	}

	// Grab id
	oid := bson.ObjectIdHex(id)

	// Remove user
	if err := ln.session.DB("tripplanner").C("locationsC").RemoveId(oid); err != nil {
		w.WriteHeader(404)
		return
	}
	// Write status
	w.WriteHeader(200)
}

// CreateLoc creates a new location resource
func (ln LocNavigator) CreateLoc(w http.ResponseWriter, r *http.Request, p httprouter.Params) {

	// Stub a location resource to be populated from the body
	postrs := postresp{}

	// Populate the location data
	json.NewDecoder(r.Body).Decode(&postrs)

	// Retrieve the coordinates of the address
	neww := fetchdata(&postrs)

	// Add an Id
	neww.Id = bson.NewObjectId()

	// Write the user to mongo
	ln.session.DB("tripplanner").C("locationsC").Insert(neww)

	// Marshal provided interface into JSON structure
	uj, _ := json.Marshal(neww)

	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)
}

func fetchdata(rep *postresp) postresp {

	add := rep.Address
	//zi := rep.Zip
	ci := rep.City

	gs := strings.Replace(rep.State, " ", "+", -1)
	gadd := strings.Replace(add, " ", "+", -1)
	gci := strings.Replace(ci, " ", "+", -1)

	uri := "http://maps.google.com/maps/api/geocode/json?address=" + gadd + "+" + gci + "+" + gs + "&sensor=false"

	//  fmt.Println("URI IS ........... ")
	//  fmt.Println(uri)

	resp, _ := http.Get(uri)

	body, _ := ioutil.ReadAll(resp.Body)

	C := Startresults{}

	err := json.Unmarshal(body, &C)
	if err != nil {
		panic(err)
	}

	for _, Sample := range C.Results {
		rep.Loc.Lat = Sample.Geometry.Location.Lat
		rep.Loc.Lng = Sample.Geometry.Location.Lng
	}

	return *rep
}

func getSession() *mgo.Session {
	//    s, err := mgo.Dial("mongodb://localhost:27017")

	// Connect to the mongo deployment
	s, err := mgo.Dial("mongodb://admin:admin@ds043714.mongolab.com:43714/tripplanner")

	// Check if connection error, is mongo running?
	if err != nil {
		panic(err)
	}
	return s
}

func main() {

	r := httprouter.New()
	// Get a Location Navigator instance
	ln := NewNavigator(getSession())

	// Get a locations resource
	r.GET("/locations/:id", ln.GetLoc)
	r.POST("/locations", ln.CreateLoc)
	r.PUT("/locations/:id", ln.UpdateLoc)
	r.DELETE("/locations/:id", ln.RemoveLoc)

	http.ListenAndServe("localhost:3022", r)

}

/*****curl -XPOST -H 'Content-Type: application/json' -d '{"name": "Bob Smith", "address": "123 main street", "city" : "San Jose", "state" : "CA", "zip": "95112"}' http://localhost:3022/locations
 */
