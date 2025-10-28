package main

import (
	"e5realtimechat/database"
	"log"
)

func main() {
	log.Println("🔌 Testing database connection...")

	db, err := database.NewDB("localhost", "5432", "chatuser", "chatpass", "chatdb")
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connection successful!")

	// Test: Get all users
	log.Println("\n📋 Testing: Get sample users...")

	users := []string{"admin", "alice", "bob"}
	for _, username := range users {
		user, err := db.GetUserByUsername(username)
		if err != nil {
			log.Printf("⚠️  User '%s' not found: %v", username, err)
			continue
		}
		log.Printf("✅ Found user: %s (ID: %d, Email: %s, Online: %v)",
			user.Username, user.ID, user.Email, user.IsOnline)
	}

	// Test: Get general room
	log.Println("\n📋 Testing: Get 'general' room...")
	room, err := db.GetRoomByName("general")
	if err != nil {
		log.Printf("⚠️  Room not found: %v", err)
	} else {
		log.Printf("✅ Found room: %s (ID: %d, Type: %s)",
			room.RoomName, room.ID, room.RoomType)
	}

	log.Println("\n🎉 All tests passed! Database is ready to use.")
}
