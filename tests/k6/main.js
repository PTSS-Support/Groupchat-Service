import { options } from './options.js';
import healthTest from "./scripts/health-test.js";
import { API_URL, HEADERS } from './config.js';

export { options };

export function setup() {
    return {
        apiUrl: API_URL,
        headers: HEADERS
    };
}

export default function (data) {
    healthTest(data);
}