package main

import (
	"fmt"
	"testing"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func checkInit(t *testing.T, stub *shim.MockStub, args [][]byte) {
	fmt.Println("Entering TestInit")
	res := stub.MockInit("init", args)
	if res.Status != shim.OK {
		fmt.Println("Init failed", string(res.Message))
		t.FailNow()
	}
}

func checkStateValue(t *testing.T, stub *shim.MockStub, name string, value string) {
	bytes := stub.State[name]
	if bytes == nil {
		fmt.Println("State", name, "failed to get value")
		t.FailNow()
	}
	if string(bytes) != value {
		fmt.Println("State value", name, "was not", value, "as expected")
		t.FailNow()
	}
}

func checkQuery(t *testing.T, stub *shim.MockStub, name string) {
	res := stub.MockInvoke("1", [][]byte{[]byte("query"), []byte(name)})
	if res.Status != shim.OK {
		fmt.Println("Query", name, "failed", string(res.Message))
		t.FailNow()
	}
	// if res.Payload == nil {
	// 	fmt.Println("Query", name, "failed to get value")
	// 	t.FailNow()
	// }
	// if string(res.Payload) != value {
	// 	fmt.Println("Query value", name, "was not", value, "as expected")
	// 	t.FailNow()
	// }
}

func checkInvoke(t *testing.T, stub *shim.MockStub, args [][]byte) {
	res := stub.MockInvoke("1", args)
	if res.Status != shim.OK {
		fmt.Println("Invoke", args, "failed", string(res.Message))
		t.FailNow()
	}
}

func checkState(t *testing.T, stub *shim.MockStub, name string) {
	bytes := stub.State[name]
	if bytes == nil {
		fmt.Println("State", name, "failed to get value")
		t.FailNow()
	}
}

func TestLoyaltyChainCode_Init(t *testing.T) {

	stub := shim.NewMockStub("mockStub", new(LoyaltyChaincode))

	// Init with ACL
	jsonCond := `[{ "FunctionName" : "getACLConditions" , 
					"ConditionsList": [ {"Org":"HILTON","OU":"Bangalore","Role":"ADMIN"}, 
										{"Org":"ACCOR", "OU":"Bangalore","Role":"ADMIN"}, 
										{"Org":"LUFTHANSA", "OU":"Bangalore","Role":"ADMIN"}] } , 
				{ "FunctionName" : "requestRewardPoints" , 
					"ConditionsList": [ {"Org":"HILTON", "OU":"Bangalore","Role":"ADMIN"}, 
										{"Org":"ACCOR", "OU":"Bangalore","Role":"ADMIN"}]} ]`

	checkInit(t, stub, [][]byte{[]byte(jsonCond)})
	fmt.Println()
	fmt.Println("1.Init - Passed")

	checkStateValue(t, stub, "TOTAL_REWARD_POINTS", "100000000000")
	fmt.Println()
	fmt.Println("2.GetState TOTAL_REWARD_POINTS - Passed")

	fmt.Println()
	fmt.Println("3.GetState ACCESS_CONTROL_LIST - Passed")
	checkState(t, stub, "ACCESS_CONTROL_LIST")

	fmt.Println()
	fmt.Println("4.Request Reward Points - Passed")
	checkInvoke(t, stub, [][]byte{[]byte("requestRewardPoints"), []byte("123")})

	fmt.Println()
	fmt.Println("5.Get Request Detail By ID")
	checkInvoke(t, stub, [][]byte{[]byte("getRequestDetailByRequestID"), []byte("REQUEST_1"), []byte("ISSUE_REQ")})

	fmt.Println()
	fmt.Println("6.Request Reward Points - Passed")
	checkInvoke(t, stub, [][]byte{[]byte("requestRewardPoints"), []byte("245"), []byte("ISSUE_REQ")})

	fmt.Println()
	fmt.Println("7.Request Burn Points - Passed")
	checkInvoke(t, stub, [][]byte{[]byte("burnRewardPoints"), []byte("245"), []byte("BURN_REQ")})

	fmt.Println()
	fmt.Println("8.Get Request Detail By ID")
	checkInvoke(t, stub, [][]byte{[]byte("getRequestDetailByRequestID"), []byte("REQUEST_2"), []byte("ISSUE_REQ")})

	fmt.Println()
	fmt.Println("9.Get Request Detail By ID")
	checkInvoke(t, stub, [][]byte{[]byte("getRequestDetailByRequestID"), []byte("REQUEST_3"), []byte("BURN_REQ")})

	fmt.Println()
	fmt.Println("10.Get Request Detail By ID")
	checkInvoke(t, stub, [][]byte{[]byte("getRequestDetailByRequestID"), []byte("REQUEST_2"), []byte("BURN_REQ")})
	checkInvoke(t, stub, [][]byte{[]byte("getAllRewardRequestDetails")})

	fmt.Println()
	fmt.Println("11.Get Request Detail By ID")
	checkInvoke(t, stub, [][]byte{[]byte("approveRequest"), []byte("REQUEST_1"), []byte("KAVIN"), []byte("ISSUE_REQ")})

	fmt.Println()
	fmt.Println("5.Get Request Detail By ID")
	checkInvoke(t, stub, [][]byte{[]byte("getRequestDetailByRequestID"), []byte("REQUEST_1"), []byte("ISSUE_REQ")})

	fmt.Println()
	checkInvoke(t, stub, [][]byte{[]byte("setMembershipIdentities"), []byte("SUNIL_KTG"),
		[]byte("SUNIL"),
		[]byte("GUNA"),
		[]byte("7094792597"),
		[]byte("STARWOOD"),
		[]byte("HOTEL"),
		[]byte("ADMIN"),
		[]byte("0x1234"),
	})

	fmt.Println()
	checkInvoke(t, stub, [][]byte{[]byte("setMembershipIdentities"), []byte("SUNIL_KTG1"),
		[]byte("SUNIL"),
		[]byte("GUNA"),
		[]byte("7094792597"),
		[]byte("STARWOOD"),
		[]byte("HOTEL"),
		[]byte("ADMIN"),
		[]byte("0x1234"),
	})

	logger.Debug("Exit : ")

	fmt.Println()
	fmt.Println("5.Get Request Detail By ID")
	checkInvoke(t, stub, [][]byte{[]byte("getMemberDetailsByMemberID"), []byte("MEMBER"), []byte("SUNIL_KTG1")})

	fmt.Println()
	fmt.Println()
	checkInvoke(t, stub, [][]byte{[]byte("getAllMemberDetails")})

	fmt.Println()
	fmt.Println()
	checkInvoke(t, stub, [][]byte{[]byte("getRewardPoints")})

	fmt.Println()
	fmt.Println()
	checkInvoke(t, stub, [][]byte{[]byte("getAllACLConditions")})

	fmt.Println()
	fmt.Println()
	checkInvoke(t, stub, [][]byte{[]byte("getACLConditionsByFuncAndOrg"), []byte("getAllACLConditions"), []byte("HILTON")})

	fmt.Println()
	fmt.Println("5.Get Request Detail By ID")
	checkInvoke(t, stub, [][]byte{[]byte("updatePurchase"), []byte("PURCHASE_1"), []byte("100"), []byte("HILTON")})

	fmt.Println()
	fmt.Println("5.Get Request Detail By ID")
	checkInvoke(t, stub, [][]byte{[]byte("getTransferDetailbyTransferID"), []byte("9")})

}
