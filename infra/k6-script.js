import ws from "k6/ws";
import { check, sleep } from "k6";

export let options = {
  vus: 200,            // thay khi cần
  duration: "30s",
  // thresholds: {},   // bỏ thresholds hoặc cấu hình metric hợp lệ
};

export default function () {
  const host = __ENV.WS_HOST || "host.docker.internal"; // or "localhost" if running locally
  const url = `ws://${host}:8080/ws`;
  const params = { tags: { my_tag: "e5" } };

  const res = ws.connect(url, params, function (socket) {
    socket.on("open", function () {
      const join = JSON.stringify({ type: "join", user: `vu-${__VU}` });
      socket.send(join);
    });

    socket.setInterval(function () {
      const msg = JSON.stringify({ type: "message", from: `vu-${__VU}`, text: "hello from k6" });
      socket.send(msg);
    }, 1000);

    socket.on("message", function (data) {
      // optional: handle incoming message
    });

    socket.on("close", function () {});

    socket.on("error", function (e) {});
    // keep connection a bit
    sleep(10);
    socket.close();
  });

  check(res, { "status is 101": (r) => r && r.status === 101 });
}