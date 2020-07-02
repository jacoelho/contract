package contract_test

const fixture = `{
	"provider": {
	  "name": "a provider"
	},
	"consumer": {
	  "name": "a consumer"
	},
	"interactions": [
	  {
		"description": "request one",
		"request": {
		  "method": "get",
		  "path": "/path_one"
		},
		"response": {
		  "status" : 200
		},
		"provider_state": "state one"
	  }
	],
	"metadata": {
	  "pactSpecification": {
		"version": "1.0"
	  }
	}
  }`
