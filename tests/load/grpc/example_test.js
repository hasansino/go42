// noinspection JSUnusedGlobalSymbols

import grpc from 'k6/net/grpc';
import { check, group } from 'k6';
import { Counter, Rate } from 'k6/metrics';
import * as helpers from '../helpers.js';

const successRate = new Rate('success_rate');
const getFruitsErrors = new Counter('get_fruits_errors');
const getFruitErrors = new Counter('get_fruit_errors');
const createFruitErrors = new Counter('create_fruit_errors');
const updateFruitErrors = new Counter('update_fruit_errors');
const deleteFruitErrors = new Counter('delete_fruit_errors');

const ADDR = helpers.GRPCServerAddress();

const client = new grpc.Client();
client.load(['../../..'], './internal/example/provider/grpc/example.proto');

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
        'grpc_req_duration': ['p(95)<50'],
        'grpc_req_duration{rpc:ListFruits}': ['p(95)<50'],
        'grpc_req_duration{rpc:GetFruit}': ['p(95)<50'],
        'grpc_req_duration{rpc:CreateFruit}': ['p(95)<50'],
        'grpc_req_duration{rpc:UpdateFruit}': ['p(95)<50'],
        'grpc_req_duration{rpc:DeleteFruit}': ['p(95)<50'],
        'success_rate': ['rate>=1.0'],
    },
};

export default function() {
    client.connect(ADDR, {plaintext: true});
    runCrudOperations();
    runGetOperations();
    client.close();
}

const createdFruitIds = [];

function runCrudOperations() {
    group('Fruit CRUD operations', () => {
        let fruitId = createFruit(client);
        if (fruitId) {
            createdFruitIds.push(fruitId);
            updateFruit(client, fruitId);
            helpers.randomSleep(1)
            if (Math.random() < 0.3) {
                deleteFruit(client, fruitId);
                const index = createdFruitIds.indexOf(fruitId);
                if (index > -1) {
                    createdFruitIds.splice(index, 1);
                }
            }
        }
        helpers.randomSleep(2)
    });
}

function runGetOperations() {
    group('Fruit GET operations', () => {
        getFruits(client, Math.floor(Math.random() * 20), Math.floor(Math.random() * 5));
        if (createdFruitIds.length > 0) {
            const randomId = createdFruitIds[Math.floor(Math.random() * createdFruitIds.length)];
            getFruitById(client, randomId);
        } else {
            getFruitById(client, Math.floor(Math.random() * 10) + 1);
        }
        helpers.randomSleep(0.5)
    });
}

function createFruit(client) {
    const data = {
        name: helpers.GenerateRandomString()
    };
    const response = client.invoke('grpc.ExampleService/CreateFruit', data, {
        tags: { rpc: 'CreateFruit' }
    });
    const success = check(response, {
        'create fruit status is OK': (r) => r.status === grpc.StatusOK,
        'create fruit response has fruit': (r) => r && r.message && r.message.fruit,
        'create fruit response has id': (r) => r && r.message && r.message.fruit && r.message.fruit.id,
    });
    successRate.add(success);
    if (!success) {
        createFruitErrors.add(1);
        console.log(`Failed to create fruit: ${response.status} ${JSON.stringify(response.error)}`);
        return null;
    }
    return response.message.fruit.id;
}

function updateFruit(client, id) {
    const data = {
        id: id,
        name: helpers.GenerateRandomString()
    };
    const response = client.invoke('grpc.ExampleService/UpdateFruit', data, {
        tags: { rpc: 'UpdateFruit' }
    });
    const success = check(response, {
        'update fruit status is OK': (r) => r.status === grpc.StatusOK,
        'update fruit response has fruit': (r) => r && r.message && r.message.fruit,
    });
    successRate.add(success);
    if (!success) {
        updateFruitErrors.add(1);
        console.log(`Failed to update fruit ${id}: ${response.status} ${JSON.stringify(response.error)}`);
    }
}

function deleteFruit(client, id) {
    const data = {
        id: id
    };
    const response = client.invoke('grpc.ExampleService/DeleteFruit', data, {
        tags: { rpc: 'DeleteFruit' }
    });
    const success = check(response, {
        'delete fruit status is OK': (r) => r.status === grpc.StatusOK,
        'delete fruit response has success': (r) => r && r.message && r.message.success === true,
    });
    successRate.add(success);
    if (!success) {
        deleteFruitErrors.add(1);
        console.log(`Failed to delete fruit ${id}: ${response.status} ${JSON.stringify(response.error)}`);
    }
}

function getFruitById(client, id) {
    const data = {
        id: id
    };
    const response = client.invoke('grpc.ExampleService/GetFruit', data, {
        tags: { rpc: 'GetFruit' }
    });
    const success = check(response, {
        'get fruit status is valid': (r) => r.status === grpc.StatusOK || r.status === grpc.StatusNotFound,
        'get fruit response has name': (r) => r.status === grpc.StatusNotFound || (r && r.message && r.message.name),
    });
    successRate.add(success);
    if (!success) {
        getFruitErrors.add(1);
        console.log(`Failed to get fruit ${id}: ${response.status} ${JSON.stringify(response.error)}`);
    }
}

function getFruits(client, limit = 10, offset = 0) {
    const data = {
        limit: limit,
        offset: offset
    };
    const response = client.invoke('grpc.ExampleService/ListFruits', data, {
        tags: { rpc: 'ListFruits' }
    });
    const success = check(response, {
        'list fruits status is OK': (r) => r.status === grpc.StatusOK,
        'list fruits response has fruits': (r) => r && r.message && Array.isArray(r.message.fruits),
    });
    successRate.add(success);
    if (!success) {
        getFruitsErrors.add(1);
        console.log(`Failed to get fruits: ${response.status} ${JSON.stringify(response.error)}`);
    }
    if (success && response.message && response.message.fruits && response.message.fruits.length > 0) {
        response.message.fruits.forEach(fruit => {
            if (!createdFruitIds.includes(fruit.id)) {
                createdFruitIds.push(fruit.id);
            }
        });
    }
}