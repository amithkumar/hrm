/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

// ====CHAINCODE EXECUTION SAMPLES (CLI) ==================

// ==== Invoke marbles ====
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble1","blue","35","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble2","red","50","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["initMarble","marble3","blue","70","tom"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["transferMarble","marble2","jerry"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["transferMarblesBasedOnColor","blue","jerry"]}'
// peer chaincode invoke -C myc1 -n marbles -c '{"Args":["delete","marble1"]}'

// ==== Query marbles ====
// peer chaincode query -C myc1 -n marbles -c '{"Args":["readMarble","marble1"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["getMarblesByRange","marble1","marble3"]}'
// peer chaincode query -C myc1 -n marbles -c '{"Args":["getHistoryForMarble","marble1"]}'

// Rich Query (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarblesByOwner","tom"]}'
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"owner\":\"tom\"}}"]}'

//The following examples demonstrate creating indexes on CouchDB
//Example hostname:port configurations
//
//Docker or vagrant environments:
// http://couchdb:5984/
//
//Inside couchdb docker container
// http://127.0.0.1:5984/

// Index for chaincodeid, docType, owner.
// Note that docType and owner fields must be prefixed with the "data" wrapper
// chaincodeid must be added for all queries
//
// Definition for use with Fauxton interface
// {"index":{"fields":["chaincodeid","data.docType","data.owner"]},"ddoc":"indexOwnerDoc", "name":"indexOwner","type":"json"}
//
// example curl definition for use with command line
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"chaincodeid\",\"data.docType\",\"data.owner\"]},\"name\":\"indexOwner\",\"ddoc\":\"indexOwnerDoc\",\"type\":\"json\"}" http://hostname:port/myc1/_index
//

// Index for chaincodeid, docType, owner, size (descending order).
// Note that docType, owner and size fields must be prefixed with the "data" wrapper
// chaincodeid must be added for all queries
//
// Definition for use with Fauxton interface
// {"index":{"fields":[{"data.size":"desc"},{"chaincodeid":"desc"},{"data.docType":"desc"},{"data.owner":"desc"}]},"ddoc":"indexSizeSortDoc", "name":"indexSizeSortDesc","type":"json"}
//
// example curl definition for use with command line
// curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"data.size\":\"desc\"},{\"chaincodeid\":\"desc\"},{\"data.docType\":\"desc\"},{\"data.owner\":\"desc\"}]},\"ddoc\":\"indexSizeSortDoc\", \"name\":\"indexSizeSortDesc\",\"type\":\"json\"}" http://hostname:port/myc1/_index

// Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":\"marble\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

// Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
//   peer chaincode query -C myc1 -n marbles -c '{"Args":["queryMarbles","{\"selector\":{\"docType\":{\"$eq\":\"marble\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/satori/go.uuid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type marble struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Name       string `json:"name"`    //the fieldtags are needed to keep case from bouncing around
	Color      string `json:"color"`
	Size       int    `json:"size"`
	Owner      string `json:"owner"`
}

type healthRecord struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	RecordId string 	`json:"recordId"`
	PatientName string `json:"patientName"`
	DoctorName string `json:"doctorName"`
	// Name  ccv     string `json:"name"`    //the fieldtags are needed to keep case from bouncing around
	TestType  string `json:"testType"`
	Value       string    `json:"value"`
	
}

// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init initializes chaincode
// ===========================
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

// Invoke - Our entry point for Invocations
// ========================================
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("amith invoked is running " + function)

	// Handle different functions
	if function == "initHealthRecord" { //create a new health record
		return t.initHealthRecord(stub, args)
	} else if function == "readHealthRecord" { //read a health record
		return t.readHealthRecord(stub, args)
	// } else if function == "transferMarble" { //change owner of a specific marble
	// 	return t.transferMarble(stub, args)
	// } else if function == "transferMarblesBasedOnColor" { //transfer all marbles of a certain color
	// 	return t.transferMarblesBasedOnColor(stub, args)
	// } else if function == "delete" { //delete a marble
	// 	return t.delete(stub, args)
	// } else if function == "readMarble" { //read a marble
	// 	return t.readMarble(stub, args)
	// } else if function == "queryMarblesByOwner" { //find marbles for owner X using rich query
	// 	return t.queryMarblesByOwner(stub, args)
	// } else if function == "queryMarbles" { //find marbles based on an ad hoc rich query
	// 	return t.queryMarbles(stub, args)
	// } else if function == "getHistoryForMarble" { //get history of values for a marble
	// 	return t.getHistoryForMarble(stub, args)
	// } else if function == "getMarblesByRange" { //get marbles based on range query
	// 	return t.getMarblesByRange(stub, args)
	// }

	fmt.Println("amith invoke did not find func: " + function) //error
	return shim.Error("Received unknown function invocation")
}

// ============================================================
// initHealthRecord - create a new health record, store into chaincode state
// ============================================================
func (t *SimpleChaincode) initHealthRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var err error
	var jsonResp string

	//   0       1       2     3
	// "asdf", "blue", "35", "bob"
	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// ==== Input sanitation ====
	fmt.Println("- start init health record")
	if len(args[0]) <= 0 {
		return shim.Error("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return shim.Error("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return shim.Error("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return shim.Error("4th argument must be a non-empty string")
	}
	recordId := uuid.NewV4().String();
	patientName := args[0]
	doctorName := args[1]
	testType := args[2]
	value := args[3]

	// ==== Check if marble already exists ====
	// recordAsBytes, err := stub.GetState(recordId)
	// if err != nil {
	// 	return shim.Error("Failed to get marble: " + err.Error())
	// } else if marbleAsBytes != nil {
	// 	fmt.Println("This marble already exists: " + marbleName)
	// 	return shim.Error("This marble already exists: " + marbleName)
	// }

	// ==== Create marble object and marshal to JSON ====
	objectType := "healthRecord"
	record := &healthRecord{objectType, recordId, patientName, doctorName, testType, value}
	recordJSONasBytes, err := json.Marshal(record)
	if err != nil {
		return shim.Error(err.Error())
	}
	//Alternatively, build the marble json string manually if you don't want to use struct marshalling
	//marbleJSONasString := `{"docType":"Marble",  "name": "` + marbleName + `", "color": "` + color + `", "size": ` + strconv.Itoa(size) + `, "owner": "` + owner + `"}`
	//marbleJSONasBytes := []byte(str)

	// === Save marble to state ===
	err = stub.PutState(recordId, recordJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}
/*
	//  ==== Index the marble to enable color-based range queries, e.g. return all blue marbles ====
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~color~*
	indexName := "color~name"
	colorNameIndexKey, err := stub.CreateCompositeKey(indexName, []string{marble.Color, marble.Name})
	if err != nil {
		return shim.Error(err.Error())
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the marble.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(colorNameIndexKey, value)
	*/

	// ==== Marble saved and indexed. Return success ====
	fmt.Println("- end init marble")
	jsonResp = "{\"recordId\":\"" + recordId +"\"}"
	return shim.Success([]byte(jsonResp))
}

// ===============================================
// readHealthRecord - read a health from chaincode state
// ===============================================
func (t *SimpleChaincode) readHealthRecord(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var jsonResp, recordId string
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting record id of the health record to query")
	}

	recordId = args[0]
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to parse " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}
	valAsbytes, err := stub.GetState(recordId) //get the record from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + args[0] + "\"}"
		return shim.Error(jsonResp)
	} else if valAsbytes == nil {
		jsonResp = "{\"Error\":\"Marble does not exist: " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(valAsbytes)
}


