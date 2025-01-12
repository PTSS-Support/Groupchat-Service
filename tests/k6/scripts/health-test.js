import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { errorRate } from '../metrics.js';

export default function (data) {
    group('Health Check Endpoints', () => {
        // Test /q/health/ready endpoint
        group('Readiness Check', () => {
            try {
                const res = http.get(`${data.apiUrl}/q/health/ready`, {
                    headers: data.headers
                });

                const body = JSON.parse(res.body);
                const checks = body.checks || [];

                const checkRes = check(res, {
                    'status is 200': (r) => r.status === 200,
                    'response is JSON': (r) => r.headers['Content-Type'].includes('application/json'),
                    'status is UP': (r) => body.status === 'UP',
                    'checks array exists': (r) => Array.isArray(checks),
                    'keycloak check exists': (r) => checks.some(check =>
                        check.name === 'Keycloak health check' && check.status === 'UP'
                    )
                });

                if (!checkRes) {
                    errorRate.add(1);
                }
            } catch (error) {
                console.error(`Readiness check failed: ${error.message}`);
                errorRate.add(1);
            }
        });

        // Test /q/health/live endpoint
        group('Liveness Check', () => {
            try {
                const res = http.get(`${data.apiUrl}/q/health/live`, {
                    headers: data.headers
                });

                const body = JSON.parse(res.body);

                const checkRes = check(res, {
                    'status is 200': (r) => r.status === 200,
                    'response is JSON': (r) => r.headers['Content-Type'].includes('application/json'),
                    'status is UP': (r) => body.status === 'UP'
                });

                if (!checkRes) {
                    errorRate.add(1);
                }
            } catch (error) {
                console.error(`Liveness check failed: ${error.message}`);
                errorRate.add(1);
            }
        });

        // Test general health endpoint
        group('General Health Check', () => {
            try {
                const res = http.get(`${data.apiUrl}/q/health`, {
                    headers: data.headers
                });

                const body = JSON.parse(res.body);

                const checkRes = check(res, {
                    'status is 200': (r) => r.status === 200,
                    'response is JSON': (r) => r.headers['Content-Type'].includes('application/json'),
                    'status is UP': (r) => body.status === 'UP'
                });

                if (!checkRes) {
                    errorRate.add(1);
                }
            } catch (error) {
                console.error(`Health check failed: ${error.message}`);
                errorRate.add(1);
            }
        });
    });

    sleep(1);
}