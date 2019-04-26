package main

// Type of request
const (
	IssueRequest = "ISSUE_REQUEST"
	BurnRequest  = "BURN_REQUEST"
)

// Reward points Request Status
const (
	Issued          = "ISSUED"
	Requested       = "REQUESTED"
	PendingApproval = "PENDING_APPROVAL"
	SettledUp       = "SETTLED_UP"
)

// Approval status from each entity
const (
	ApprovedBy     = "APPROVED_BY"
	ApprovedDate   = "APPROVED_DATE"
	Comment        = "COMMENT"
	ApprovalStatus = "APPROVAL_STATUS"
)

// Approval status
const (
	Rejected = "REJECTED"
	Accepted = "ACCEPTED"
)

// Used for Event handling
const (
	InitializeEvent       = "LOYALTY_PGM_INIT_EVENT"
	MemberRegisterEvent   = "MEMBER_REGISTER_EVENT"
	RequestRewardPtsEvent = "REQUEST_REWARD_PTS_EVENT"
	BurnRewardPtsEvent    = "BURN_REWARD_PTS_EVENT"
	PurchaseEvent         = "PURCHASE_EVENT"
	ApproveEvent          = "APPROVE_EVENT"
	TransferEvent         = "TRANSFER_REWARD_PTS_EVENT"
)
