package chaincode

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Representation struct {
	ID             string `json:"ID"`
	ClientUsername string `json:"user_id"`
	AgentUsername  string `json:"agent_id"`
	StartDate      string `json:"start_date"`   // ISO 8601 format for timestamps
	EndDate        string `json:"end_date"`     // ISO 8601 format for timestamps
	Status         string `json:"status"`       // AgreementStatus enum-like string
	RequestedAt    string `json:"requested_at"` // ISO 8601 format for timestamps
	SignedAt       string `json:"signed_at"`    // ISO 8601 format for timestamps
	IsActive       bool   `json:"is_active"`
}

type Property struct {
	ID      string         `json:"ID"`
	Address string         `json:"Address"`
	Owner   string         `json:"Owner"`
	Agent   string         `json:"Agent"`
	State   string         `json:"State"`
	Bids    map[string]Bid `json:"Bids"`
}

type Bid struct {
	ID              string `json:"ID"`
	Amount          int    `json:"Amount"`
	Bidder          string `json:"Bidder"`
	Agent           string `json:"Agent"`
	BuyerCountered  bool   `json:"BuyerCountered"`
	SellerCountered bool   `json:"SellerCountered"`
}

func (s *SmartContract) RepresentationExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	representationJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", id)
	}

	return representationJSON != nil, nil
}

func (s *SmartContract) RequestRepresentation(ctx contractapi.TransactionContextInterface, representationID string, clientUsername string, agentUsername string, startDate string, endDate string) error {
	exists, err := s.RepresentationExists(ctx, representationID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("a representation request with ID %s already exists", representationID)
	}

	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}

	representation := Representation{
		ID:             representationID,
		ClientUsername: clientUsername,
		AgentUsername:  agentUsername,
		StartDate:      startDate,
		EndDate:        endDate,
		Status:         "pending",
		RequestedAt:    txTimestamp.String(),
		IsActive:       false,
	}

	representationJSON, err := json.Marshal(representation)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(representationID, representationJSON)
}

func (s *SmartContract) DeclineRepresentation(ctx contractapi.TransactionContextInterface, representationID string) error {
	representationJSON, err := ctx.GetStub().GetState(representationID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if representationJSON == nil {
		return fmt.Errorf("representation %s does not exist", representationID)
	}

	var representation Representation
	err = json.Unmarshal(representationJSON, &representation)
	if err != nil {
		return fmt.Errorf("failed to unmarshal representation: %v", err)
	}

	representation.Status = "declined"
	representation.IsActive = false

	updatedRepresentationJSON, err := json.Marshal(representation)
	if err != nil {
		return fmt.Errorf("failed to marshal updated representation: %v", err)
	}

	return ctx.GetStub().PutState(representationID, updatedRepresentationJSON)
}

func (s *SmartContract) AcceptRepresentation(ctx contractapi.TransactionContextInterface, representationID string) error {
	representationJSON, err := ctx.GetStub().GetState(representationID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if representationJSON == nil {
		return fmt.Errorf("representation %s does not exist", representationID)
	}

	var representation Representation
	err = json.Unmarshal(representationJSON, &representation)
	if err != nil {
		return fmt.Errorf("failed to unmarshal representation: %v", err)
	}

	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return fmt.Errorf("failed to get transaction timestamp: %v", err)
	}

	representation.Status = "accepted"
	representation.SignedAt = txTimestamp.String()
	representation.IsActive = true

	updatedRepresentationJSON, err := json.Marshal(representation)
	if err != nil {
		return fmt.Errorf("failed to marshal updated representation: %v", err)
	}

	return ctx.GetStub().PutState(representationID, updatedRepresentationJSON)
}

func (s *SmartContract) ListRepresentations(ctx contractapi.TransactionContextInterface, username string) ([]*Representation, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf("failed to get state by range: %v", err)
	}
	defer resultsIterator.Close()

	var representations []*Representation
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to iterate over results: %v", err)
		}

		var representation Representation
		err = json.Unmarshal(queryResponse.Value, &representation)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal representation: %v", err)
		}

		// Check if the username matches either ClientUsername or AgentUsername
		if representation.ClientUsername == username || representation.AgentUsername == username {
			representations = append(representations, &representation)
		}
	}

	return representations, nil
}

func (s *SmartContract) GetRepresentation(ctx contractapi.TransactionContextInterface, representationID string) (*Representation, error) {
	representationJSON, err := ctx.GetStub().GetState(representationID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if representationJSON == nil {
		return nil, fmt.Errorf("representation %s does not exist", representationID)
	}

	var representation Representation
	err = json.Unmarshal(representationJSON, &representation)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal representation: %v", err)
	}

	return &representation, nil
}

func (s *SmartContract) PropertyExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	propertyJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", id)
	}

	return propertyJSON != nil, nil
}

func (s *SmartContract) RegisterProperty(ctx contractapi.TransactionContextInterface, id string, address string, owner string, agent string) error {
	exists, err := s.PropertyExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the property %s already exists", id)
	}

	property := Property{
		ID:      id,
		Address: address,
		Owner:   owner,
		Agent:   agent,
		State:   "Not for sale",
		Bids:    make(map[string]Bid),
	}
	propertyJSON, err := json.Marshal(property)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, propertyJSON)
}

func (s *SmartContract) ListProperty(ctx contractapi.TransactionContextInterface, id string) error {
	//*insert access control here*

	propertyJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed read from world state: %v", id)
	}
	if propertyJSON == nil {
		return fmt.Errorf("the property %s does not exist", id)
	}

	var property Property

	err = json.Unmarshal(propertyJSON, &property)
	if err != nil {
		return err
	}
	property.State = "Listed"
	propertyJSON, err = json.Marshal(property)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, propertyJSON)

}

func (s *SmartContract) ViewProperties(ctx contractapi.TransactionContextInterface) ([]*Property, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var properties []*Property
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var property Property
		err = json.Unmarshal(queryResponse.Value, &property)
		if err != nil {
			return nil, err
		}
		properties = append(properties, &property)
	}

	return properties, nil
}

func (s *SmartContract) PlaceBid(ctx contractapi.TransactionContextInterface, propertyId string, id string, amount int, bidder string, agent string) error {
	propertyJSON, err := ctx.GetStub().GetState(propertyId)
	if err != nil {
		return fmt.Errorf("failed read from world state: %v", propertyId)
	}
	if propertyJSON == nil {
		return fmt.Errorf("the property %s does not exist", propertyId)
	}

	var property Property
	err = json.Unmarshal(propertyJSON, &property)
	if err != nil {
		return err
	}

	if _, exists := property.Bids[id]; exists {
		return fmt.Errorf("the bid %s already exists", id)
	}

	bid := Bid{
		ID:              id,
		Amount:          amount,
		Bidder:          bidder,
		Agent:           agent,
		BuyerCountered:  false,
		SellerCountered: false,
	}
	property.Bids[id] = bid

	propertyJSON, err = json.Marshal(property)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(propertyId, propertyJSON)
}

func (s *SmartContract) RejectBid(ctx contractapi.TransactionContextInterface, propertyId string, id string) error {
	// TODO: Add access control, make sure client is the property owner
	propertyJSON, err := ctx.GetStub().GetState(propertyId)
	if err != nil {
		return fmt.Errorf("failed to get world state: %v", propertyId)
	}
	if propertyJSON == nil {
		return fmt.Errorf("property %v does not exist", propertyId)
	}

	var property Property
	err = json.Unmarshal(propertyJSON, &property)
	if err != nil {
		return err
	}

	if _, exists := property.Bids[id]; !exists {
		return fmt.Errorf("bid %s does not exist", id)
	}
	delete(property.Bids, id)

	propertyJSON, err = json.Marshal(property)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(propertyId, propertyJSON)

}

func (s *SmartContract) CounterBid(ctx contractapi.TransactionContextInterface, propertyId string, id string, amount int) error {

	// Logic is a bit redundant, maybe put into another func
	propertyJSON, err := ctx.GetStub().GetState(propertyId)
	if err != nil {
		return fmt.Errorf("failed to get world state: %v", propertyId)
	}
	if propertyJSON == nil {
		return fmt.Errorf("property %v does not exist", propertyId)
	}

	var property Property
	err = json.Unmarshal(propertyJSON, &property)
	if err != nil {
		return err
	}

	bid, exists := property.Bids[id]
	if !exists {
		return fmt.Errorf("bid %s does not exist", id)
	}
	bid.Amount = amount
	property.Bids[id] = bid

	//check if client is Buyer or Seller and set flags accordingly

	propertyJSON, err = json.Marshal(property)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(propertyId, propertyJSON)
}
