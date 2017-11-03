package main

/*
This API would accept JSON string as POST body and
create a Ticket in Zen Desk/Fresh Desk

INPUT - Zen Desk Create Ticket compliant JSON

OUTPUT - Ticket Meta Data JSON from Response Object
{
	"id": 133382282992,
	"ticket_id": 39,
	"created_at": "2017-10-25T18:32:55Z",
	"author_id": 115428050612,
	"metadata": {
		"system": {
		"ip_address": "2.122.25.146",
		"location": "Solihull, M2, United Kingdom",
		"latitude": 52.41669999999999,
		"longitude": -1.783299999999997
	},
	"custom": {}
	}
}
*/
import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//Default values
var (
	endPoint    = "https://landg.zendesk.com/api/v2/tickets.json"
	apiKey      = ""
	apiPassword = ""
	namesapce   = "default"
	secretName  = "zendesk-secret"
)

func Handler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Executing Register Ticket API end point...", endPoint)
	// validate request body
	if r.Body == nil {
		http.Error(w, "Please send a valid JSON", 400)
		createErrorResponse(w, "Please send a valid JSON", 400)
		return
	}

	//get API keys
	getAPIKeys(w)

	req, err := http.NewRequest("POST", endPoint, r.Body)
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(apiKey, apiPassword)

	client := &http.Client{}
	zendeskAPIResp, err := client.Do(req)

	if err != nil {
		http.Error(w, err.Error(), 400)
		createErrorResponse(w, err.Error(), 400)
		return
	}

	fmt.Println("request status for ticket creation :" + zendeskAPIResp.Status)

	var ticketResponse TicketResponse
	err = json.NewDecoder(zendeskAPIResp.Body).Decode(&ticketResponse)
	if err != nil {
		http.Error(w, err.Error(), 400)
		createErrorResponse(w, err.Error(), 400)
		return
	}
	defer zendeskAPIResp.Body.Close()

	//marshal response to JSON
	ticketAuditData := ticketResponse.Audit
	ticketResponseJSON, err := json.Marshal(&ticketAuditData)
	if err != nil {
		http.Error(w, err.Error(), 400)
		createErrorResponse(w, err.Error(), 400)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.Write([]byte(ticketResponseJSON))
}

func createErrorResponse(w http.ResponseWriter, message string, status int64) {
	errorJSON, _ := json.Marshal(&Error{
		Code:    status,
		Message: message})

	w.Header().Set("content-type", "application/json")
	w.Write([]byte(errorJSON))
}

type Error struct {
	Code    int64  `json:"status"`
	Message string `json:"message"`
}

// func main() {
// 	fmt.Println("staritng app..")
// 	http.HandleFunc("/", Handler)
// 	http.ListenAndServe(":8085", nil)
// }

func getAPIKeys(w http.ResponseWriter) {
	fmt.Println("[CONFIG] Reading Env variables")

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	secret, err := clientset.Core().Secrets(namesapce).Get(secretName, meta_v1.GetOptions{})
	fmt.Println("Zen Desk API Key : " + string(secret.Data[apiKey]))

	//endPointFromENV := os.Getenv("ENV_HELPDESK_API_EP")
	apiKey = string(secret.Data["apiKey"])
	apiPassword = string(secret.Data["password"])

	// if len(endPointFromENV) > 0 {
	// 	log.Print("[CONFIG] Setting Env variables", endPointFromENV)
	// 	endPoint = endPointFromENV
	// }
	if len(apiKey) == 0 {
		createErrorResponse(w, "Missing API Key", 400)
	}
	if len(apiPassword) == 0 {
		createErrorResponse(w, "Missing API Password", 400)
	}

}

type TicketResponse struct {
	Ticket struct {
		URL        string      `json:"url"`
		ID         int         `json:"id"`
		ExternalID interface{} `json:"external_id"`

		CreatedAt    time.Time   `json:"created_at"`
		UpdatedAt    time.Time   `json:"updated_at"`
		DueAt        interface{} `json:"due_at"`
		TicketFormID int64       `json:"ticket_form_id"`
	} `json:"ticket"`
	Audit struct {
		ID        int64     `json:"id"`
		TicketID  int       `json:"ticket_id"`
		CreatedAt time.Time `json:"created_at"`
		AuthorID  int64     `json:"author_id"`
		Metadata  struct {
			System struct {
				IPAddress string  `json:"ip_address"`
				Location  string  `json:"location"`
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"system"`
			Custom struct {
			} `json:"custom"`
		} `json:"metadata"`
	} `json:"audit"`
}
