import http from 'k6/http';

import { sleep } from 'k6';

export const options = {

    stages: [
        { duration: '60s', target: 200 },
        // { duration: '20s', target:  0 },
    ],

};
export default function () {
    const url = 'http://localhost:8080/api/order/v1/order';
    const payload = JSON.stringify({
        client_id: 1
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    http.post(url, payload, params);

    // sleep(0.01);
}