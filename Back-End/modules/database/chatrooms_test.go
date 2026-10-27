package database

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestChatRoomsKeepInviteOnlyRoomsPrivateAndMessagesMemberScoped(t *testing.T) {
	suffix := fmt.Sprintf("chat-test-%d", time.Now().UnixNano())
	ownerID := createLibraryTestUser(t, "chat-owner-"+suffix+"@example.test")
	visitorID := createLibraryTestUser(t, "chat-visitor-"+suffix+"@example.test")
	t.Cleanup(func() {
		if _, err := DB.Exec("DELETE FROM users WHERE id IN ($1, $2)", ownerID, visitorID); err != nil {
			t.Errorf("clean up chat users: %v", err)
		}
	})

	publicRoom, err := CreateChatRoom(ownerID, "Public room "+suffix, ChatRoomPublic)
	if err != nil || !publicRoom.IsOwner || !publicRoom.IsMember {
		t.Fatalf("create public room = (%#v, %v)", publicRoom, err)
	}
	visitorRooms, err := ListChatRoomsForUser(visitorID)
	publicRoomSummary, found := findChatRoom(visitorRooms, publicRoom.ID)
	if err != nil || !found || publicRoomSummary.IsMember {
		t.Fatalf("visitor public rooms = (%#v, %v), want discoverable non-member public room", visitorRooms, err)
	}
	if _, err := GetChatRoomForMember(publicRoom.ID, visitorID); !errors.Is(err, ErrChatRoomAccess) {
		t.Fatalf("visitor should not read public room before joining, got %v", err)
	}
	joined, err := JoinPublicChatRoom(publicRoom.ID, visitorID)
	if err != nil || !joined {
		t.Fatalf("join public room = (%t, %v), want (true, nil)", joined, err)
	}
	message, err := CreateChatMessage(publicRoom.ID, visitorID, "Hello public room", "/images/chat/messages/public.png")
	if err != nil || message.Body != "Hello public room" || message.ImagePath != "/images/chat/messages/public.png" {
		t.Fatalf("create public message = (%#v, %v)", message, err)
	}
	messages, err := GetChatMessages(publicRoom.ID, ownerID, 10)
	if err != nil || len(messages) != 1 || messages[0].Body != "Hello public room" {
		t.Fatalf("owner messages = (%#v, %v)", messages, err)
	}

	privateRoom, err := CreateChatRoom(ownerID, "Invite room "+suffix, ChatRoomInviteOnly)
	if err != nil {
		t.Fatalf("create invite room: %v", err)
	}
	visitorRooms, err = ListChatRoomsForUser(visitorID)
	if err != nil || hasChatRoom(visitorRooms, privateRoom.ID) {
		t.Fatalf("invite-only room must not be discoverable: (%#v, %v)", visitorRooms, err)
	}
	if _, err := JoinPublicChatRoom(privateRoom.ID, visitorID); !errors.Is(err, ErrChatInviteRequired) {
		t.Fatalf("join invite-only room without link = %v, want invite required", err)
	}
	if err := CreateChatRoomInvite(privateRoom.ID, ownerID, "hash-"+suffix); err != nil {
		t.Fatalf("create owner invite: %v", err)
	}
	roomID, joined, err := JoinChatRoomWithInvite("hash-"+suffix, visitorID)
	if err != nil || !joined || roomID != privateRoom.ID {
		t.Fatalf("join invite-only room = (%d, %t, %v)", roomID, joined, err)
	}
	if _, err := CreateChatMessage(privateRoom.ID, visitorID, "Hello invite room", ""); err != nil {
		t.Fatalf("visitor message after invite: %v", err)
	}
}

func TestChatRoomOwnerCanChangeVisibilityWithoutRemovingMembers(t *testing.T) {
	suffix := fmt.Sprintf("chat-visibility-test-%d", time.Now().UnixNano())
	ownerID := createLibraryTestUser(t, "chat-visibility-owner-"+suffix+"@example.test")
	memberID := createLibraryTestUser(t, "chat-visibility-member-"+suffix+"@example.test")
	visitorID := createLibraryTestUser(t, "chat-visibility-visitor-"+suffix+"@example.test")
	t.Cleanup(func() {
		if _, err := DB.Exec("DELETE FROM users WHERE id IN ($1, $2, $3)", ownerID, memberID, visitorID); err != nil {
			t.Errorf("clean up visibility users: %v", err)
		}
	})

	room, err := CreateChatRoom(ownerID, "Visibility room "+suffix, ChatRoomPublic)
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if _, err := JoinPublicChatRoom(room.ID, memberID); err != nil {
		t.Fatalf("join original public room: %v", err)
	}
	if err := UpdateChatRoomVisibility(room.ID, ownerID, ChatRoomInviteOnly); err != nil {
		t.Fatalf("make room invite-only: %v", err)
	}
	if _, err := GetChatRoomForMember(room.ID, memberID); err != nil {
		t.Fatalf("existing member lost access after privacy change: %v", err)
	}
	visitorRooms, err := ListChatRoomsForUser(visitorID)
	if err != nil || hasChatRoom(visitorRooms, room.ID) {
		t.Fatalf("visitor rooms after invite-only change = (%#v, %v), want empty", visitorRooms, err)
	}
	if _, err := JoinPublicChatRoom(room.ID, visitorID); !errors.Is(err, ErrChatInviteRequired) {
		t.Fatalf("visitor joined invite-only room without invite: %v", err)
	}
	if err := UpdateChatRoomVisibility(room.ID, memberID, ChatRoomPublic); !errors.Is(err, ErrChatRoomAccess) {
		t.Fatalf("non-owner visibility update = %v, want access denied", err)
	}
	if err := UpdateChatRoomVisibility(room.ID, ownerID, ChatRoomPublic); err != nil {
		t.Fatalf("make room public: %v", err)
	}
	if err := UpdateChatRoomTheme(room.ID, ownerID, DisplayThemeForest, "#387c5d"); err != nil {
		t.Fatalf("set room theme: %v", err)
	}
	updatedRoom, err := GetChatRoomForMember(room.ID, ownerID)
	if err != nil || updatedRoom.Theme != DisplayThemeForest || updatedRoom.CustomThemeColor != "#387c5d" {
		t.Fatalf("get saved room theme = (%#v, %v)", updatedRoom, err)
	}
	if err := UpdateChatRoomTheme(room.ID, memberID, DisplayThemePlum, ""); !errors.Is(err, ErrChatRoomAccess) {
		t.Fatalf("non-owner theme update = %v, want access denied", err)
	}
	visitorRooms, err = ListChatRoomsForUser(visitorID)
	if err != nil || !hasChatRoom(visitorRooms, room.ID) {
		t.Fatalf("visitor rooms after public change = (%#v, %v)", visitorRooms, err)
	}
}

func TestChatRoomImagesAreStoredAndOwnerScoped(t *testing.T) {
	suffix := fmt.Sprintf("chat-image-test-%d", time.Now().UnixNano())
	ownerID := createLibraryTestUser(t, "chat-image-owner-"+suffix+"@example.test")
	visitorID := createLibraryTestUser(t, "chat-image-visitor-"+suffix+"@example.test")
	t.Cleanup(func() {
		if _, err := DB.Exec("DELETE FROM users WHERE id IN ($1, $2)", ownerID, visitorID); err != nil {
			t.Errorf("clean up image users: %v", err)
		}
	})

	room, err := CreateChatRoom(ownerID, "Image room "+suffix, ChatRoomPublic)
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if err := UpdateChatRoomImagePath(room.ID, ownerID, "/images/chat/rooms/test.png"); err != nil {
		t.Fatalf("set room image: %v", err)
	}
	updated, err := GetChatRoomForMember(room.ID, ownerID)
	if err != nil || updated.ImagePath != "/images/chat/rooms/test.png" {
		t.Fatalf("get stored room image = (%#v, %v)", updated, err)
	}
	if err := UpdateChatRoomImagePath(room.ID, visitorID, "/images/chat/rooms/not-allowed.png"); !errors.Is(err, ErrChatRoomAccess) {
		t.Fatalf("non-owner room image update = %v, want access denied", err)
	}
}

func hasChatRoom(rooms []ChatRoomSummary, roomID int64) bool {
	_, found := findChatRoom(rooms, roomID)
	return found
}

func findChatRoom(rooms []ChatRoomSummary, roomID int64) (ChatRoomSummary, bool) {
	for _, room := range rooms {
		if room.ID == roomID {
			return room, true
		}
	}
	return ChatRoomSummary{}, false
}
