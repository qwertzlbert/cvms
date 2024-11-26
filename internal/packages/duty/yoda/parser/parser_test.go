package parser_test

import (
	"testing"

	parser "github.com/cosmostation/cvms/internal/packages/duty/yoda/parser"
	"github.com/stretchr/testify/assert"
)

func TestParseYodaRequestParser(t *testing.T) {
	just_started_resp := []byte(`{
  "request": {
    "oracle_script_id": "46",
    "calldata": "",
    "requested_validators": [
      "bandvaloper1zm5p8gg3ugjcdwz9yrxaf6fdptxa4gw04rplr9",
      "bandvaloper1xs2penspev206jj0egh5qu7qmr6mjzfgj299xl"
    ],
    "min_count": "2",
    "request_height": "123",
    "request_time": "1732532680",
    "client_id": "110",
    "raw_requests": [
      {
        "external_id": "76",
        "data_source_id": "76",
        "calldata": "asdfqwertz"
      },
      {
        "external_id": "54",
        "data_source_id": "54",
        "calldata": "qwertzuio"
      }
    ],
    "ibc_channel": {
      "port_id": "oracle",
      "channel_id": "channel-7"
    },
    "execute_gas": "178750"
  },
  "reports": [
  ],
  "result": null
}`)

	long_expired_resp := []byte(`{
	  "request": null,
	  "reports": [
	  ],
	  "result": {
	    "client_id": "test2",
	    "oracle_script_id": "46",
	    "calldata": "=",
	    "ask_count": "16",
	    "min_count": "10",
	    "request_id": "24369521",
	    "ans_count": "16",
	    "request_time": "1732098953",
	    "resolve_time": "1732098966",
	    "resolve_status": "RESOLVE_STATUS_SUCCESS",
	    "result": "+=="
	  }
	}`)

	expired_missed_resp := []byte(`{
	  "request": {
	    "oracle_script_id": "46",
	    "calldata": "AAAAAgAAAANDUk8AAAADRVRIAw==",
	    "requested_validators": [
	      "bandvaloper1ycs4g7xu8wmf7n4vwwtfsvhtfm7tekvwk34rxz",
	      "bandvaloper1cuc594q8uam3rx90tztgplssnymlqnr5epmsjq",
	      "bandvaloper10grrhfawl98ypn5x25zmqcp8f20eg9sw56v3e0",
	      "bandvaloper1v7teuxaulwukplu90f0ra60vs2xpqystg89g95",
	      "bandvaloper17d3uhcjlh4jyqcep82jfsrg8avsngklxxan5tg",
	      "bandvaloper1wja9nds8klxcjurvxerdsun4t52zfjvqpj6y54",
	      "bandvaloper153znf8a89gmunexzrg2ud2yz47gyer9c05t2l8",
	      "bandvaloper1c9ye54e3pzwm3e0zpdlel6pnavrj9qqvn4hl7n",
	      "bandvaloper128dkddh98pfwdxg07t36fxedczthg0p2mdcs42",
	      "bandvaloper1zm5p8gg3ugjcdwz9yrxaf6fdptxa4gw04rplr9",
	      "bandvaloper1f6htx23e4xfu0dkpa2ck2kk63la2ej947cq0w4",
	      "bandvaloper1m88fha4982ev7smptzu8a7wvt8wkxdvfewh2md",
	      "bandvaloper1r00x80djyu6wkxpceegmvn5w9nx65prgqhxkzq",
	      "bandvaloper1dr64r506ln5n3aqqgkccf56uwdd9d9jvlh9l4p",
	      "bandvaloper1lksevk5362eterq75haf7wxx6jdpgdzyd9ms9h",
	      "bandvaloper1zsd6xc82dv2lqtt6hx9ew6hptlvtk0nyk7elqw"
	    ],
	    "min_count": "10",
	    "request_height": "1234",
	    "request_time": "1234",
	    "client_id": "test",
	    "raw_requests": [
	      {
	        "external_id": "76",
	        "data_source_id": "76",
	        "calldata": "RVRI"
	      },
	      {
	        "external_id": "54",
	        "data_source_id": "54",
	        "calldata": "RVRI"
	      }
	    ],
	    "ibc_channel": null,
	    "execute_gas": "86000"
	  },
	  "reports": [
	    {
	      "validator": "bandvaloper1lksevk5362eterq75haf7wxx6jdpgdzyd9ms9h",
	      "in_before_resolve": true,
	      "raw_reports": [
	        {
	          "external_id": "58",
	          "exit_code": 0,
	          "data": "asdf"
	        },
	        {
	          "external_id": "71",
	          "exit_code": 0,
	          "data": "asdf"
	        }
	      ]
	    }
	  ],
	  "result": {
	    "client_id": "test",
	    "oracle_script_id": "46",
	    "calldata": "asdf",
	    "ask_count": "16",
	    "min_count": "10",
	    "request_id": "24403290",
	    "ans_count": "16",
	    "request_time": "1234",
	    "resolve_time": "1234",
	    "resolve_status": "RESOLVE_STATUS_SUCCESS",
	    "result": "asdf"
	  }
	}`)

	requestBlock, validatorsFailedToRespond, status, err := parser.BandYodaRequestParser(just_started_resp)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, int64(123), requestBlock)
	assert.Equal(t, int(2), len(validatorsFailedToRespond))
	assert.Equal(t, string("running"), status)

	requestBlock, validatorsFailedToRespond, status, err = parser.BandYodaRequestParser(long_expired_resp)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, int64(0), requestBlock)
	assert.Equal(t, int(0), len(validatorsFailedToRespond))
	assert.Equal(t, string("success"), status)

	requestBlock, validatorsFailedToRespond, status, err = parser.BandYodaRequestParser(expired_missed_resp)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, int64(1234), requestBlock)
	assert.Equal(t, int(15), len(validatorsFailedToRespond))
	assert.Equal(t, string("success"), status)

}
