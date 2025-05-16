import http from "k6/http";
import { check, sleep } from "k6";

function generateRandomStages(totalDurationMinutes, stageDurationMinutes, minVUs, maxVUs) {
    const stages = [];
    const numberOfStages = totalDurationMinutes / stageDurationMinutes;
    for (let i = 0; i < numberOfStages; i++) {
        const randomVUs = Math.floor(Math.random() * (maxVUs - minVUs + 1)) + minVUs;
        stages.push({ duration: `${stageDurationMinutes}m`, target: randomVUs });
    }
    stages.push({ duration: `${stageDurationMinutes}m`, target: 0 });
    return stages;
}

const totalTestDurationMin = 60;
const stageDurationMin = 1;
const minVirtualUsers = 1;
const maxVirtualUsers = 150;

export let options = {
    stages: generateRandomStages(totalTestDurationMin, stageDurationMin, minVirtualUsers, maxVirtualUsers),
};

export default function () {
    let res = http.get("http://localhost:8080/api/v1/fruits");
    check(res, { "status was 200": (r) => r.status == 200 });
    sleep(1);
}