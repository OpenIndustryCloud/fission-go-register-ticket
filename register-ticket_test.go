package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	server *httptest.Server
	//Test Data TV
	userJson          = ` {"ticket":{"comment":{"html_body":"<p><b>If there has been any recent maintenance carried out on your home, please describe it<\/b> : No maintenance carried out<\/p><hr><p><b>If you have any other insurance or warranties covering your home, please advise us of the company name.<\/b> : No<\/p><hr><p><b>We have made the following assumptions about your property, you and anyone living with you<\/b> : <\/p><hr><p><b>When did the incident happen?<\/b> : 2017-01-01<\/p><hr><p><b>Are you still have possession of the damage items (i.e. damaged guttering)?<\/b> : <\/p><hr><p><b>Are you aware of anything else relevant to your claim that you would like to advise us of at this stage?<\/b> : I would need the vendors contact for repairing the roof<\/p><hr><p><b>Would you like to upload more images?<\/b> : <\/p><hr><p><b>Where did the incident happen? (City/town name)<\/b> : birmingham<\/p><hr><p><b>In as much detail as possible, please use the text box below to describe the full extent of the damage to your home and how you discovered it.<\/b> : Roof Damaged<\/p><hr><p><b>Please describe the details of the condition of your home prior to discovering the damage<\/b> : Tiles blown away<\/p><hr>"},"custom_fields":[{"id":114100596852,"value":"28"},{"id":114099964311,"value":"Storm Surge"},{"id":114100712171,"value":"50 : Possible Stormy weather"},{"id":114100658992,"value":"09876512345"},{"id":114100659172,"value":"amitkumarvarman@gmail.com"}],"requester":{"locale_id":1,"name":"Amit Varman","email":"amitkumarvarman@gmail.com"},"email":"amitkumarvarman@gmail.com","phone":"09876512345","priority":"normal","status":"new","subject":"Storm surge risk data","type":"incident","ticket_form_id":114093996871}}`
	userJsonMalformed = ` {"ticket":,t":{"html_body":"<p><b>If there has been any recent maintenance carried out on your home, please describe it<\/b> : No maintenance carried out<\/p><hr><p><b>If you have any other insurance or warranties covering your home, please advise us of the company name.<\/b> : No<\/p><hr><p><b>We have made the following assumptions about your property, you and anyone living with you<\/b> : <\/p><hr><p><b>When did the incident happen?<\/b> : 2017-01-01<\/p><hr><p><b>Are you still have possession of the damage items (i.e. damaged guttering)?<\/b> : <\/p><hr><p><b>Are you aware of anything else relevant to your claim that you would like to advise us of at this stage?<\/b> : I would need the vendors contact for repairing the roof<\/p><hr><p><b>Would you like to upload more images?<\/b> : <\/p><hr><p><b>Where did the incident happen? (City/town name)<\/b> : birmingham<\/p><hr><p><b>In as much detail as possible, please use the text box below to describe the full extent of the damage to your home and how you discovered it.<\/b> : Roof Damaged<\/p><hr><p><b>Please describe the details of the condition of your home prior to discovering the damage<\/b> : Tiles blown away<\/p><hr>"},"custom_fields":[{"id":114100596852,"value":"28"},{"id":114099964311,"value":"Storm Surge"},{"id":114100712171,"value":"50 : Possible Stormy weather"},{"id":114100658992,"value":"09876512345"},{"id":114100659172,"value":"amitkumarvarman@gmail.com"}],"requester":{"locale_id":1,"name":"Amit Varman","email":"amitkumarvarman@gmail.com"},"email":"amitkumarvarman@gmail.com","phone":"09876512345","priority":"normal","status":"new","subject":"Storm surge risk data","type":"incident","ticket_form_id":114093996871}}`
	// ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr = httptest.NewRecorder()
)

func TestHandler(t *testing.T) {
	//Convert string to reader and
	//Create request with JSON body
	req, err := http.NewRequest("POST", "", strings.NewReader(userJson))
	reqMalformed, err := http.NewRequest("POST", "", strings.NewReader(userJsonMalformed))
	if err != nil {
		t.Error(err) //Something is wrong while sending request
	}

	//TEST CASES
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"Test Data-1", args{rr, req}, 12},
		{"Test Data-2", args{rr, reqMalformed}, 0},
	}
	for _, tt := range tests {
		// call ServeHTTP method
		// directly and pass Request and ResponseRecorder.
		handler := http.HandlerFunc(Handler)
		handler.ServeHTTP(tt.args.w, tt.args.r)

		// Check the status code is what we expect.
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
		//check content type
		if ctype := rr.Header().Get("Content-Type"); ctype != "application/json" {
			t.Errorf("content type header does not match: got %v want %v",
				ctype, "application/json")
		}
		// check the output
		res, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error(err) //Something is wrong while read res
		}
		got := TicketResponse{}
		err = json.Unmarshal(res, &got)

		if err != nil && got.Audit.TicketID == tt.want {
			t.Errorf("%q. compute weather risk() = %v, want %v", tt.name, got.Audit.TicketID, tt.want)
		}
	}
}
