package kong

import (
	"context"
	"errors"
	"fmt"
	"log"
)

// ConsumerService handles Consumers in Kong.
type ConsumerService service

// Create creates a Consumer in Kong.
// If an ID is specified, it will be used to
// create a consumer in Kong, otherwise an ID
// is auto-generated.
func (s *ConsumerService) Create(ctx context.Context, consumer *Consumer) (*Consumer, error) {

	queryPath := "/consumers"
	method := "POST"
	if consumer.ID != nil {
		queryPath = queryPath + "/" + *consumer.ID
		method = "PUT"
	}
	req, err := s.client.newRequest(method, queryPath, nil, consumer)

	if err != nil {
		return nil, err
	}

	var createdConsumer Consumer
	_, err = s.client.Do(ctx, req, &createdConsumer)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &createdConsumer, nil
}

// Get fetches a Consumer in Kong.
func (s *ConsumerService) Get(ctx context.Context, usernameOrID *string) (*Consumer, error) {

	if usernameOrID == nil {
		return nil, errors.New("usernameOrID cannot be nil for Get operation")
	}

	endpoint := fmt.Sprintf("/consumers/%v", *usernameOrID)
	req, err := s.client.newRequest("GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var consumer Consumer
	_, err = s.client.Do(ctx, req, &consumer)
	if err != nil {
		return nil, err
	}
	return &consumer, nil
}

// Update updates a Consumer in Kong
func (s *ConsumerService) Update(ctx context.Context, consumer *Consumer) (*Consumer, error) {

	if consumer.ID == nil {
		return nil, errors.New("ID cannot be nil for Update operation")
	}

	endpoint := fmt.Sprintf("/consumers/%v", *consumer.ID)
	req, err := s.client.newRequest("PATCH", endpoint, nil, consumer)
	if err != nil {
		return nil, err
	}

	var updatedAPI Consumer
	_, err = s.client.Do(ctx, req, &updatedAPI)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &updatedAPI, nil
}

// Delete deletes a Consumer in Kong
func (s *ConsumerService) Delete(ctx context.Context, usernameOrID *string) error {

	if usernameOrID == nil {
		return errors.New("usernameOrID cannot be nil for Delete operation")
	}

	endpoint := fmt.Sprintf("/consumers/%v", *usernameOrID)
	req, err := s.client.newRequest("DELETE", endpoint, nil, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(ctx, req, nil)
	return err
}
