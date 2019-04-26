package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// Used for iterating ACL conditions
var functionConditionsMap = make(map[string][]ACL)
var fnCondArray []FunctionConditions
var logger = shim.NewLogger("LoyaltyChaincode")

// LoyaltyChaincode definition
type LoyaltyChaincode struct {
}

/*****************************************  CHAINCODE - INIT **********************************************/

// Init ACL & Reward points are intialized
func (t *LoyaltyChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	logger.SetLevel(shim.LogDebug)
	logger.Debug("Enter : Init - Fabric method")

	// 1. Setting Reward Points
	err := stub.PutState("TOTAL_REWARD_POINTS", []byte(strconv.Itoa(TotalRewardPoints)))
	if err != nil {
		logger.Error("Unable to init TOTAL_REWARD_POINTS")
		logger.Error(err)
		return shim.Error(err.Error())
	}

	logger.Info("TOTAL_REWARD_POINTS is initialized with points=", TotalRewardPoints)
	logger.Info("ACL used", ACL_DEFINITION)

	// 2. Setting ACL - JSON String
	ret := t.setACL(stub, ACL_DEFINITION)
	if !ret {
		logger.Error("Unable to set ACL conditions")
		return shim.Error("Unable to set ACL conditions")
	}

	// 3. Setting REQUEST_NO number
	err1 := stub.PutState("REQUEST_NO", []byte("0"))
	if err1 != nil {
		logger.Error("Unable to init REQUEST_NO")
		logger.Error(err)
		return shim.Error(err1.Error())
	}

	eventData := fmt.Sprintf("{ TOTAL_REWARD_POINTS: %s, ACL-SET: %t}", strconv.Itoa(TotalRewardPoints), true)
	eventPayload, _ := json.Marshal(eventData)

	stub.SetEvent("InitializeEvent", eventPayload)
	logger.Debug("Exit : Init - Fabric method")

	return shim.Success(nil)
}

// This method sets receives the access control list in the form of JSON and initializes
func (t *LoyaltyChaincode) setACL(stub shim.ChaincodeStubInterface, jsonACL string) bool {

	logger.Debug("Entry : setACL")
	//logger.Debug("ACL JSON string ", jsonACL)
	json.Unmarshal([]byte(jsonACL), &fnCondArray)

	for _, result := range fnCondArray {
		functionConditionsMap[result.FunctionName] = result.ConditionsList
	}

	temp, err1 := json.Marshal(functionConditionsMap)
	if err1 != nil {
		logger.Error("Unable to Marshal")
		logger.Error(err1.Error())
		return false
	}

	err := stub.PutState("ACCESS_CONTROL_LIST", temp)
	if err != nil {
		logger.Error("Unable to init ACCESS_CONTROL_LIST")
		logger.Error(err.Error())
		return false
	}

	logger.Debug("Exit : setACL")
	return true
}

// This method acts as a ROUTER in invoking different functions
func (t *LoyaltyChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	logger.Debug("Enter : Invoke - Fabric method")
	function, args := stub.GetFunctionAndParameters()

	if function == "requestRewardPoints" { // Org request for reward points
		return t.requestRewardPoints(stub, args[0], args[1])
	} else if function == "getRequestDetailByRequestID" { // Get the request id information
		return t.getRequestDetailByRequestID(stub, args[0], args[1])
	} else if function == "getAllRewardRequestDetails" { // Get All request details
		return t.getAllRewardRequestDetails(stub)
	} else if function == "getAllBurnRequestDetails" { // Get All request details
		return t.getAllBurnRequestDetails(stub)
	} else if function == "approveRequest" { // Approval from the other participating org memebers
		return t.approveRequest(stub, args[0], args[1], args[2])
	} else if function == "setMembershipIdentities" { // Set Member identities
		return t.setMembershipIdentities(stub, args[0], args[1], args[2], args[3])
	} else if function == "getAllMemberDetails" { // Get All member details
		return t.getAllMemberDetails(stub)
	} else if function == "getMemberDetailsByMemberID" { // Get Member details by member id
		return t.getMemberDetailsByMemberID(stub, args[0])
	} else if function == "burnRewardPoints" { // Org request for reward points
		return t.burnRewardPoints(stub, args[0], args[1])
	} else if function == "getRewardPoints" { // Get total reward points
		return t.getRewardPoints(stub)
	} else if function == "getAllACLConditions" { // Get the current ACL conditions
		return t.getAllACLConditions(stub)
	} else if function == "getACLConditionsByFuncAndOrg" { // Get the condition based on a function and org
		return t.getACLConditionsByFuncAndOrg(stub, args[0], args[1])
	} else if function == "updatePurchase" { // Consumer's purchase details sent by org
		return t.updatePurchase(stub, args[0], args[1], args[2], args[3], args[4], args[5])
	} else if function == "getPurchaseDetailsByPurchaseID" { // Get Purchase details by Purchase id
		return t.getPurchaseDetailsByPurchaseID(stub, args[0])
	} else if function == "getAllPurchaseDetails" { // Get All Purchase details
		return t.getAllPurchaseDetails(stub)
	} else if function == "transferRewardPoints" { // Transfer Reward points
		return t.transferRewardPoints(stub, args[0], args[1], args[2], args[3])
	} else if function == "getTransferDetailbyTransferID" { // Get transfer detail by transfer id
		return t.getTransferDetailbyTransferID(stub, args[0])
	} else if function == "getAllTransferDetails" { // Get all the transfer details
		return t.getAllTransferDetails(stub)
	} else if function == "getRewardPointsBalanceByMemberID" { // Get reward points balance using member id
		return t.getRewardPointsBalanceByMemberID(stub, args[0])
	}

	logger.Debug("Exit : Invoke - Fabric method")
	return shim.Error("Invalid function name." + function + ":Function does not exist")
}

// ******************************************* ACL RETRIEVAL FUNCTIONS ************************************************************//

// This method is used to retrieve access permission for the given method & org

func (t *LoyaltyChaincode) getACLConditionsByFuncAndOrg(stub shim.ChaincodeStubInterface, funcName string, orgName string) pb.Response {

	logger.Debug("Entry : getACLConditionsByFuncAndOrg")

	isAllowed := checkAccessPermissions(stub, "getACLConditionsByFuncAndOrg")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	// Retrive the existing ACL from the state
	temp, err := stub.GetState("ACCESS_CONTROL_LIST")
	if err != nil {
		logger.Error("Unable to get state ACCESS_CONTROL_LIST")
		logger.Error(err.Error())
		return shim.Error("Unable to get state ACCESS_CONTROL_LIST")
	}

	var tempMap = make(map[string][]ACL)
	err = json.Unmarshal(temp, &tempMap)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return shim.Error("Unable to get tempMap details")
	}

	tempConditions := tempMap[funcName] // Get All the conditions

	// Loop through the ACL array
	for _, acl := range tempConditions {
		logger.Debug("acl.Org", acl.Org)
		logger.Debug("orgName", orgName)

		if orgName == acl.Org { // If ORG Name matches then permissions isn't there
			return shim.Error("Permission Denied")
		}
	}

	logger.Debug("Exit : getACLConditionsByFuncAndOrg")
	return shim.Success(temp)
}

func (t *LoyaltyChaincode) getAllACLConditions(stub shim.ChaincodeStubInterface) pb.Response {

	logger.Debug("Entry : getACLConditions")

	isAllowed := checkAccessPermissions(stub, "getACLConditions")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	// Retrive the existing ACL from the state
	temp, err := stub.GetState("ACCESS_CONTROL_LIST")
	if err != nil {
		logger.Error("Unable to get state ACCESS_CONTROL_LIST")
		logger.Error(err.Error())
		return shim.Error("Unable to get state ACCESS_CONTROL_LIST")
	}

	logger.Debug("temp", string(temp))
	tempMap := make(map[string][]ACL)

	json.Unmarshal(temp, &tempMap) // Typecast into MAP

	logger.Debug("Exit : getACLConditions")
	return shim.Success(temp)
}

/*****************************************  REWARD POINTS - REQUEST, BURN, APPROVE, GET FUNCTIONS **********************************************/

// This method is used to find the number of reward points issued till now
func (t *LoyaltyChaincode) getRewardPoints(stub shim.ChaincodeStubInterface) pb.Response {

	logger.Debug("Entry : getRewardPoints")

	isAllowed := checkAccessPermissions(stub, "getRewardPoints")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	// GetState
	temp, err := stub.GetState("TOTAL_REWARD_POINTS")
	if err != nil {
		logger.Error("Unable to get TOTAL_REWARD_POINTS from World State")
		logger.Error(err.Error())
		return shim.Error("Unable to get TOTAL_REWARD_POINTS from World State")
	}

	logger.Debug("TOTAL_REWARD_POINTS=", fmt.Sprintf("%s", temp))
	logger.Debug("Exit : getRewardPoints")

	return shim.Success(temp)
}

// This method is used by participating orgs. Reward points are requested using this method
func (t *LoyaltyChaincode) requestRewardPoints(stub shim.ChaincodeStubInterface, rewardPts string, memberID string) pb.Response {

	logger.Debug("Entry : requestRewardPoints")

	isAllowed := checkAccessPermissions(stub, "requestRewardPoints")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	isMemberInfoValid := checkMemberInfoWithCertificate(stub, memberID)

	if !isMemberInfoValid {
		logger.Error("Permission denied. Please make sure member information is valid.")
		return shim.Error("Permission denied. Please make sure member information is valid.")
	}

	currentDate := time.Now().Format("2006-01-02")

	// GetState
	tempReqNoByte, err1 := stub.GetState("REQUEST_NO")
	if err1 != nil {
		logger.Error("Unable to get REQUEST_NO", tempReqNoByte)
		logger.Error(err1.Error())
		return shim.Error("Unable to get REQUEST_NO")
	}

	// Concat REQUEST_ + <id>
	newReqNoStr := fmt.Sprintf("%s", tempReqNoByte) // Byte to string
	newReqNoInt, _ := strconv.Atoi(newReqNoStr)     // String to Int
	newReqNo := strconv.Itoa(newReqNoInt + 1)       // Add int and convert to ascii

	temp := []string{"REQUEST_", newReqNo}
	reqID := strings.Join(temp, "")

	// 3. Setting REWARD_REQUEST number
	err2 := stub.PutState("REQUEST_NO", []byte(newReqNo))
	if err2 != nil {
		logger.Error("Unable to init REQUEST_NO")
		logger.Error(err2)
		return shim.Error(err2.Error())
	}

	// Create new Request
	var RequestDetails Request
	RequestDetails.RequestType = IssueRequest
	RequestDetails.RewardPoints, _ = strconv.Atoi(rewardPts)
	RequestDetails.RequestStatus = Requested
	RequestDetails.RequestedDate = currentDate
	RequestDetails.RequestedBy = memberID
	RequestDetails.RequestID = reqID
	RequestDetails.Identities = make(map[string]ApproveDetails)

	issueReqKey, errComp := stub.CreateCompositeKey("ISSUE_REQUEST", []string{reqID})
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return shim.Error("Unable to set composite key")
	}

	logger.Debug("Request ID : ", fmt.Sprintf("%s", issueReqKey))

	temp2, err2 := json.Marshal(RequestDetails)
	if err2 != nil {
		logger.Error("Unable to Marshal")
		logger.Error(err2.Error())
		return shim.Error("Unable to Marshal")
	}

	// Call PutState
	err := stub.PutState(issueReqKey, temp2)
	if err != nil {
		logger.Error("Unable to request reward points")
		logger.Error(err.Error())
		return shim.Error("Unable to request reward points")
	}

	eventData := fmt.Sprintf("{ REQUEST_ID: %s, REWARD_PTS: %d, REQUEST_DATE : %s , REQUEST_BY : %s}",
		reqID,
		RequestDetails.RewardPoints,
		RequestDetails.RequestedDate,
		RequestDetails.RequestedBy)

	eventPayload, _ := json.Marshal(eventData)
	stub.SetEvent("RequestRewardPtsEvent", eventPayload)

	logger.Debug("Exit : requestRewardPoints")
	return shim.Success([]byte(reqID))
}

// This method is used to approve request
func (t *LoyaltyChaincode) approveRequest(stub shim.ChaincodeStubInterface, requestID string, memberID string, requestType string) pb.Response {

	logger.Debug("Entry : approveRequest")

	isAllowed := checkAccessPermissions(stub, "requestRewardPoints")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	isMemberInfoValid := checkMemberInfoWithCertificate(stub, memberID)

	if !isMemberInfoValid {
		logger.Error("Permission denied. Please make sure member information is valid.")
		return shim.Error("Permission denied. Please make sure member information is valid.")
	}

	approvedDate := time.Now().Format("2006-01-02")

	compoundKey, errComp := stub.CreateCompositeKey(requestType, []string{requestID})
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return shim.Error("Unable to set composite key")
	}

	temp, err := stub.GetState(compoundKey)
	if temp == nil {
		logger.Error("Unable to retrieve the given request id", requestID)
		logger.Error(err.Error())
		return shim.Error("Unable to retrieve the given request id")
	}

	var tempRequestDetails Request
	json.Unmarshal([]byte(temp), &tempRequestDetails)

	// if tempRequestDetails.RequestedBy == memberID {
	// 	logger.Error("Requestor cannot approve his own request", requestID)
	// 	return shim.Error("Requestor cannot approve his own request")
	// }

	_, approvedAlready := tempRequestDetails.Identities[memberID]

	if approvedAlready {
		logger.Error("You have already approved this Request", requestID)
		return shim.Error("You have already approved this Request")
	}

	tempRequestDetails.Identities[memberID] = ApproveDetails{memberID, approvedDate}

	// TODO - Change based on consortia rules
	if len(tempRequestDetails.Identities) > 0 {
		tempRequestDetails.RequestStatus = Issued

	} else {
		tempRequestDetails.RequestStatus = PendingApproval
	}

	temp2, err2 := json.Marshal(tempRequestDetails)
	if err2 != nil {
		logger.Error("Unable to Marshal")
		logger.Error(err2.Error())
		return shim.Error("Unable to Marshal")
	}

	// Call PutState
	err3 := stub.PutState(compoundKey, temp2)
	if err3 != nil {
		logger.Error("Unable to request reward points")
		logger.Error(err3.Error())
		return shim.Error("Unable to request reward points")
	}

	// TODO - Change based on consortia rules
	if len(tempRequestDetails.Identities) > 0 {

		// Minus Reward Points from Max Cap and assign it to Identity
		tempRew, err := stub.GetState("TOTAL_REWARD_POINTS")
		if err != nil {
			logger.Error("Unable to get TOTAL_REWARD_POINTS from World State")
			logger.Error(err.Error())
			return shim.Error("Unable to get TOTAL_REWARD_POINTS from World State")
		}

		totalPoints, _ := strconv.Atoi(fmt.Sprintf("%s", tempRew))
		totalPoints = totalPoints - tempRequestDetails.RewardPoints

		// 1. Setting Reward Points
		errPoint := stub.PutState("TOTAL_REWARD_POINTS", []byte(strconv.Itoa(totalPoints)))
		if errPoint != nil {
			logger.Error("Unable to minus TOTAL_REWARD_POINTS")
			logger.Error(err)
			return shim.Error(err.Error())
		}

		if len(memberID) < 1 {
			// Invalid Request ID
			logger.Error("Member ID is invalid", memberID)
			return shim.Error("Member ID is invalid")
		}

		compoundMemberKey, errComp := stub.CreateCompositeKey("MEMBER", []string{memberID})
		fmt.Println("QueryObject() : Compound Key : ", compoundKey)
		if errComp != nil {
			logger.Error("Unable to set composite key")
			logger.Error(errComp.Error())
			return shim.Error("Unable to set composite key")
		}

		// Retrive the existing ACL from the state
		tempMem, err := stub.GetState(compoundMemberKey)
		if tempMem == nil {
			logger.Error("Unable to retrieve the given member id", memberID)
			logger.Error(err.Error())
			return shim.Error("Unable to retrieve the given memberID ")
		}

		var tempMemberInfo MemberDetails
		json.Unmarshal([]byte(tempMem), &tempMemberInfo)
		tempMemberInfo.RewardPoints = tempRequestDetails.RewardPoints

		temp3, err2 := json.Marshal(tempMemberInfo)
		if err2 != nil {
			logger.Error("Unable to Marshal")
			logger.Error(err2.Error())
			return shim.Error("Unable to Marshal")
		}

		err1 := stub.PutState(compoundMemberKey, temp3)
		if err1 != nil {
			logger.Error("Unable to update reward points for the member", memberID)
			logger.Error(err1.Error())
			return shim.Error("Unable to update reward points for the member")
		}

	} else {
		tempRequestDetails.RequestStatus = PendingApproval
	}

	eventData := fmt.Sprintf("{ REQUEST_ID: %s, APPROVED_BY : %s, APPROVED_DATE : %s }",
		requestID,
		memberID,
		approvedDate)

	eventPayload, _ := json.Marshal(eventData)

	stub.SetEvent("ApproveEvent", eventPayload)

	logger.Debug("Exit : approveRequest")

	return shim.Success(nil)
}

// This method is used by participating orgs. Reward points are requested using this method
func (t *LoyaltyChaincode) burnRewardPoints(stub shim.ChaincodeStubInterface, rewardPts string, memberID string) pb.Response {

	logger.Debug("Entry : burnRewardPoints")
	isAllowed := checkAccessPermissions(stub, "burnRewardPoints")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	isMemberInfoValid := checkMemberInfoWithCertificate(stub, memberID)

	if !isMemberInfoValid {
		logger.Error("Permission denied. Please make sure member information is valid.")
		return shim.Error("Permission denied. Please make sure member information is valid.")
	}

	currentDate := time.Now().Format("2006-01-02")

	// GetState
	tempReqNoByte, err1 := stub.GetState("REQUEST_NO")
	if err1 != nil {
		logger.Error("Unable to get REQUEST_NO", tempReqNoByte)
		logger.Error(err1.Error())
		return shim.Error("Unable to get REQUEST_NO")
	}

	// Concat REQUEST_ + <id>
	newReqNoStr := fmt.Sprintf("%s", tempReqNoByte) // Byte to string
	newReqNoInt, _ := strconv.Atoi(newReqNoStr)     // String to Int
	newReqNo := strconv.Itoa(newReqNoInt + 1)       // Add int and convert to ascii

	temp := []string{"REQUEST_", newReqNo}
	reqID := strings.Join(temp, "")

	// 3. Setting REWARD_REQUEST number
	err2 := stub.PutState("REQUEST_NO", []byte(newReqNo))
	if err2 != nil {
		logger.Error("Unable to init REQUEST_NO")
		logger.Error(err2)
		return shim.Error(err2.Error())
	}

	// Create new Request
	var RequestDetails Request
	RequestDetails.RequestType = BurnRequest
	RequestDetails.RewardPoints, _ = strconv.Atoi(rewardPts)
	RequestDetails.RequestStatus = Requested
	RequestDetails.RequestedDate = currentDate
	RequestDetails.RequestedBy = memberID
	RequestDetails.RequestID = reqID
	RequestDetails.Identities = make(map[string]ApproveDetails)

	burnReqKey, errComp := stub.CreateCompositeKey("BURN_REQUEST", []string{reqID})

	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return shim.Error("Unable to set composite key")
	}

	temp2, err2 := json.Marshal(RequestDetails)
	if err2 != nil {
		logger.Error("Unable to Marshal")
		logger.Error(err2.Error())
		return shim.Error("Unable to Marshal")
	}

	// Call PutState
	err := stub.PutState(burnReqKey, temp2)
	if err != nil {
		logger.Error("Unable to request reward points")
		logger.Error(err.Error())
		return shim.Error("Unable to request reward points")
	}

	eventData := fmt.Sprintf("{ REQUEST_ID: %s, REWARD_PTS : %d, REQUEST_DATE : %s, REQUEST_BY : %s}",
		reqID,
		RequestDetails.RewardPoints,
		RequestDetails.RequestedDate,
		RequestDetails.RequestedBy)

	eventPayload, _ := json.Marshal(eventData)

	stub.SetEvent("BurnRewardPtsEvent", eventPayload)

	logger.Debug("Exit : burnRewardPoints")
	return shim.Success([]byte(reqID))
}

// This method is used to get request details by request ID
func (t *LoyaltyChaincode) getRequestDetailByRequestID(stub shim.ChaincodeStubInterface, reqID string, requestType string) pb.Response {

	logger.Debug("Entry : getRequestDetailByRequestID")
	isAllowed := checkAccessPermissions(stub, "getRequestDetailByRequestID")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	if len(reqID) < 1 {
		// Invalid Request ID
		logger.Error("Request ID is invalid", reqID)
		return shim.Error("Request ID is invalid")
	}

	compoundKey, errComp := stub.CreateCompositeKey(requestType, []string{reqID})
	fmt.Println("QueryObject() : Compound Key : ", compoundKey)
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return shim.Error("Unable to set composite key")
	}

	// Retrive the existing ACL from the state
	temp, err := stub.GetState(compoundKey)
	if err != nil {
		logger.Error("Unable to retrieve the given request id", reqID)
		logger.Error(err.Error())
		return shim.Error("Unable to retrieve the given request id")
	}

	logger.Debug("temp", string(temp))

	//var RequestDetails Request
	//json.Unmarshal([]byte(temp), &RequestDetails) // Typecast into RequestDetailsMAP

	logger.Debug("Exit : getRequestDetailByRequestID")
	return shim.Success(temp)
}

/*****************************************  PURCHASE - PURCHASE/RECEIVE REWARD POINTS, GET FUNCTIONS **********************************************/
// UPDATE PURCHASE INFORMATION IN THE LEDGER
func (t *LoyaltyChaincode) updatePurchase(stub shim.ChaincodeStubInterface, purchaseID string, purchaser string, purchaseReceiptFile string,
	purchaseReceiptHash string, rewardPtsElig string, memberID string) pb.Response {

	logger.Debug("Entry : updatePurchase")
	isAllowed := checkAccessPermissions(stub, "updatePurchase")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	isMemberInfoValid := checkMemberInfoWithCertificate(stub, memberID)

	if !isMemberInfoValid {
		logger.Error("Permission denied. Please make sure member information is valid.")
		return shim.Error("Permission denied. Please make sure member information is valid.")
	}

	Org, _, _ := cid.GetAttributeValue(stub, "Org")
	points, _ := strconv.Atoi(rewardPtsElig)
	currentDate := time.Now().Format("2006-01-02")

	isPurchaserValid := isMemberExists(stub, purchaser)

	if !isPurchaserValid {
		logger.Error("Purchaser doesn't exists.")
		return shim.Error("Purchaser doesn't exists.")
	}

	// Add Purchase Information
	var PurchaseInfo PurchaseDetails

	purchaseKey, errComp := stub.CreateCompositeKey("PURCHASE_REF_ID", []string{Org, purchaseID})
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return shim.Error("Unable to set composite key")
	}

	// Retrive the existing member
	tempKey, err := stub.GetState(purchaseKey)
	if tempKey != nil {
		logger.Error("Purchase ID already exists, create a new one", purchaseKey)
		logger.Error(err.Error())
		return shim.Error("Purchase ID already exists, create a new one")
	}

	PurchaseInfo.PurchaseID = purchaseID
	PurchaseInfo.PurchaseReceiptFile = purchaseReceiptFile
	PurchaseInfo.PurchaseReceiptHash = purchaseReceiptHash
	PurchaseInfo.IssuedOrg = Org
	PurchaseInfo.IssuedMember = memberID
	PurchaseInfo.PurchaseBy = purchaser
	PurchaseInfo.PurchaseDate = currentDate
	PurchaseInfo.RewardPtsElig = points
	PurchaseInfo.RewardPtsTrans = points // Points transfered for this request
	var transferID string
	var isFailure bool
	if points > 0 {
		isFailure, transferID = quickTransfer(stub, memberID, purchaser, rewardPtsElig, "via Purchase")

		if isFailure {
			logger.Error("Unable to transfer reward points")
			return shim.Error("Unable to transfer reward points")
		}

		PurchaseInfo.TransferID = transferID
	}

	temp2, err2 := json.Marshal(PurchaseInfo)
	if err2 != nil {
		logger.Error("Unable to Marshal")
		logger.Error(err2.Error())
		return shim.Error("Unable to Marshal")
	}

	// Call PutState
	err3 := stub.PutState(purchaseKey, temp2)
	if err3 != nil {
		logger.Error("Unable to add purchase information")
		logger.Error(err3.Error())
		return shim.Error("Unable to add purchase information")
	}

	eventData := fmt.Sprintf("{ PURCHASE_REF_ID: %s, REWARD_PTS_ELIG : %s, REQUEST_DATE : %s, REQUEST_BY : %s, TRANSFER_ID : %s}",
		purchaseID,
		rewardPtsElig,
		currentDate,
		PurchaseInfo.IssuedMember, transferID)

	eventPayload, _ := json.Marshal(eventData)

	stub.SetEvent("PurchaseEvent", eventPayload)

	logger.Debug("Exit : updatePurchase")
	return shim.Success([]byte(purchaseID))
}

// This method is used to get request details by request ID
func (t *LoyaltyChaincode) getPurchaseDetailsByPurchaseID(stub shim.ChaincodeStubInterface, purchaseID string) pb.Response {

	logger.Debug("Entry : getPurchaseDetailsByPurchaseID")
	isAllowed := checkAccessPermissions(stub, "getPurchaseDetailsByPurchaseID")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	Org, _, _ := cid.GetAttributeValue(stub, "Org")

	if len(purchaseID) < 1 {
		// Invalid Request ID
		logger.Error("purchaseID  is invalid", purchaseID)
		return shim.Error("purchaseID  is invalid")
	}

	compoundKey, errComp := stub.CreateCompositeKey("PURCHASE_REF_ID", []string{Org, purchaseID})
	fmt.Println("QueryObject() : Compound Key : ", compoundKey)
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return shim.Error("Unable to set composite key")
	}

	temp, err := stub.GetState(compoundKey)
	if err != nil {
		logger.Error("Unable to retrieve the given purchaseID", temp)
		logger.Error(err.Error())
		return shim.Error("Unable to retrieve the given purchaseID")
	}

	logger.Debug("Exit : getPurchaseDetailsByPurchaseID")
	return shim.Success(temp)
}

/*****************************************  TRANSFER - REWARD POINTS, GET FUNCTIONS **********************************************/

func quickTransfer(stub shim.ChaincodeStubInterface, from string, to string, value string, remarks string) (bool, string) {

	logger.Debug("Entry : quickTransfer")
	isFailure := true
	isAllowed := checkAccessPermissions(stub, "quickTransfer")
	var transferID string

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return isFailure, transferID
	}

	isMemberFrom := isMemberExists(stub, from)

	if !isMemberFrom {
		logger.Error("from Member id doesnt exists.")
		return isFailure, transferID
	}

	isMemberTo := isMemberExists(stub, to)

	if !isMemberTo {
		logger.Error("to Member id doesnt exists.")
		return isFailure, transferID
	}

	// GetState
	tempReqNoByte, err1 := stub.GetState("REQUEST_NO")
	if err1 != nil {
		logger.Error("Unable to get REQUEST_NO", tempReqNoByte)
		logger.Error(err1.Error())
		return isFailure, transferID
	}

	// Concat REQUEST_ + <id>
	newReqNoStr := fmt.Sprintf("%s", tempReqNoByte) // Byte to string
	newReqNoInt, _ := strconv.Atoi(newReqNoStr)     // String to Int
	newReqNo := strconv.Itoa(newReqNoInt + 1)       // Add int and convert to ascii

	// 3. Setting REWARD_REQUEST number
	err2 := stub.PutState("REQUEST_NO", []byte(newReqNo))
	if err2 != nil {
		logger.Error("Unable to init REQUEST_NO")
		logger.Error(err2)
		return isFailure, transferID
	}

	temp := []string{"TRANSFER_", newReqNo}
	transferID = strings.Join(temp, "")

	transferKey, errComp := stub.CreateCompositeKey("TRANSFER", []string{transferID})
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return isFailure, transferID
	}

	var transferInfo TransferDetails
	transferInfo.TransferID = transferID
	transferInfo.From = from
	transferInfo.To = to
	transferInfo.Value, _ = strconv.Atoi(value)
	transferInfo.Remarks = remarks

	temp2, err2 := json.Marshal(transferInfo)
	if err2 != nil {
		logger.Error("Unable to Marshal")
		logger.Error(err2.Error())
		return isFailure, transferID
	}

	err := stub.PutState(transferKey, temp2)
	if err != nil {
		logger.Error("Unable to transfer reward points ")
		logger.Error(err.Error())
		return isFailure, transferID
	}

	// UPDATING FROM

	compoundMemberKey, errComp := stub.CreateCompositeKey("MEMBER", []string{from})
	fmt.Println("QueryObject() : Compound Key : ", compoundMemberKey)
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return isFailure, transferID
	}

	// Retrive the existing ACL from the state
	tempMem, err := stub.GetState(compoundMemberKey)
	if tempMem == nil {
		logger.Error("Unable to retrieve the given member id", from)
		logger.Error(err.Error())
		return isFailure, transferID
	}

	var tempMemberInfo MemberDetails
	json.Unmarshal([]byte(tempMem), &tempMemberInfo)

	if tempMemberInfo.RewardPoints < transferInfo.Value {
		logger.Error("Dont have enough points to transfer")
		return isFailure, transferID
	}

	tempMemberInfo.RewardPoints = tempMemberInfo.RewardPoints - transferInfo.Value
	temp3, err2 := json.Marshal(tempMemberInfo)
	if err2 != nil {
		logger.Error("Unable to Marshal")
		logger.Error(err2.Error())
		return isFailure, transferID
	}

	errFrom := stub.PutState(compoundMemberKey, temp3)
	if errFrom != nil {
		logger.Error("Unable to update reward points for the member", from)
		logger.Error(err1.Error())
		return isFailure, transferID
	}

	// UPDATING TO

	compoundToKey, errComp := stub.CreateCompositeKey("MEMBER", []string{to})
	fmt.Println("QueryObject() : Compound Key : ", compoundToKey)
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return isFailure, transferID
	}

	// Retrive the existing ACL from the state
	tempMemTo, err := stub.GetState(compoundToKey)
	if tempMemTo == nil {
		logger.Error("Unable to retrieve the given member id", to)
		logger.Error(err.Error())
		return isFailure, transferID
	}

	var tempMemberToInfo MemberDetails
	json.Unmarshal([]byte(tempMemTo), &tempMemberToInfo)
	tempMemberToInfo.RewardPoints = tempMemberToInfo.RewardPoints + transferInfo.Value

	temp4, err4 := json.Marshal(tempMemberToInfo)
	if err4 != nil {
		logger.Error("Unable to Marshal")
		logger.Error(err2.Error())
		return isFailure, transferID
	}

	errTo := stub.PutState(compoundToKey, temp4)
	if errTo != nil {
		logger.Error("Unable to update reward points for the member", to)
		logger.Error(err1.Error())
		return isFailure, transferID
	}

	isFailure = false
	logger.Debug("Exit : quickTransfer")

	return isFailure, transferID
}

// UPDATE PURCHASE INFORMATION IN THE LEDGER
func (t *LoyaltyChaincode) transferRewardPoints(stub shim.ChaincodeStubInterface, from string, to string, value string, remarks string) pb.Response {

	logger.Debug("Entry : transferRewardPoints")
	isAllowed := checkAccessPermissions(stub, "transferRewardPoints")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	isMemberFrom := isMemberExists(stub, from)

	if !isMemberFrom {
		logger.Error("from Member id doesnt exists.")
		return shim.Error("from Member id doesnt exists.")
	}

	isMemberTo := isMemberExists(stub, to)

	if !isMemberTo {
		logger.Error("to Member id doesnt exists.")
		return shim.Error("to Member id doesnt exists.")
	}

	// GetState
	tempReqNoByte, err1 := stub.GetState("REQUEST_NO")
	if err1 != nil {
		logger.Error("Unable to get REQUEST_NO", tempReqNoByte)
		logger.Error(err1.Error())
		return shim.Error("Unable to get REQUEST_NO")
	}

	// Concat REQUEST_ + <id>
	newReqNoStr := fmt.Sprintf("%s", tempReqNoByte) // Byte to string
	newReqNoInt, _ := strconv.Atoi(newReqNoStr)     // String to Int
	newReqNo := strconv.Itoa(newReqNoInt + 1)       // Add int and convert to ascii

	// 3. Setting REWARD_REQUEST number
	err2 := stub.PutState("REQUEST_NO", []byte(newReqNo))
	if err2 != nil {
		logger.Error("Unable to init REQUEST_NO")
		logger.Error(err2)
		return shim.Error(err2.Error())
	}

	temp := []string{"TRANSFER_", newReqNo}
	transferID := strings.Join(temp, "")

	transferKey, errComp := stub.CreateCompositeKey("TRANSFER", []string{transferID})
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return shim.Error("Unable to set composite key")
	}

	var transferInfo TransferDetails
	transferInfo.TransferID = transferID
	transferInfo.From = from
	transferInfo.To = to
	transferInfo.Value, _ = strconv.Atoi(value)
	transferInfo.Remarks = remarks

	temp2, err2 := json.Marshal(transferInfo)
	if err2 != nil {
		logger.Error("Unable to Marshal")
		logger.Error(err2.Error())
		return shim.Error("Unable to Marshal")
	}

	err := stub.PutState(transferKey, temp2)
	if err != nil {
		logger.Error("Unable to transfer reward points ")
		logger.Error(err.Error())
		return shim.Error("Unable to transfer reward points")
	}

	// UPDATING FROM

	compoundMemberKey, errComp := stub.CreateCompositeKey("MEMBER", []string{from})
	fmt.Println("QueryObject() : Compound Key : ", compoundMemberKey)
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return shim.Error("Unable to set composite key")
	}

	// Retrive the existing ACL from the state
	tempMem, err := stub.GetState(compoundMemberKey)
	if tempMem == nil {
		logger.Error("Unable to retrieve the given member id", from)
		logger.Error(err.Error())
		return shim.Error("Unable to retrieve the given memberID ")
	}

	var tempMemberInfo MemberDetails
	json.Unmarshal([]byte(tempMem), &tempMemberInfo)

	if tempMemberInfo.RewardPoints < transferInfo.Value {
		logger.Error("Dont have enough points to transfer")
		return shim.Error("Dont have enough points to transfer")
	}

	tempMemberInfo.RewardPoints = tempMemberInfo.RewardPoints - transferInfo.Value
	temp3, err2 := json.Marshal(tempMemberInfo)
	if err2 != nil {
		logger.Error("Unable to Marshal")
		logger.Error(err2.Error())
		return shim.Error("Unable to Marshal")
	}

	errFrom := stub.PutState(compoundMemberKey, temp3)
	if errFrom != nil {
		logger.Error("Unable to update reward points for the member", from)
		logger.Error(err1.Error())
		return shim.Error("Unable to update reward points for the member")
	}

	// UPDATING TO

	compoundToKey, errComp := stub.CreateCompositeKey("MEMBER", []string{to})
	fmt.Println("QueryObject() : Compound Key : ", compoundToKey)
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return shim.Error("Unable to set composite key")
	}

	// Retrive the existing ACL from the state
	tempMemTo, err := stub.GetState(compoundToKey)
	if tempMemTo == nil {
		logger.Error("Unable to retrieve the given member id", to)
		logger.Error(err.Error())
		return shim.Error("Unable to retrieve the given memberID ")
	}

	var tempMemberToInfo MemberDetails
	json.Unmarshal([]byte(tempMemTo), &tempMemberToInfo)

	tempMemberToInfo.RewardPoints = tempMemberToInfo.RewardPoints + transferInfo.Value
	temp4, err4 := json.Marshal(tempMemberToInfo)
	if err4 != nil {
		logger.Error("Unable to Marshal")
		logger.Error(err2.Error())
		return shim.Error("Unable to Marshal")
	}

	errTo := stub.PutState(compoundToKey, temp4)
	if errTo != nil {
		logger.Error("Unable to update reward points for the member", to)
		logger.Error(err1.Error())
		return shim.Error("Unable to update reward points for the member")
	}

	eventData := fmt.Sprintf("{ TRANSFER_ID: %s, FROM : %s, TO : %s, VALUE : %d}",
		transferID,
		transferInfo.From,
		transferInfo.To,
		transferInfo.Value)

	eventPayload, _ := json.Marshal(eventData)

	stub.SetEvent("TransferEvent", eventPayload)
	logger.Debug("Exit : transferRewardPoints")
	return shim.Success([]byte(nil))

}

// This method is used to get request details by request ID
func (t *LoyaltyChaincode) getTransferDetailbyTransferID(stub shim.ChaincodeStubInterface, transferID string) pb.Response {

	logger.Debug("Entry : getTransferDetailbyTransferID")
	isAllowed := checkAccessPermissions(stub, "getTransferDetailbyTransferID")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	if len(transferID) < 1 {
		// Invalid Request ID
		logger.Error("Request ID is invalid", transferID)
		return shim.Error("Request ID is invalid")
	}

	compoundKey, errComp := stub.CreateCompositeKey("TRANSFER", []string{transferID})
	fmt.Println("QueryObject() : Compound Key : ", compoundKey)
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return shim.Error("Unable to set composite key")
	}

	// Retrive the existing ACL from the state
	temp, err := stub.GetState(compoundKey)
	if err != nil {
		logger.Error("Unable to retrieve the given request id", transferID)
		logger.Error(err.Error())
		return shim.Error("Unable to retrieve the given request id")
	}

	logger.Debug("temp", string(temp))

	logger.Debug("Exit : getTransferDetailbyTransferID")
	return shim.Success(temp)
}

/*****************************************  MEMBER DETAILS - REWARD POINTS, GET FUNCTIONS **********************************************/

// This method is used by participating orgs. Reward points are requested using this method
func (t *LoyaltyChaincode) setMembershipIdentities(stub shim.ChaincodeStubInterface, memberID string,
	firstName string,
	lastName string,
	phone string) pb.Response {

	logger.Debug("Entry : setMembershipIdentities")
	isAllowed := checkAccessPermissions(stub, "setMembershipIdentities")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	memberReqKey, errComp := stub.CreateCompositeKey("MEMBER", []string{memberID})
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return shim.Error("Unable to set composite key")
	}

	// Retrive the existing member
	temp, _ := stub.GetState(memberReqKey)
	if temp != nil {
		logger.Error("Member ID already exists, create a new one", memberID)
		return shim.Error("Member ID already exists, create a new one")
	}

	// Create new Request
	var MemberInfo MemberDetails
	MemberInfo.MemberID = memberID
	MemberInfo.FirstName = firstName
	MemberInfo.LastName = lastName
	MemberInfo.Phone = phone

	Identifier, _ := cid.GetID(stub)
	MspID, _ := cid.GetMSPID(stub)

	Dept, deptExists, deptErr := cid.GetAttributeValue(stub, "Dept")
	Org, orgExists, orgErr := cid.GetAttributeValue(stub, "Org")
	Role, roleExists, roleErr := cid.GetAttributeValue(stub, "Role")

	if !deptExists || !orgExists || !roleExists {
		logger.Error("Missing attributes in certificate.")
		return shim.Error("Unable to create new member")
	}

	if deptErr != nil || orgErr != nil || roleErr != nil {
		logger.Error("Failed in fetching the certificate details.")
		return shim.Error("Unable to create new member")
	}

	MemberInfo.Org = Org
	MemberInfo.Dept = Dept
	MemberInfo.Role = Role
	MemberInfo.PublicKey = Identifier
	MemberInfo.MspID = MspID

	temp2, err2 := json.Marshal(MemberInfo)
	if err2 != nil {
		logger.Error("Unable to Marshal")
		logger.Error(err2.Error())
		return shim.Error("Unable to Marshal")
	}

	err1 := stub.PutState(memberReqKey, temp2)
	if err1 != nil {
		logger.Error("Unable to create new member", memberID)
		logger.Error(err1.Error())
		return shim.Error("Unable to create new member")
	}

	eventData := fmt.Sprintf("{ MEMBER_ID: %s, PUBLIC_KEY : %s}",
		memberID,
		MemberInfo.PublicKey)

	eventPayload, _ := json.Marshal(eventData)

	stub.SetEvent("MemberRegisterEvent", eventPayload)

	logger.Debug("Exit : setMembershipIdentities")
	return shim.Success([]byte(memberID))
}

// This method is used to get request details by request ID
func (t *LoyaltyChaincode) getMemberDetailsByMemberID(stub shim.ChaincodeStubInterface, memberID string) pb.Response {

	logger.Debug("Entry : getMemberDetailsByMemberID")

	isAllowed := checkAccessPermissions(stub, "getMemberDetailsByMemberID")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	if len(memberID) < 1 {
		// Invalid Request ID
		logger.Error("Request ID is invalid", memberID)
		return shim.Error("Request ID is invalid")
	}

	compoundKey, errComp := stub.CreateCompositeKey("MEMBER", []string{memberID})
	fmt.Println("QueryObject() : Compound Key : ", compoundKey)
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return shim.Error("Unable to set composite key")
	}

	// Retrive the existing ACL from the state
	temp, err := stub.GetState(compoundKey)
	if err != nil {
		logger.Error("Unable to retrieve the given member id", memberID)
		logger.Error(err.Error())
		return shim.Error("Unable to retrieve the given memberID ")
	}

	logger.Debug("Exit : getMemberDetailsByMemberID")
	return shim.Success(temp)
}

func (t *LoyaltyChaincode) getAllRewardRequestDetails(stub shim.ChaincodeStubInterface) pb.Response {

	logger.Debug("Entry : getAllRewardRequestDetails")

	isAllowed := checkAccessPermissions(stub, "getAllRewardRequestDetails")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	resultsIterator, err := stub.GetStateByPartialCompositeKey("ISSUE_REQUEST", []string{})
	if err != nil {
		return shim.Error("Invalid composite key")
	}

	defer resultsIterator.Close()

	var allList []string
	// Iterate through result set
	var i int
	for i = 0; resultsIterator.HasNext(); i++ {

		// Retrieve the Key and Object
		myCompositeKey, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Invalid composite key")
		}

		Detail := string(myCompositeKey.Value)
		fmt.Println(Detail)

		allList = append(allList, Detail)
	}
	fmt.Println(allList)
	allInfo, _ := json.Marshal(allList)

	logger.Debug("Exit : getAllRewardRequestDetails")
	return shim.Success(allInfo)
}

func (t *LoyaltyChaincode) getAllBurnRequestDetails(stub shim.ChaincodeStubInterface) pb.Response {

	logger.Debug("Entry : getAllBurnRequestDetails")

	isAllowed := checkAccessPermissions(stub, "getAllBurnRequestDetails")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	resultsIterator, err := stub.GetStateByPartialCompositeKey("BURN_REQUEST", []string{})
	if err != nil {
		return shim.Error("Invalid composite key")
	}

	defer resultsIterator.Close()

	var allList []string
	// Iterate through result set
	var i int
	for i = 0; resultsIterator.HasNext(); i++ {

		// Retrieve the Key and Object
		myCompositeKey, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Invalid composite key")
		}

		Detail := string(myCompositeKey.Value)
		allList = append(allList, Detail)
	}
	allInfo, _ := json.Marshal(allList)
	logger.Debug("Exit : getAllBurnRequestDetails")
	return shim.Success(allInfo)
}

func (t *LoyaltyChaincode) getAllMemberDetails(stub shim.ChaincodeStubInterface) pb.Response {

	logger.Debug("Entry : getAllMemberDetails")
	isAllowed := checkAccessPermissions(stub, "getAllMemberDetails")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	resultsIterator, err := stub.GetStateByPartialCompositeKey("MEMBER", []string{})
	if err != nil {
		return shim.Error("Invalid composite key")
	}

	defer resultsIterator.Close()
	var allList []string
	// Iterate through result set
	var i int
	for i = 0; resultsIterator.HasNext(); i++ {

		// Retrieve the Key and Object
		myCompositeKey, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Invalid composite key")
		}

		Detail := string(myCompositeKey.Value)
		allList = append(allList, Detail)
	}

	allInfo, _ := json.Marshal(allList)
	logger.Debug("Exit : getAllMemberDetails")
	return shim.Success(allInfo)
}

func (t *LoyaltyChaincode) getAllPurchaseDetails(stub shim.ChaincodeStubInterface) pb.Response {

	logger.Debug("Entry : getAllPurchaseDetails")

	isAllowed := checkAccessPermissions(stub, "getAllPurchaseDetails")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	resultsIterator, err := stub.GetStateByPartialCompositeKey("PURCHASE_REF_ID", []string{})
	if err != nil {
		return shim.Error("Invalid composite key")
	}

	defer resultsIterator.Close()

	var allList []string
	// Iterate through result set
	var i int
	for i = 0; resultsIterator.HasNext(); i++ {

		// Retrieve the Key and Object
		myCompositeKey, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Invalid composite key")
		}

		Detail := string(myCompositeKey.Value)
		allList = append(allList, Detail)
	}

	allInfo, _ := json.Marshal(allList)

	logger.Debug("Exit : getAllPurchaseDetails")
	return shim.Success(allInfo)
}

func (t *LoyaltyChaincode) getAllTransferDetails(stub shim.ChaincodeStubInterface) pb.Response {

	logger.Debug("Entry : getAllTransferDetails")

	isAllowed := checkAccessPermissions(stub, "getAllTransferDetails")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	resultsIterator, err := stub.GetStateByPartialCompositeKey("TRANSFER", []string{})
	if err != nil {
		return shim.Error("Invalid composite key")
	}

	defer resultsIterator.Close()

	var allList []string
	// Iterate through result set
	var i int
	for i = 0; resultsIterator.HasNext(); i++ {

		// Retrieve the Key and Object
		myCompositeKey, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Invalid composite key")
		}

		Detail := string(myCompositeKey.Value)
		allList = append(allList, Detail)
	}

	allInfo, _ := json.Marshal(allList)
	logger.Debug("Exit : getAllTransferDetails")
	return shim.Success(allInfo)
}

func (t *LoyaltyChaincode) getRewardPointsBalanceByMemberID(stub shim.ChaincodeStubInterface, memberID string) pb.Response {

	logger.Debug("Entry : getRewardPointsBalanceByMemberID")

	isAllowed := checkAccessPermissions(stub, "getRewardPointsBalanceByMemberID")

	if !isAllowed {
		logger.Error("Permission denied. Please make sure you have appropriate access.")
		return shim.Error("Permission denied. Please make sure you have appropriate access.")
	}

	if len(memberID) < 1 {
		// Invalid Request ID
		logger.Error("Request ID is invalid", memberID)
		return shim.Error("Request ID is invalid")
	}

	compoundKey, errComp := stub.CreateCompositeKey("MEMBER", []string{memberID})
	fmt.Println("QueryObject() : Compound Key : ", compoundKey)
	if errComp != nil {
		logger.Error("Unable to set composite key")
		logger.Error(errComp.Error())
		return shim.Error("Unable to set composite key")
	}

	temp, err := stub.GetState(compoundKey)
	if err != nil {
		logger.Error("Unable to retrieve the given member id", memberID)
		logger.Error(err.Error())
		return shim.Error("Unable to retrieve the given memberID ")
	}

	var MemberInfo MemberDetails
	json.Unmarshal(temp, &MemberInfo)

	logger.Debug("Exit : getRewardPointsBalanceByMemberID")
	rewardPtStr := strconv.Itoa(MemberInfo.RewardPoints)
	return shim.Success([]byte(rewardPtStr))
}

func main() {

	logger.Debug("Enter : main method")

	err := shim.Start(new(LoyaltyChaincode))
	if err != nil {
		logger.Critical("Failed to start chaincode -", err)
	}
	logger.Debug("Exit : main method")

}
