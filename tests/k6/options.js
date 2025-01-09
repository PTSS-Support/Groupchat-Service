export const options = {
    stages: [
        { duration: '30s', target: 2 },  // Ramp up to 2 users
        { duration: '1m', target: 4 },   // Stay at 4 users for 1 minute
        { duration: '30s', target: 0 },  // Ramp down
    ],
    thresholds: {
        http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
        'errors': ['rate<0.01'],          // Error rate should be below 1%
    }
};