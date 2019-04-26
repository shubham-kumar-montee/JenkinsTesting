package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// FunctionConditions : Mapping between FunctioName : Conditions []
type FunctionConditions struct {
	FunctionName   string `json:"FunctionName"`
	ConditionsList []ACL  `json:"ConditionsList"`
}

// ACL : Access Control List Grouping for restricting access
type ACL struct {
	Org  string `json:"Org"`
	Dept string `json:"Dept"`
	Role string `json:"Role"`
}

func checkAccessPermissions(stub shim.ChaincodeStubInterface, funcName string) bool {

	logger.Debug("Entry : checkAccessPermissions")

	isFound := false
	Identifier, _ := cid.GetID(stub)
	MSPID, _ := cid.GetMSPID(stub)
	logger.Info("Request is received from ID : ", Identifier, MSPID)

	// Certificate type check - only allow app , peer - denied
	typevalue, _, _ := cid.GetAttributeValue(stub, "hf.Type")
	logger.Info("Certificate type : ", typevalue)

	if typevalue != "app" {
		logger.Error("Invalid certificate type.")
		return isFound
	}

	deptValue, deptExists, deptErr := cid.GetAttributeValue(stub, "Dept")
	orgValue, orgExists, orgErr := cid.GetAttributeValue(stub, "Org")
	roleValue, roleExists, roleErr := cid.GetAttributeValue(stub, "Role")

	if !deptExists || !orgExists || !roleExists {
		logger.Error("Missing attributes in certificate.")
		return isFound
	}

	if deptErr != nil || orgErr != nil || roleErr != nil {
		logger.Error("Failed in fetching the certificate details.")
		return isFound
	}

	// Retrive the existing ACL from the state
	temp, err := stub.GetState("ACCESS_CONTROL_LIST")
	if err != nil {
		logger.Error("Unable to get state ACCESS_CONTROL_LIST")
		logger.Error(err.Error())
		return isFound
	}

	var tempMap = make(map[string][]ACL)
	err = json.Unmarshal(temp, &tempMap)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return isFound
	}

	tempConditions := tempMap[funcName] // Get All the conditions

	// Loop through the ACL array
	for _, acl := range tempConditions {
		if orgValue == acl.Org && deptValue == acl.Dept && roleValue == acl.Role {
			isFound = true
		}
	}
	logger.Debug("Exit : checkAccessPermissions")

	return isFound
}

func checkMemberInfoWithCertificate(stub shim.ChaincodeStubInterface, memberID string) bool {

	logger.Debug("Entry : checkAccessPermissions")

	isFound := false
	Identifier, _ := cid.GetID(stub)

	if len(memberID) < 1 {
		// Invalid Request ID
		logger.Error("Member ID is invalid", memberID)
	}

	compoundKey, errComp := stub.CreateCompositeKey("MEMBER", []string{memberID})
	fmt.Println("QueryObject() : Compound Key : ", compoundKey)
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
	}

	// Retrive the existing ACL from the state
	temp, err := stub.GetState(compoundKey)
	if err != nil {
		logger.Error("Unable to retrieve the given member id", memberID)
		logger.Error(err.Error())
	}

	var MemberInfo MemberDetails
	json.Unmarshal([]byte(temp), &MemberInfo)

	if MemberInfo.PublicKey == Identifier {
		isFound = true
	}

	return isFound
}

func isMemberExists(stub shim.ChaincodeStubInterface, memberID string) bool {

	isMember := false
	memberReqKey, errComp := stub.CreateCompositeKey("MEMBER", []string{memberID})
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
	}

	// Retrive the existing member
	temp, _ := stub.GetState(memberReqKey)
	if temp != nil {
		logger.Info("Member ID already exists, create a new one", memberID)
		isMember = true
	}

	return isMember
}
