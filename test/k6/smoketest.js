// smotest.js edited after copy/pasting from https://k6.io docs
// Using this like loadtest because of limited cpu/memory/other

import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  // testbed created with "make dev-env" requires this name resolution
  // this does not set the host header
  hosts: {
    'test.ingress-nginx-controller.ga:80': '127.0.0.1:80',
    'test.ingress-nginx-controller.ga:443': '127.0.0.1:443',
  },
  // below 3 lines documented at https://k6.io
  duration: '1m',
  vus: 50,
  thresholds: {
    http_req_failed: ['rate<0.01'], // http errors should be less than 1%
    http_req_duration: ['p(95)<500'], // 95 percent of response times must be below 500ms
    http_req_duration: ['p(99)<1500'], // 99 percent of response times must be below 1500ms
  },
};

export default function () {
  // docs of k6 say this is how to adds host header
  // needed as ingress is created with this host value
  const params = {
    headers: {'host': 'test.ingress-nginx-controller.ga'},
  };
  // httpbin.org documents these requests
  const req1 = {
  	method: 'GET',
  	url: 'http://test.ingress-nginx-controller.ga/ip',
  };
  const req2 = {
  	method: 'GET',
  	url: 'http://test.ingress-nginx-controller.ga/image/svg',
  };
  const req3 = {
	params: {
	  headers: {
	    'Content-Type': 'application/json' 
      },
	},
  	method: 'POST',
  	url: 'https://test.ingress-nginx-controller.ga/post',
  	body: {
  	  'key1': 'Hello World!',
  	},
  };
  const req4 = {
    method: 'GET',
    url: 'https://test.ingress-nginx-controller.ga/basic-auth/admin/admin',
    params: {
      headers: {
        'accept': 'application/jsom',
      }
    }
  }
  for(let i=0; i<20; i++){
    const res = http.batch([req0, req1, req2, req3, req4], params);
    sleep(1);
  }
}
