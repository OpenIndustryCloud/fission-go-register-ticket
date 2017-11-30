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
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/mediocregopher/radix.v2/redis"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//Default values
var (
	endPoint     = "https://landg.zendesk.com/api/v2/tickets.json"
	apiKey       = ""
	apiPassword  = ""
	namesapce    = "default"
	secretName   = "zendesk-secret"
	REDIS_SERVER = "redis-redis.redis.svc.cluster.local:6379"
	TCP          = "tcp"
)

func Handler(w http.ResponseWriter, r *http.Request) {

	println("Executing Register Ticket API end point...", endPoint)

	buf, _ := ioutil.ReadAll(r.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	var ticketResponseJSON []byte
	ticketDetails := TicketDetails{}
	err := json.NewDecoder(rdr1).Decode(&ticketDetails)
	if err == io.EOF || err != nil {
		createErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	println("Submission ID to be validated ---> " + ticketDetails.Ticket.EventID)

	//check if ticker registered for this submission
	if validateRecord(w, ticketDetails.Ticket.EventID) == 1 {
		//get API keys
		getAPIKeys(w)

		req, err := http.NewRequest("POST", endPoint, rdr2)
		req.Header.Add("Content-Type", "application/json")
		req.SetBasicAuth(apiKey, apiPassword)
		client := &http.Client{}
		zendeskAPIResp, err := client.Do(req)
		if err != nil {
			createErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		} else if zendeskAPIResp.StatusCode != 201 {
			println("request status for ticket creation :" + zendeskAPIResp.Status)
			switch zendeskAPIResp.StatusCode {
			case 401:
				createErrorResponse(w, "Unauthorized", zendeskAPIResp.StatusCode)
			default:
				createErrorResponse(w, "error creating tickets", zendeskAPIResp.StatusCode)
			}
			return
		}
		var ticketResponse TicketResponse
		err = json.NewDecoder(zendeskAPIResp.Body).Decode(&ticketResponse)
		if err != nil || ticketResponse == (TicketResponse{}) {
			createErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer zendeskAPIResp.Body.Close()
		//marshal response to JSON
		ticketAuditData := ticketResponse.Audit
		ticketResponseJSON, err = json.Marshal(&ticketAuditData)
		if err != nil {
			createErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		}

	} else {
		ticketResponseJSON = []byte(`{"status":208,"message":"ticket Already created"}`)
	}
	w.Header().Set("content-type", "application/json")
	w.Write([]byte(ticketResponseJSON))
}

//this function stores submissionID to Redis DB
// in a SET, return 1 if inserted , 0 if already exists
func validateRecord(w http.ResponseWriter, submissionID string) int {

	println("Validating if record exist for submissionID", submissionID)

	if submissionID == "" {
		return 1 //cannot validate
	}
	//conn, err := redis.Dial("tcp", "localhost:6379")
	conn, err := redis.Dial(TCP, REDIS_SERVER)
	if err != nil {
		println("unable to create Redis Connection", err.Error())
		createErrorResponse(w, err.Error(), http.StatusBadRequest)
		return 0 //cannot validate
	}
	defer conn.Close()
	noOfRecord, err := conn.Cmd("SADD", "submissionID", submissionID).Int()
	// Check the Err field of the *Resp object for any errors.
	if err != nil {
		createErrorResponse(w, err.Error(), http.StatusBadRequest)
		return 0 //cannot validate
	}
	println("no of record added to redis DB : ", noOfRecord)
	return noOfRecord
}

func createErrorResponse(w http.ResponseWriter, message string, status int) {
	errorJSON, _ := json.Marshal(&Error{
		Status:  status,
		Message: message})
	//Send custom error message to caller
	w.WriteHeader(status)
	w.Header().Set("content-type", "application/json")
	w.Write([]byte(errorJSON))
}

type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func main() {
	println("staritng app.. :8085")
	http.HandleFunc("/", Handler)
	http.ListenAndServe(":8085", nil)
}

func getAPIKeys(w http.ResponseWriter) {
	println("[CONFIG] Reading Env variables")

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
	println(len(string(secret.Data["apiKey"])))

	//endPointFromENV := os.Getenv("ENV_HELPDESK_API_EP")
	apiKey = string(secret.Data["apiKey"])
	apiPassword = string(secret.Data["password"])

	// if len(endPointFromENV) > 0 {
	// 	log.Print("[CONFIG] Setting Env variables", endPointFromENV)
	// 	endPoint = endPointFromENV
	// }
	if len(apiKey) == 0 {
		createErrorResponse(w, "Missing API Key", http.StatusBadRequest)
	}
	if len(apiPassword) == 0 {
		createErrorResponse(w, "Missing API Password", http.StatusBadRequest)
	}

}

type TicketResponse struct {
	Ticket struct {
		URL        string      `json:"url,omitempty"`
		ID         int         `json:"id,omitempty"`
		ExternalID interface{} `json:"external_id,omitempty"`

		CreatedAt    time.Time   `json:"created_at,omitempty"`
		UpdatedAt    time.Time   `json:"updated_at,omitempty"`
		DueAt        interface{} `json:"due_at,omitempty"`
		TicketFormID int64       `json:"ticket_form_id,omitempty"`
	} `json:"ticket"`
	Audit struct {
		ID        int64     `json:"id,omitempty"`
		TicketID  int       `json:"ticket_id,omitempty"`
		CreatedAt time.Time `json:"created_at,omitempty"`
		AuthorID  int64     `json:"author_id,omitempty"`
		Metadata  struct {
			System struct {
				IPAddress string  `json:"ip_address,omitempty"`
				Location  string  `json:"location,omitempty"`
				Latitude  float64 `json:"latitude,omitempty"`
				Longitude float64 `json:"longitude,omitempty"`
			} `json:"system"`
			Custom struct {
			} `json:"custom"`
		} `json:"metadata"`
	} `json:"audit"`
}

type TicketDetails struct {
	Status int `json:"status"`
	Ticket struct {
		Type     string `json:"type"`
		Subject  string `json:"subject"`
		Priority string `json:"priority"`
		Status   string `json:"status"`
		Comment  struct {
			HTMLBody string   `json:"html_body"`
			Uploads  []string `json:"uploads,omitempty"`
		} `json:"comment"`
		CustomFields []CustomFields `json:"custom_fields,omitempty"`
		Requester    struct {
			LocaleID     int    `json:"locale_id"`
			Name         string `json:"name"`
			Email        string `json:"email"`
			Phone        string `json:"phone"`
			PolicyNumber string `json:"policy_number"`
		} `json:"requester"`
		TicketFormID int64     `json:"ticket_form_id"`
		EventID      string    `json:"event_id"`
		Token        string    `json:"token"`
		SubmittedAt  time.Time `json:"submitted_at"`
	} `json:"ticket"`
}

type CustomFields struct {
	ID    int64  `json:"id"`
	Value string `json:"value"`
}
