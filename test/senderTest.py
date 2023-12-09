import base64
import json
import socket
import time
from typing import Tuple


def senderTest(ip: str, port: int) -> None:
    udp_socket: socket.socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    server_address: Tuple[str, int] = (ip, port)
    udp_socket.bind(server_address)

    print(f"udp listen start... ip:{ip}, port:{port}")

    try:
        while True:
            data, client_address = udp_socket.recvfrom(4096)
            message: bytes = base64.b64decode(json.loads(data.decode("utf-8"))["Data"])
            print(
                f"receive address: {client_address}, data:{data.decode('utf-8')}, message:{message.decode('utf-8')}"
            )
            time.sleep(1)
    except KeyboardInterrupt:
        return
    finally:
        udp_socket.close()


if __name__ == "__main__":
    senderTest("127.0.0.1", 7777)
