import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
    thresholds: {
        http_req_duration: ["p(99) < 3000"],
    },

    stages: [
        { duration: "5s", target: 10 },
        { duration: "5s", target: 15 },
        { duration: "5s", target: 20 },
        { duration: "5s", target: 50 },
        { duration: "5s", target: 1 },
    ],
};

export default function () {
    let res = http.get("http://localhost:8080/api/v1/fruits");
    check(res, { "status was 200": (r) => r.status == 200 });
    sleep(1);
}