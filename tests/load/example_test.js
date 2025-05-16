import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

const createFruitErrors = new Counter('create_fruit_errors');
const updateFruitErrors = new Counter('update_fruit_errors');
const deleteFruitErrors = new Counter('delete_fruit_errors');
const getFruitErrors = new Counter('get_fruit_errors');
const getFruitsErrors = new Counter('get_fruits_errors');

const requestDuration = new Trend('request_duration');
const successRate = new Rate('success_rate');

export const options = {
    scenarios: {
        crud_operations: {
            executor: 'ramping-vus',
            startVUs: 5,
            stages: [
                { duration: '20s', target: 5 },
                { duration: '30s', target: 25 },
                { duration: '10s', target: 10 },
            ],
        },
        get_operations: {
            executor: 'constant-arrival-rate',
            rate: 500,
            timeUnit: '1m',
            duration: '1m',
            preAllocatedVUs: 10,
            maxVUs: 25,
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<1000'],
        'http_req_duration{name:get_fruits}': ['p(95)<500'],
        'http_req_duration{name:get_fruit_by_id}': ['p(95)<300'],
        'success_rate': ['rate>0.95'],
    },
};

const BASE_URL = 'http://localhost:8080/api/v1';
const createdFruitIds = [];

export default function() {
    const scenario = __ENV.SCENARIO;

    if (scenario === 'crud_operations' || !scenario) {
        runCrudOperations();
    } else if (scenario === 'get_operations' || !scenario) {
        runGetOperations();
    }
}

function runCrudOperations() {
    group('Fruit CRUD operations', () => {

        let fruitId = createFruit();

        if (fruitId) {
            createdFruitIds.push(fruitId);
            updateFruit(fruitId);
            sleep(1);
            if (Math.random() < 0.3) {
                deleteFruit(fruitId);
                const index = createdFruitIds.indexOf(fruitId);
                if (index > -1) {
                    createdFruitIds.splice(index, 1);
                }
            }
        }
        sleep(Math.random() * 2);
    });
}

function runGetOperations() {
    group('Fruit GET operations', () => {
        getFruits(Math.floor(Math.random() * 20), Math.floor(Math.random() * 5));
        if (createdFruitIds.length > 0) {
            const randomId = createdFruitIds[Math.floor(Math.random() * createdFruitIds.length)];
            getFruitById(randomId);
        } else {
            getFruitById(Math.floor(Math.random() * 10) + 1);
        }
        sleep(Math.random() * 0.5);
    });
}

function createFruit() {
    const url = `${BASE_URL}/fruits`;
    const payload = JSON.stringify({
        name: `Fruit-${randomString(5)}`
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
        tags: { name: 'create_fruit' },
    };

    const startTime = new Date().getTime();
    const res = http.post(url, payload, params);
    const duration = new Date().getTime() - startTime;

    requestDuration.add(duration);

    const success = check(res, {
        'create fruit status is 201': (r) => r.status === 201,
        'create fruit response has id': (r) => r.json('id') !== undefined,
    });

    successRate.add(success);

    if (!success) {
        createFruitErrors.add(1);
        console.log(`Failed to create fruit: ${res.status} ${res.body}`);
        return null;
    }

    return res.json('id');
}

function updateFruit(id) {
    const url = `${BASE_URL}/fruits/${id}`;
    const payload = JSON.stringify({
        name: `Updated-${randomString(5)}`
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
        tags: { name: 'update_fruit' },
    };

    const startTime = new Date().getTime();
    const res = http.put(url, payload, params);
    const duration = new Date().getTime() - startTime;

    requestDuration.add(duration);

    const success = check(res, {
        'update fruit status is 200': (r) => r.status === 200,
    });

    successRate.add(success);

    if (!success) {
        updateFruitErrors.add(1);
        console.log(`Failed to update fruit ${id}: ${res.status} ${res.body}`);
    }
}

function deleteFruit(id) {
    const url = `${BASE_URL}/fruits/${id}`;
    const params = {
        tags: { name: 'delete_fruit' },
    };

    const startTime = new Date().getTime();
    const res = http.del(url, null, params);
    const duration = new Date().getTime() - startTime;

    requestDuration.add(duration);

    const success = check(res, {
        'delete fruit status is 200': (r) => r.status === 200,
    });

    successRate.add(success);

    if (!success) {
        deleteFruitErrors.add(1);
        console.log(`Failed to delete fruit ${id}: ${res.status} ${res.body}`);
    }
}

function getFruitById(id) {
    const url = `${BASE_URL}/fruits/${id}`;
    const params = {
        tags: { name: 'get_fruit_by_id' },
    };

    const startTime = new Date().getTime();
    const res = http.get(url, params);
    const duration = new Date().getTime() - startTime;

    requestDuration.add(duration);

    const success = check(res, {
        'get fruit by id status is 200': (r) => r.status === 200 || r.status === 404,
        'get fruit by id has name': (r) => r.status === 404 || r.json('name') !== undefined,
    });

    successRate.add(success);

    if (!success) {
        getFruitErrors.add(1);
        console.log(`Failed to get fruit ${id}: ${res.status} ${res.body}`);
    }
}

function getFruits(limit = 10, offset = 0) {
    const url = `${BASE_URL}/fruits?limit=${limit}&offset=${offset}`;
    const params = {
        tags: { name: 'get_fruits' },
    };

    const startTime = new Date().getTime();
    const res = http.get(url, params);
    const duration = new Date().getTime() - startTime;

    requestDuration.add(duration);

    const success = check(res, {
        'get fruits status is 200': (r) => r.status === 200,
        'get fruits returns array': (r) => Array.isArray(r.json()),
    });

    successRate.add(success);

    if (!success) {
        getFruitsErrors.add(1);
        console.log(`Failed to get fruits: ${res.status} ${res.body}`);
    }

    // Store fruit IDs for later use
    if (success && res.json().length > 0) {
        res.json().forEach(fruit => {
            if (!createdFruitIds.includes(fruit.id)) {
                createdFruitIds.push(fruit.id);
            }
        });
    }
}
