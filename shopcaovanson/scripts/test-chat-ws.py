import json
import sys
import time
import urllib.request

import websocket

API = "http://localhost:8080"
WS = "ws://localhost:8084"


def login(email: str, password: str) -> str:
    req = urllib.request.Request(
        f"{API}/api/auth/login",
        data=json.dumps({"email": email, "password": password}).encode(),
        headers={"Content-Type": "application/json"},
    )
    return json.load(urllib.request.urlopen(req))["access_token"]


def support_room(token: str, customer_id: str) -> str:
    req = urllib.request.Request(
        f"{API}/api/chat/admin/support-rooms",
        data=json.dumps({"customer_id": customer_id}).encode(),
        headers={"Authorization": f"Bearer {token}", "Content-Type": "application/json"},
        method="POST",
    )
    return json.load(urllib.request.urlopen(req))["data"]["id"]


def main() -> None:
    customer_id = sys.argv[1] if len(sys.argv) > 1 else "1daf4b90-93b7-4662-a5ff-3478634b431a"
    admin_token = login("admin@shop.com", "Admin@123")
    room_id = support_room(admin_token, customer_id)
    print("room", room_id)

    ws = websocket.create_connection(f"{WS}/ws?token={admin_token}", timeout=5)
    ws.settimeout(3)
    print("connected")

    join = json.dumps({"type": "join", "room_id": room_id})
    ws.send(join)
    print("sent join", join)
    time.sleep(0.5)

    msg = json.dumps(
        {
            "type": "message",
            "room_id": room_id,
            "content": "python test",
            "msg_type": "text",
        }
    )
    ws.send(msg)
    print("sent message", msg)

    for i in range(5):
        try:
            print("recv", ws.recv())
        except Exception as exc:
            print("recv_err", exc)
            break
        time.sleep(0.3)

    ws.close()


if __name__ == "__main__":
    main()
