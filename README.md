[![Coverage Status](https://coveralls.io/repos/github/OpenIndustryCloud/fission-go-register-ticket/badge.svg?branch=master)](https://coveralls.io/github/OpenIndustryCloud/fission-go-register-ticket?branch=master)


# Register Ticket API


`register-ticket.go` is an API which accepts JSON payload compliant to [Zendesk](https://www.zendesk.com/) and creates Ticket with the Payload Data


## Zendesk API reference

Register Ticket API is an implementation of [Zendesk API](https://developer.zendesk.com/rest_api/docs/core/tickets)

## Authentication

Authentication is implemented using API token, you can either configure a [Secret in Kubernetes](https://kubernetes.io/docs/concepts/configuration/secret/) 
or  have it directly configured within the API (not recommended as it exposes your secrets)

apiKey field should be - {enduser_email_address}/token

apiToken should be - {api_token}


## Error hanlding
- Technical error : `{"status":400,"message":"Error specific message"}`
- Duplicate Attempt Warning : `{"status":208,"message":"ticket Already created"}`

## Sample Input/Output

### Request payload
- Simple request payload

```{"ticket": {"subject": "My printer is on fire!", "comment": {"body": "The smoke is very colorful."}}}```

- Complex request payload 

` {"ticket":{"comment":{"html_body":"<p><b>If there has been any recent maintenance carried out on your home, please describe it<\/b> : No maintenance carried out<\/p><hr><p><b>If you have any other insurance or warranties covering your home, please advise us of the company name.<\/b> : No<\/p><hr><p><b>We have made the following assumptions about your property, you and anyone living with you<\/b> : <\/p><hr><p><b>When did the incident happen?<\/b> : 2017-01-01<\/p><hr><p><b>Are you still have possession of the damage items (i.e. damaged guttering)?<\/b> : <\/p><hr><p><b>Are you aware of anything else relevant to your claim that you would like to advise us of at this stage?<\/b> : I would need the vendors contact for repairing the roof<\/p><hr><p><b>Would you like to upload more images?<\/b> : <\/p><hr><p><b>Where did the incident happen? (City/town name)<\/b> : birmingham<\/p><hr><p><b>In as much detail as possible, please use the text box below to describe the full extent of the damage to your home and how you discovered it.<\/b> : Roof Damaged<\/p><hr><p><b>Please describe the details of the condition of your home prior to discovering the damage<\/b> : Tiles blown away<\/p><hr>"},"custom_fields":[{"id":114100596852,"value":"28"},{"id":114099964311,"value":"Storm Surge"},{"id":114100712171,"value":"50 : Possible Stormy weather"},{"id":114100658992,"value":"09876512345"},{"id":114100659172,"value":"amitkumarvarman@gmail.com"}],"requester":{"locale_id":1,"name":"Amit Varman","email":"amitkumarvarman@gmail.com"},"email":"amitkumarvarman@gmail.com","phone":"09876512345","priority":"normal","status":"new","subject":"Storm surge risk data","type":"incident","ticket_form_id":114093996871}}`

### API Response

On succesful call, API will create a ticket in Zendesk and return Ticket Meta Data

`{"id":133382282992,"ticket_id":39,"created_at":"2017-10-25T18:32:55Z","author_id":115428050612,"metadata":{"system":{"ip_address":"2.122.25.146","location":"Solihull, M2, United Kingdom","latitude":52.41669999999999,"longitude":-1.783299999999997},"custom":{}}}`

# Example Usage

1.  Deploy as Fission Functions

First, set up your fission deployment with the go environment.

```
fission env create --name go-env --image fission/go-env:1.8.1
```

To ensure that you build functions using the same version as the
runtime, fission provides a docker image and helper script for
building functions.



- Download the build helper script

```
$ curl https://raw.githubusercontent.com/fission/fission/master/environments/go/builder/go-function-build > go-function-build
$ chmod +x go-function-build
```

- Build the function as a plugin. Outputs result to 'function.so'

`$ go-function-build register-ticket.go`

- Upload the function to fission

`$ fission function create --name register-ticket --env go-env --package function.so`

- Map /register-ticket to the register-ticket function

`$ fission route create --method POST --url /register-ticket --function register-ticket`

- Run the function

```$ curl -d `{"ticket": {"subject": "My printer is on fire!", "comment": {"body": "The smoke is very colorful."}}}` -H "Content-Type: application/json" -X POST http://$FISSION_ROUTER/register-ticket```

2. Deploy as AWS Lambda

> to be updated