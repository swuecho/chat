package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/swuecho/chat_backend/sqlc_queries"
)

func TestChatSessionService(t *testing.T) {
	sqlc_q := sqlc_queries.New(db)
	service := NewChatSessionService(sqlc_q)
	// Create a new database connection

	// Insert a new chat session into the database
	session_params := sqlc_queries.CreateChatSessionParams{UserID: 1, Topic: "Test Session", MaxLength: 100}
	session, err := service.CreateChatSession(context.Background(), session_params)
	if err != nil {
		t.Fatalf("failed to create chat session: %v", err)
	}

	// Retrieve the chat session by ID and check that it matches the expected values
	retrievedSession, err := service.GetChatSessionByID(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("failed to get chat session: %v", err)
	}
	if retrievedSession.UserID != session.UserID || retrievedSession.Topic != session.Topic || retrievedSession.MaxLength != session.MaxLength {
		t.Error("retrieved chat session does not match expected values")
	}

	// Update the chat session and check that it was updated in the database
	updated_params := sqlc_queries.UpdateChatSessionParams{ID: session.ID,
		UserID: session.UserID,
		Topic:  "Updated Test Session",
	}
	if _, err := service.UpdateChatSession(context.Background(), updated_params); err != nil {
		t.Fatalf("failed to update chat session: %v", err)
	}
	retrievedSession, err = service.GetChatSessionByID(context.Background(), session.ID)
	if err != nil {
		t.Fatalf("failed to get chat session: %v", err)
	}

	// Check that updated chat session matches expected values
	// TODO: timezone time || !retrievedSession.UpdatedAt.Equal(updated_params.UpdatedAt)
	if retrievedSession.Topic != updated_params.Topic {
		t.Errorf("chat session mismatch: expected Topic=%s,  got Topic=%s ",
			updated_params.Topic, retrievedSession.Topic)
	}

	// Delete the chat session and check that it was deleted from the database

	if err := service.DeleteChatSession(context.Background(), session.ID); err != nil {
		t.Fatalf("failed to delete chat session: %v", err)
	}
	deletedSession, err := service.GetChatSessionByID(context.Background(), session.ID)
	if err == nil || !errors.Is(err, sql.ErrNoRows) {
		fmt.Printf("%+v", deletedSession)
		t.Error("expected error due to missing chat session, but got no error or different error")
	}
}

func TestGetChatSessionsByUserID(t *testing.T) {
	sqlc_q := sqlc_queries.New(db)
	service := NewChatSessionService(sqlc_q)

	// Insert two chat sessions into the database with different user IDs
	session1_params := sqlc_queries.CreateChatSessionParams{UserID: 1, Topic: "Test Session 1", MaxLength: 100, Uuid: "uuid1"}
	session1, err := service.CreateChatSession(context.Background(), session1_params)
	if err != nil {
		t.Fatalf("failed to create chat session: %v", err)
	}
	session2_params := sqlc_queries.CreateChatSessionParams{UserID: 2, Topic: "Test Session 2", MaxLength: 150, Uuid: "uuid2"}
	session2, err := service.CreateChatSession(context.Background(), session2_params)
	if err != nil {
		t.Fatalf("failed to create chat session: %v", err)
	}

	// Retrieve chat sessions by user ID and check that they match the expected values
	userID := int32(1)
	sessions, err := service.GetChatSessionsByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to retrieve chat sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Errorf("expected 1 chat session, but got %d", len(sessions))
	}
	if sessions[0].UserID != session1.UserID || sessions[0].Topic != session1.Topic || sessions[0].MaxLength != session1.MaxLength {
		t.Error("retrieved chat sessions do not match expected values")
	}
	if err := service.DeleteChatSession(context.Background(), session1.ID); err != nil {
		t.Fatalf("failed to delete chat session: %v", err)
	}
	if err := service.DeleteChatSession(context.Background(), session2.ID); err != nil {
		t.Fatalf("failed to delete chat session: %v", err)
	}
}

func TestGetAllChatSessions(t *testing.T) {

	// Create a new ChatSessionService with the test database connection
	q := sqlc_queries.New(db)
	service := NewChatSessionService(q)

	session1_params := sqlc_queries.CreateChatSessionParams{UserID: 1, Topic: "Test Session 1", MaxLength: 100, Uuid: "uuid1"}
	session1, err := service.CreateChatSession(context.Background(), session1_params)
	if err != nil {
		t.Fatalf("failed to create chat session: %v", err)
	}
	session2_params := sqlc_queries.CreateChatSessionParams{UserID: 2, Topic: "Test Session 2", MaxLength: 150, Uuid: "uuid2"}
	session2, err := service.CreateChatSession(context.Background(), session2_params)
	if err != nil {
		t.Fatalf("failed to create chat session: %v", err)
	}

	// Retrieve all chat sessions and check that they match the expected values
	sessions, err := service.GetAllChatSessions(context.Background())
	if err != nil {
		t.Fatalf("failed to retrieve chat sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 chat sessions, but got %d", len(sessions))
	}
	if sessions[0].Topic != session1.Topic || sessions[1].Topic != session2.Topic {
		t.Error("retrieved chat sessions do not match expected values")
	}

	if err := service.DeleteChatSession(context.Background(), session1.ID); err != nil {
		t.Fatalf("failed to delete chat session: %v", err)
	}
	if err := service.DeleteChatSession(context.Background(), session2.ID); err != nil {
		t.Fatalf("failed to delete chat session: %v", err)
	}
}
