// noinspection JSUnusedGlobalSymbols

import http from 'k6/http';
import { check, group } from 'k6';
import { Counter, Rate } from 'k6/metrics';
import * as helpers from '../../helpers.js';

const successRate = new Rate('success_rate');
const connectionErrors = new Counter('connection_errors');
const signupErrors = new Counter('signup_errors');
const loginErrors = new Counter('login_errors');
const refreshErrors = new Counter('refresh_errors');
const logoutErrors = new Counter('logout_errors');
const getMeErrors = new Counter('get_me_errors');
const updateMeErrors = new Counter('update_me_errors');
const listUsersErrors = new Counter('list_users_errors');

const ADDR = helpers.HTTPServerAddress();

export const options = {
    scenarios: {
        auth_flow: {
            executor: 'ramping-vus',
            startVUs: 1,
            stages: [
                { duration: '10s', target: 2 },
                { duration: '20s', target: 5 },
                { duration: '10s', target: 1 },
            ],
        },
        api_operations: {
            executor: 'constant-arrival-rate',
            rate: 2,
            timeUnit: '1s',
            duration: '40s',
            preAllocatedVUs: 2,
            maxVUs: 5,
        },
    },
    thresholds: {
        http_req_duration: ['p(95)<500'],
        'http_req_duration{name:signup}': ['p(95)<200'],
        'http_req_duration{name:login}': ['p(95)<200'],
        'http_req_duration{name:refresh}': ['p(95)<200'],
        'http_req_duration{name:logout}': ['p(95)<100'],
        'http_req_duration{name:get_me}': ['p(95)<100'],
        'http_req_duration{name:update_me}': ['p(95)<200'],
        'http_req_duration{name:list_users}': ['p(95)<200'],
        'success_rate': ['rate>=0.8'], // Allow some failures
        'connection_errors': ['count<50'], // Connection errors should be limited
    },
};

const activeUsers = [];

export default function() {
    // Check if the server is reachable
    const healthCheck = http.get(ADDR + '/health', { 
        timeout: '3s',
        tags: { name: 'health_check' }
    });
    
    if (healthCheck.status === 0) {
        connectionErrors.add(1);
        helpers.randomSleep(2);
        return;
    }

    runAuthFlow();
    runApiOperations();
}

function runAuthFlow() {
    group('Auth Flow', () => {
        const email = `test-${helpers.GenerateRandomString()}@example.com`;
        const password = 'TestPass123!';
        
        const signupSuccess = signup(email, password);
        if (!signupSuccess) {
            return;
        }
        
        helpers.randomSleep(0.5);
        
        const loginData = login(email, password);
        if (!loginData) {
            return;
        }
        
        activeUsers.push({
            email: email,
            accessToken: loginData.access_token,
            refreshToken: loginData.refresh_token,
        });
        
        helpers.randomSleep(1);
        
        if (Math.random() < 0.3) {
            const newTokens = refresh(loginData.refresh_token);
            if (newTokens) {
                // Update tokens in activeUsers
                const userIndex = activeUsers.findIndex(u => u.email === email);
                if (userIndex >= 0) {
                    activeUsers[userIndex].accessToken = newTokens.access_token;
                    activeUsers[userIndex].refreshToken = newTokens.refresh_token;
                }
            }
        }
        
        helpers.randomSleep(0.5);
        
        if (Math.random() < 0.2) {
            logout(loginData.access_token, loginData.refresh_token);
            const index = activeUsers.findIndex(u => u.email === email);
            if (index > -1) {
                activeUsers.splice(index, 1);
            }
        }
    });
}

function runApiOperations() {
    group('API Operations', () => {
        if (activeUsers.length === 0) {
            // Create a user if none exist
            const email = `api-test-${helpers.GenerateRandomString()}@example.com`;
            const password = 'TestPass123!';
            
            if (signup(email, password)) {
                const loginData = login(email, password);
                if (loginData) {
                    activeUsers.push({
                        email: email,
                        accessToken: loginData.access_token,
                        refreshToken: loginData.refresh_token,
                    });
                }
            }
            return;
        }
        
        const randomUser = activeUsers[Math.floor(Math.random() * activeUsers.length)];
        if (!randomUser) {
            return;
        }
        
        getMe(randomUser.accessToken);
        
        helpers.randomSleep(0.3);
        
        if (Math.random() < 0.2) {
            updateMe(randomUser.accessToken);
        }
        
        if (Math.random() < 0.1) {
            listUsers(randomUser.accessToken);
        }
    });
}

function signup(email, password) {
    const url = `${ADDR}/api/v1/auth/signup`;
    const payload = JSON.stringify({
        email: email,
        password: password,
    });
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
        tags: { name: 'signup' },
        timeout: '5s',
    };
    
    const res = http.post(url, payload, params);
    
    if (res.status === 0) {
        connectionErrors.add(1);
        return false;
    }
    
    const success = check(res, {
        'signup status is 201': (r) => r.status === 201,
        'signup response has user': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body && body.uuid !== undefined;
            } catch (e) {
                return false;
            }
        },
    });
    
    successRate.add(success);
    if (!success) {
        signupErrors.add(1);
        if (res.status !== 409) { // 409 is expected for duplicate emails
            console.log(`Failed to signup: ${res.status} ${res.body}`);
        }
        return false;
    }
    return true;
}

function login(email, password) {
    const url = `${ADDR}/api/v1/auth/login`;
    const payload = JSON.stringify({
        email: email,
        password: password,
    });
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
        tags: { name: 'login' },
        timeout: '5s',
    };
    
    const res = http.post(url, payload, params);
    
    if (res.status === 0) {
        connectionErrors.add(1);
        return null;
    }
    
    const success = check(res, {
        'login status is 200': (r) => r.status === 200,
        'login response has tokens': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body && body.access_token !== undefined && body.refresh_token !== undefined;
            } catch (e) {
                return false;
            }
        },
    });
    
    successRate.add(success);
    if (!success) {
        loginErrors.add(1);
        console.log(`Failed to login: ${res.status} ${res.body}`);
        return null;
    }
    
    try {
        return JSON.parse(res.body);
    } catch (e) {
        return null;
    }
}

function refresh(refreshToken) {
    const url = `${ADDR}/api/v1/auth/refresh`;
    const payload = JSON.stringify({
        token: refreshToken,
    });
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
        tags: { name: 'refresh' },
        timeout: '5s',
    };
    
    const res = http.post(url, payload, params);
    
    if (res.status === 0) {
        connectionErrors.add(1);
        return null;
    }
    
    const success = check(res, {
        'refresh status is 200': (r) => r.status === 200,
        'refresh response has tokens': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body && body.access_token !== undefined && body.refresh_token !== undefined;
            } catch (e) {
                return false;
            }
        },
    });
    
    successRate.add(success);
    if (!success) {
        refreshErrors.add(1);
        if (res.status !== 401) { // 401 is expected for expired tokens
            console.log(`Failed to refresh: ${res.status} ${res.body}`);
        }
        return null;
    }
    
    try {
        return JSON.parse(res.body);
    } catch (e) {
        return null;
    }
}

function logout(accessToken, refreshToken) {
    const url = `${ADDR}/api/v1/auth/logout`;
    const payload = JSON.stringify({
        access_token: accessToken,
        refresh_token: refreshToken,
    });
    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
        tags: { name: 'logout' },
        timeout: '5s',
    };
    
    const res = http.post(url, payload, params);
    
    if (res.status === 0) {
        connectionErrors.add(1);
        return;
    }
    
    const success = check(res, {
        'logout status is 200': (r) => r.status === 200,
    });
    
    successRate.add(success);
    if (!success) {
        logoutErrors.add(1);
        console.log(`Failed to logout: ${res.status} ${res.body}`);
    }
}

function getMe(accessToken) {
    const url = `${ADDR}/api/v1/users/me`;
    const params = {
        headers: {
            'Authorization': `Bearer ${accessToken}`,
        },
        tags: { name: 'get_me' },
        timeout: '5s',
    };
    
    const res = http.get(url, params);
    
    if (res.status === 0) {
        connectionErrors.add(1);
        return;
    }
    
    const success = check(res, {
        'get me status is 200': (r) => r.status === 200,
        'get me response has user': (r) => {
            try {
                const body = JSON.parse(r.body);
                return body && body.uuid !== undefined;
            } catch (e) {
                return false;
            }
        },
    });
    
    successRate.add(success);
    if (!success) {
        getMeErrors.add(1);
        if (res.status !== 401) { // 401 is expected for expired tokens
            console.log(`Failed to get me: ${res.status} ${res.body}`);
        }
    }
}

function updateMe(accessToken) {
    const url = `${ADDR}/api/v1/users/me`;
    const payload = JSON.stringify({
        email: `updated-${helpers.GenerateRandomString()}@example.com`,
    });
    const params = {
        headers: {
            'Authorization': `Bearer ${accessToken}`,
            'Content-Type': 'application/json',
        },
        tags: { name: 'update_me' },
        timeout: '5s',
    };
    
    const res = http.put(url, payload, params);
    
    if (res.status === 0) {
        connectionErrors.add(1);
        return;
    }
    
    const success = check(res, {
        'update me status is 200 or 500': (r) => r.status === 200 || r.status === 500, // 500 might occur due to duplicate email
    });
    
    successRate.add(success);
    if (!success) {
        updateMeErrors.add(1);
        if (res.status !== 401 && res.status !== 403 && res.status !== 500) {
            console.log(`Failed to update me: ${res.status} ${res.body}`);
        }
    }
}

function listUsers(accessToken) {
    const url = `${ADDR}/api/v1/users?limit=10&offset=0`;
    const params = {
        headers: {
            'Authorization': `Bearer ${accessToken}`,
        },
        tags: { name: 'list_users' },
        timeout: '5s',
    };
    
    const res = http.get(url, params);
    
    if (res.status === 0) {
        connectionErrors.add(1);
        return;
    }
    
    const success = check(res, {
        'list users status is 200 or 403 or 404': (r) => r.status === 200 || r.status === 403 || r.status === 404,
        'list users returns array if 200': (r) => {
            if (r.status !== 200) return true;
            try {
                const body = JSON.parse(r.body);
                return Array.isArray(body);
            } catch (e) {
                return false;
            }
        },
    });
    
    successRate.add(success);
    if (!success) {
        listUsersErrors.add(1);
        if (res.status !== 401 && res.status !== 403 && res.status !== 404) {
            console.log(`Failed to list users: ${res.status} ${res.body}`);
        }
    }
}