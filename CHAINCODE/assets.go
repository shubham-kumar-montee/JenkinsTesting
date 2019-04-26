package main

// TotalRewardPoints are initialized with 100 Billion - This value will increase over time
var TotalRewardPoints = 100000000000

// MemberDetails - Member information
type MemberDetails struct {
	MemberID     string `json:"MemberID"`
	FirstName    string `json:"FirstName"`
	LastName     string `json:"LastName"`
	Phone        string `json:"Phone"`
	Org          string `json:"Org"`
	Dept         string `json:"Dept"`
	Role         string `json:"Role"`
	PublicKey    string `json:"PublicKey"`
	MspID        string `json:"MspID"`
	RewardPoints int    `json:"RewardPoints"`
}

// Request - Reward Points Request Details
type Request struct {
	RequestID     string                    `json:"RequestID"`
	RequestType   string                    `json:"RequestType"`
	RewardPoints  int                       `json:"RewardPoints"`
	RequestStatus string                    `json:"RequestStatus"`
	RequestedBy   string                    `json:"RequestedBy"`
	RequestedDate string                    `json:"RequestedDate"`
	IssuedDate    string                    `json:"IssuedDate"`
	Identities    map[string]ApproveDetails `json:"Identities"`
}

// ApproveDetails - Approve the request
type ApproveDetails struct {
	ApprovedBy  string `json:"ApprovedBy"`
	ApproveDate string `json:"ApproveDate"`
}

// PurchaseDetails - Consumer spends on Hotel/Airfare/Gas station etc
type PurchaseDetails struct {
	PurchaseID          string `json:"PurchaseID"`
	PurchaseReceiptFile string `json:"PurchaseReceiptFile"`
	PurchaseReceiptHash string `json:"PurchaseReceiptHash"`
	IssuedOrg           string `json:"IssuedOrg"`
	IssuedMember        string `json:"IssuedMember"`
	RewardPtsElig       int    `json:"RewardPtsElig"`
	RewardPtsTrans      int    `json:"RewardPtsTrans"`
	PurchaseBy          string `json:"PurchaseBy"`
	PurchaseDate        string `json:"PurchaseDate"`
	TransferID          string `json:"TransferID"`
}

// TransferDetails - Transfer reward points
type TransferDetails struct {
	TransferID string `json:"TransferID"`
	From       string `json:"From"`
	To         string `json:"To"`
	Value      int    `json:"Value"`
	Remarks    string `json:"Remarks"`
}
