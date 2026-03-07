package ws_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/E-Timileyin/skill-island/services/api/internal/ws"
	"github.com/gorilla/websocket"
)

// ---------- helpers ----------

func startTestHub() *ws.Hub {
	hub := ws.NewHub(nil)
	go hub.Run()
	return hub
}

func dialTestServer(t *testing.T, hub *ws.Hub, profileID string) *websocket.Conn {
	t.Helper()
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("upgrade: %v", err)
		}
		client := ws.NewClient(hub, conn, profileID)
		hub.Register(client)
		go client.WritePump()
		go client.ReadPump()
	}))
	t.Cleanup(srv.Close)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func readMsg(t *testing.T, conn *websocket.Conn, v interface{}) {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("readMsg: %v", err)
	}
	if err := json.Unmarshal(msg, v); err != nil {
		t.Fatalf("readMsg unmarshal: %v (raw: %s)", err, msg)
	}
}

func sendMsg(t *testing.T, conn *websocket.Conn, v interface{}) {
	t.Helper()
	data, _ := json.Marshal(v)
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		t.Fatalf("sendMsg: %v", err)
	}
}

// ---------- tests ----------

func TestHub_RegisterAndCount(t *testing.T) {
	hub := startTestHub()
	_ = dialTestServer(t, hub, "profile-1")
	time.Sleep(50 * time.Millisecond)
	if got := hub.ClientCount(); got != 1 {
		t.Fatalf("expected 1 client, got %d", got)
	}
}

func TestHub_JoinRoom_CreatesWaitingRoom(t *testing.T) {
	hub := startTestHub()
	conn := dialTestServer(t, hub, "p1")
	time.Sleep(50 * time.Millisecond)
	sendMsg(t, conn, ws.JoinRoom{Type: "join_room", GameType: "team_tower"})
	time.Sleep(100 * time.Millisecond)
	if got := hub.RoomCount(); got != 1 {
		t.Fatalf("expected 1 room, got %d", got)
	}
}

func TestHub_TwoPlayers_RoomReady(t *testing.T) {
	hub := startTestHub()
	conn1 := dialTestServer(t, hub, "p1")
	conn2 := dialTestServer(t, hub, "p2")
	time.Sleep(50 * time.Millisecond)

	sendMsg(t, conn1, ws.JoinRoom{Type: "join_room", GameType: "team_tower"})
	time.Sleep(100 * time.Millisecond)
	sendMsg(t, conn2, ws.JoinRoom{Type: "join_room", GameType: "team_tower"})

	var wMsg map[string]interface{}
	readMsg(t, conn1, &wMsg)
	if wMsg["type"] != "waiting_for_partner" {
		t.Fatalf("expected waiting_for_partner, got %v", wMsg["type"])
	}

	var msg1 map[string]interface{}
	readMsg(t, conn1, &msg1)
	if msg1["type"] != "room_ready" {
		t.Fatalf("expected room_ready for p1, got %v", msg1["type"])
	}
	if msg1["player_role"] != "player_1" {
		t.Fatalf("expected player_1 role, got %v", msg1["player_role"])
	}

	var msg2 map[string]interface{}
	readMsg(t, conn2, &msg2)
	if msg2["type"] != "room_ready" {
		t.Fatalf("expected room_ready for p2, got %v", msg2["type"])
	}
	if msg2["player_role"] != "player_2" {
		t.Fatalf("expected player_2 role, got %v", msg2["player_role"])
	}
}

func TestHub_StateUpdate_BroadcastOnTick(t *testing.T) {
	hub := startTestHub()
	conn1 := dialTestServer(t, hub, "p1")
	conn2 := dialTestServer(t, hub, "p2")
	time.Sleep(50 * time.Millisecond)

	sendMsg(t, conn1, ws.JoinRoom{Type: "join_room", GameType: "team_tower"})
	time.Sleep(100 * time.Millisecond)
	sendMsg(t, conn2, ws.JoinRoom{Type: "join_room", GameType: "team_tower"})

	var wMsg map[string]interface{}
	readMsg(t, conn1, &wMsg) // discard waiting_for_partner

	var rr1, rr2 map[string]interface{}
	readMsg(t, conn1, &rr1)
	readMsg(t, conn2, &rr2)

	var su map[string]interface{}
	readMsg(t, conn1, &su)
	if su["type"] != "state_update" {
		t.Fatalf("expected state_update, got %v", su["type"])
	}
	if su["tick"] == nil || su["tick"].(float64) < 1 {
		t.Fatalf("expected tick >= 1, got %v", su["tick"])
	}
}

func TestHub_HeartbeatPingPong(t *testing.T) {
	hub := startTestHub()
	conn := dialTestServer(t, hub, "p1")
	time.Sleep(50 * time.Millisecond)

	sendMsg(t, conn, ws.HeartbeatPing{Type: "heartbeat_ping", Timestamp: 12345})

	var pong map[string]interface{}
	readMsg(t, conn, &pong)
	if pong["type"] != "heartbeat_pong" {
		t.Fatalf("expected heartbeat_pong, got %v", pong["type"])
	}
}

func TestHub_EnforceOneRoomPerProfile(t *testing.T) {
	hub := startTestHub()
	conn := dialTestServer(t, hub, "p1")
	time.Sleep(50 * time.Millisecond)

	sendMsg(t, conn, ws.JoinRoom{Type: "join_room", GameType: "team_tower"})
	time.Sleep(100 * time.Millisecond)

	sendMsg(t, conn, ws.JoinRoom{Type: "join_room", GameType: "team_tower"})

	var wMsg map[string]interface{}
	readMsg(t, conn, &wMsg)
	
	var errMsg map[string]interface{}
	readMsg(t, conn, &errMsg)
	if errMsg["type"] != "error" {
		t.Fatalf("expected error, got %v", errMsg["type"])
	}
	if errMsg["code"] != "ALREADY_IN_ROOM" {
		t.Fatalf("expected ALREADY_IN_ROOM, got %v", errMsg["code"])
	}
}

func TestHub_LeaveRoom_NotifiesPartner(t *testing.T) {
	hub := startTestHub()
	conn1 := dialTestServer(t, hub, "p1")
	conn2 := dialTestServer(t, hub, "p2")
	time.Sleep(50 * time.Millisecond)

	sendMsg(t, conn1, ws.JoinRoom{Type: "join_room", GameType: "team_tower"})
	time.Sleep(100 * time.Millisecond)
	sendMsg(t, conn2, ws.JoinRoom{Type: "join_room", GameType: "team_tower"})

	var wMsg map[string]interface{}
	readMsg(t, conn1, &wMsg) // discard waiting_for_partner

	var dummy map[string]interface{}
	readMsg(t, conn1, &dummy) // discard room_ready
	readMsg(t, conn2, &dummy) // discard room_ready

	sendMsg(t, conn1, ws.LeaveRoom{Type: "leave_room"})

	for i := 0; i < 5; i++ {
		var msg map[string]interface{}
		readMsg(t, conn2, &msg)
		if msg["type"] == "player_disconnected" {
			if msg["player_role"] != "player_1" {
				t.Fatalf("expected player_1 disconnected, got %v", msg["player_role"])
			}
			return
		}
	}
	t.Fatal("did not receive player_disconnected")
}

func TestHub_UnknownMessageType(t *testing.T) {
	hub := startTestHub()
	conn := dialTestServer(t, hub, "p1")
	time.Sleep(50 * time.Millisecond)

	sendMsg(t, conn, map[string]string{"type": "banana"})

	var errMsg map[string]interface{}
	readMsg(t, conn, &errMsg)
	if errMsg["type"] != "error" {
		t.Fatalf("expected error, got %v", errMsg["type"])
	}
	if errMsg["code"] != "UNKNOWN_TYPE" {
		t.Fatalf("expected UNKNOWN_TYPE, got %v", errMsg["code"])
	}
}

func TestHub_PlayerAction_NotInRoom(t *testing.T) {
	hub := startTestHub()
	conn := dialTestServer(t, hub, "p1")
	time.Sleep(50 * time.Millisecond)

	sendMsg(t, conn, ws.PlayerAction{
		Type:            "player_action",
		Payload:         json.RawMessage(`{}`),
		ClientTimestamp: 100,
	})

	var rejected map[string]interface{}
	readMsg(t, conn, &rejected)
	if rejected["type"] != "action_rejected" {
		t.Fatalf("expected action_rejected, got %v", rejected["type"])
	}
}

func TestRoom_MaxTwoPlayers(t *testing.T) {
	hub := startTestHub()
	conn1 := dialTestServer(t, hub, "p1")
	conn2 := dialTestServer(t, hub, "p2")
	conn3 := dialTestServer(t, hub, "p3")
	time.Sleep(50 * time.Millisecond)

	sendMsg(t, conn1, ws.JoinRoom{Type: "join_room", GameType: "team_tower"})
	time.Sleep(100 * time.Millisecond)
	sendMsg(t, conn2, ws.JoinRoom{Type: "join_room", GameType: "team_tower"})
	time.Sleep(100 * time.Millisecond)
	sendMsg(t, conn3, ws.JoinRoom{Type: "join_room", GameType: "team_tower"})
	time.Sleep(100 * time.Millisecond)

	if got := hub.RoomCount(); got != 2 {
		t.Fatalf("expected 2 rooms (one full, one waiting), got %d", got)
	}
}
