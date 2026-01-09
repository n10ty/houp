package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("=== Testing UserRegistration ===")

	// Valid registration
	validUser := &UserRegistration{
		Username:        "johndoe",
		Email:           "john@example.com",
		Password:        "securepass123",
		ConfirmPassword: "securepass123",
	}
	if err := validUser.Validate(); err != nil {
		log.Printf("Valid user failed validation: %v\n", err)
	} else {
		fmt.Println("✓ Valid user registration passed")
	}

	// Invalid registration - passwords don't match
	invalidUser := &UserRegistration{
		Username:        "janedoe",
		Email:           "jane@example.com",
		Password:        "securepass123",
		ConfirmPassword: "differentpass",
	}
	if err := invalidUser.Validate(); err != nil {
		fmt.Printf("✓ Invalid user correctly rejected: %v\n", err)
	} else {
		log.Println("Invalid user should have failed validation!")
	}

	fmt.Println("\n=== Testing OrderModificationRequest ===")

	// Valid order modification - CancelOrderId matches OrderId
	cancelId := "ORDER-12345"
	validCancel := &OrderModificationRequest{
		CancelOrderId: &cancelId,
		OrderId:       "ORDER-12345",
		Reason:        "Customer requested cancellation",
	}
	if err := validCancel.Validate(); err != nil {
		log.Printf("Valid cancellation failed validation: %v\n", err)
	} else {
		fmt.Println("✓ Valid order cancellation passed")
	}

	// Invalid order modification - CancelOrderId doesn't match
	wrongCancelId := "ORDER-99999"
	invalidCancel := &OrderModificationRequest{
		CancelOrderId: &wrongCancelId,
		OrderId:       "ORDER-12345",
		Reason:        "Customer requested cancellation",
	}
	if err := invalidCancel.Validate(); err != nil {
		fmt.Printf("✓ Invalid cancellation correctly rejected: %v\n", err)
	} else {
		log.Println("Invalid cancellation should have failed validation!")
	}

	// Valid order modification - CancelOrderId is nil (optional)
	validOrder := &OrderModificationRequest{
		CancelOrderId: nil,
		OrderId:       "ORDER-12345",
		Reason:        "Just updating order details",
	}
	if err := validOrder.Validate(); err != nil {
		log.Printf("Valid order modification failed validation: %v\n", err)
	} else {
		fmt.Println("✓ Valid order modification (no cancellation) passed")
	}
}
