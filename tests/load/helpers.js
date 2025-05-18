import { sleep } from 'k6';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

const httpServerAddressEnvVarName = 'HTTP_SERVER_ADDRESS';
const grpcServerAddressEnvVarName = 'GRPC_SERVER_ADDRESS';

const defaultHttpServerAddress = "http://localhost:8080/api/v1";
const defaultGrpcServerAddress = "localhost:50051";

export function HTTPServerAddress() {
    return __ENV[httpServerAddressEnvVarName] || defaultHttpServerAddress;
}

export function GRPCServerAddress() {
    return __ENV[grpcServerAddressEnvVarName] || defaultGrpcServerAddress;
}

const randomStringDefaultLength = 8;

export function GenerateRandomString(prefix = "") {
    return `${prefix}${randomString(randomStringDefaultLength)}`;
}

export function randomSleep(maxSeconds = 2) {
    sleep(Math.random() * maxSeconds);
}
