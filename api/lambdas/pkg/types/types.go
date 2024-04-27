package types

import (
    "fmt"
)

type AccountStatus int

const (
	Pending AccountStatus = iota
    CreatingAccount
	Created
	Failed
)

func (e AccountStatus) String() string {
	switch e {
	case Pending:
		return "Pending"
	case Created:
		return "Created"
	case Failed:
		return "Failed"
    case CreatingAccount:
        return "CreatingAccount"
	default:
		panic(fmt.Errorf("unknown AccountStatus: %d", e))
	}
}

type Organization struct {
	OrgName                  string `json:"orgName"`
	PulumiAccessToken        string `json:"pulumiAccessToken"`
	OrgManagementEnvironment string `json:"orgManagementEnvironment"`
}

type Account struct {
	AccountName     string        `json:"accountName"`
	Email           string        `json:"email"`
	ParentID        string        `json:"parentID"`
    // TODO: The AWS creds shouldn't be passed in with the request but rather retrieved from ESC or some other short lived credential service
	AwsAccessKey    string        `json:"awsAccessKey"`
	AwsSecretKey    string        `json:"awsSecretKey"`
	AwsSessionToken string        `json:"awsSessionToken"`
	Status          AccountStatus `json:"status"`
}
