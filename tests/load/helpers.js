import { sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

const httpServerAddressEnvVarName = 'HTTP_SERVER_ADDRESS';
const grpcServerAddressEnvVarName = 'GRPC_SERVER_ADDRESS';
const grpcAPIKeyEnvVarName = 'GRPC_API_KEY';

const defaultHttpServerAddress = "http://localhost:8080";
const defaultGrpcServerAddress = "localhost:50051";
const defaultGrpcAPIKey = "api_kXqdf2uQ7hmOARp-pZrhA6_IsZSeKCmSEM4YFKBGIzA";

export function HTTPServerAddress() {
    return __ENV[httpServerAddressEnvVarName] || defaultHttpServerAddress;
}

export function GRPCServerAddress() {
    return __ENV[grpcServerAddressEnvVarName] || defaultGrpcServerAddress;
}

export function GRPCAPIKey() {
    return __ENV[grpcAPIKeyEnvVarName] || defaultGrpcAPIKey;
}

const randomStringDefaultLength = 8;

export function GenerateRandomString(prefix = "") {
    return `${prefix}${randomString(randomStringDefaultLength)}`;
}

export function randomSleep(maxSeconds = 2) {
    sleep(Math.random() * maxSeconds);
}
