// noinspection JSUnusedGlobalSymbols

import grpc from 'k6/net/grpc';
import { check, group } from 'k6';
import { Counter, Rate } from 'k6/metrics';
import * as helpers from '../../helpers.js';

const successRate = new Rate('success_rate');
const connectionErrors = new Counter('connection_errors');
const authErrors = new Counter('auth_errors');
const listUsersErrors = new Counter('list_users_errors');
const getUserErrors = new Counter('get_user_errors');
const createUserErrors = new Counter('create_user_errors');
const updateUserErrors = new Counter('update_user_errors');
const deleteUserErrors = new Counter('delete_user_errors');

const ADDR = helpers.GRPCServerAddress();

// Note: We need a pre-created test user with admin permissions to run these tests
// The auth service (signup/login) is only available via HTTP, not gRPC
const TEST_ACCESS_TOKEN = 'test-admin-token'; // This should be replaced with a valid token

const client = new grpc.Client();

// Load proto files - adjust path as needed
try {
    client.load(['../../../../api/proto'], 'auth/v1/auth.proto');
} catch (e) {
    console.error('Failed to load proto files:', e);
}

export const options = {
    scenarios: {
        user_management: {
            executor: 'ramping-vus',
            startVUs: 1,
            stages: [
                { duration: '10s', target: 2 },
                { duration: '20s', target: 3 },
                { duration: '10s', target: 1 },
            ],
        },
    },
    thresholds: {
        'grpc_req_duration': ['p(95)<300'],
        'grpc_req_duration{rpc:ListUsers}': ['p(95)<200'],
        'grpc_req_duration{rpc:GetUserByUUID}': ['p(95)<150'],
        'grpc_req_duration{rpc:CreateUser}': ['p(95)<300'],
        'grpc_req_duration{rpc:UpdateUser}': ['p(95)<300'],
        'grpc_req_duration{rpc:DeleteUser}': ['p(95)<200'],
        'success_rate': ['rate>=0.7'], // Allow some failures
        'connection_errors': ['count<20'],
        'auth_errors': ['count<50'], // Auth errors are expected if token is invalid
    },
};

const createdUsers = [];

export default function() {
    try {
        client.connect(ADDR, {
            plaintext: true,
            timeout: '5s',
        });
    } catch (e) {
        connectionErrors.add(1);
        helpers.randomSleep(2);
        return;
    }

    try {
        runUserManagementOperations();
    } finally {
        client.close();
    }
}

function runUserManagementOperations() {
    group('User Management Operations', () => {
        // Try to list users first to check if we have valid auth
        const canListUsers = listUsers();
        if (!canListUsers) {
            // If we can't list users, we likely don't have valid auth
            authErrors.add(1);
            helpers.randomSleep(1);
            return;
        }

        // Perform various operations
        if (Math.random() < 0.3) {
            const userId = createUser();
            if (userId) {
                createdUsers.push(userId);
                
                helpers.randomSleep(0.5);
                
                if (Math.random() < 0.5) {
                    updateUser(userId);
                }
                
                helpers.randomSleep(0.5);
                
                if (Math.random() < 0.3) {
                    deleteUser(userId);
                    const index = createdUsers.indexOf(userId);
                    if (index > -1) {
                        createdUsers.splice(index, 1);
                    }
                }
            }
        }

        if (createdUsers.length > 0 && Math.random() < 0.5) {
            const randomId = createdUsers[Math.floor(Math.random() * createdUsers.length)];
            getUserByUUID(randomId);
        }

        helpers.randomSleep(0.5);
    });
}

function listUsers() {
    const data = {
        limit: 10,
        offset: 0
    };
    
    const params = {
        metadata: {
            'authorization': `Bearer ${TEST_ACCESS_TOKEN}`,
        },
        tags: { rpc: 'ListUsers' },
        timeout: '5s',
    };
    
    try {
        const response = client.invoke('auth.v1.AuthService/ListUsers', data, params);
        
        const success = check(response, {
            'list users status is OK': (r) => r && r.status === grpc.StatusOK,
            'list users response has users': (r) => r && r.message && Array.isArray(r.message.users),
        });
        
        successRate.add(success);
        if (!success) {
            listUsersErrors.add(1);
            if (response && response.error) {
                const errorCode = response.error.code || response.status;
                if (errorCode !== grpc.StatusUnauthenticated && errorCode !== grpc.StatusPermissionDenied) {
                    console.log(`Failed to list users: ${errorCode} ${JSON.stringify(response.error)}`);
                }
            }
            return false;
        }
        return true;
    } catch (e) {
        connectionErrors.add(1);
        return false;
    }
}

function createUser() {
    const email = `test-${helpers.GenerateRandomString()}@example.com`;
    const data = {
        email: email,
        password: 'TestPass123!'
    };
    
    const params = {
        metadata: {
            'authorization': `Bearer ${TEST_ACCESS_TOKEN}`,
        },
        tags: { rpc: 'CreateUser' },
        timeout: '5s',
    };
    
    try {
        const response = client.invoke('auth.v1.AuthService/CreateUser', data, params);
        
        const success = check(response, {
            'create user status is OK': (r) => r && r.status === grpc.StatusOK,
            'create user response has user': (r) => r && r.message && r.message.user,
            'create user response has uuid': (r) => r && r.message && r.message.user && r.message.user.uuid,
        });
        
        successRate.add(success);
        if (!success) {
            createUserErrors.add(1);
            if (response && response.error) {
                const errorCode = response.error.code || response.status;
                if (errorCode !== grpc.StatusAlreadyExists) {
                    console.log(`Failed to create user: ${errorCode} ${JSON.stringify(response.error)}`);
                }
            }
            return null;
        }
        return response.message.user.uuid;
    } catch (e) {
        connectionErrors.add(1);
        return null;
    }
}

function getUserByUUID(uuid) {
    const data = {
        uuid: uuid
    };
    
    const params = {
        metadata: {
            'authorization': `Bearer ${TEST_ACCESS_TOKEN}`,
        },
        tags: { rpc: 'GetUserByUUID' },
        timeout: '5s',
    };
    
    try {
        const response = client.invoke('auth.v1.AuthService/GetUserByUUID', data, params);
        
        const success = check(response, {
            'get user status is OK or NotFound': (r) => r && (r.status === grpc.StatusOK || r.status === grpc.StatusNotFound),
            'get user response has user if OK': (r) => r.status === grpc.StatusNotFound || (r && r.message && r.message.user),
        });
        
        successRate.add(success);
        if (!success) {
            getUserErrors.add(1);
            if (response && response.error) {
                const errorCode = response.error.code || response.status;
                console.log(`Failed to get user: ${errorCode} ${JSON.stringify(response.error)}`);
            }
        }
    } catch (e) {
        connectionErrors.add(1);
    }
}

function updateUser(uuid) {
    const newEmail = `updated-${helpers.GenerateRandomString()}@example.com`;
    const data = {
        uuid: uuid,
        email: newEmail
    };
    
    const params = {
        metadata: {
            'authorization': `Bearer ${TEST_ACCESS_TOKEN}`,
        },
        tags: { rpc: 'UpdateUser' },
        timeout: '5s',
    };
    
    try {
        const response = client.invoke('auth.v1.AuthService/UpdateUser', data, params);
        
        const success = check(response, {
            'update user status is OK': (r) => r && r.status === grpc.StatusOK,
        });
        
        successRate.add(success);
        if (!success) {
            updateUserErrors.add(1);
            if (response && response.error) {
                const errorCode = response.error.code || response.status;
                if (errorCode !== grpc.StatusNotFound) {
                    console.log(`Failed to update user: ${errorCode} ${JSON.stringify(response.error)}`);
                }
            }
        }
    } catch (e) {
        connectionErrors.add(1);
    }
}

function deleteUser(uuid) {
    const data = {
        uuid: uuid
    };
    
    const params = {
        metadata: {
            'authorization': `Bearer ${TEST_ACCESS_TOKEN}`,
        },
        tags: { rpc: 'DeleteUser' },
        timeout: '5s',
    };
    
    try {
        const response = client.invoke('auth.v1.AuthService/DeleteUser', data, params);
        
        const success = check(response, {
            'delete user status is OK': (r) => r && r.status === grpc.StatusOK,
        });
        
        successRate.add(success);
        if (!success) {
            deleteUserErrors.add(1);
            if (response && response.error) {
                const errorCode = response.error.code || response.status;
                if (errorCode !== grpc.StatusNotFound) {
                    console.log(`Failed to delete user: ${errorCode} ${JSON.stringify(response.error)}`);
                }
            }
        }
    } catch (e) {
        connectionErrors.add(1);
    }
}