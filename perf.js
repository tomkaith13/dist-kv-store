import { Trend } from 'k6/metrics';
import http from 'k6/http';
import { check,sleep } from 'k6';

export const options = {
  vus: 1, 
  duration: '10s', 
};

const trend1 = new Trend('dkv_set_key', true);
const trend2 = new Trend('dkv_get_key', true);

let counter = 1;

export default function () {

  const postUrl = 'http://localhost:8888/key';
  let payload = {
    key: counter.toString(),
    value: counter.toString(),
  };

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const res = http.post(postUrl, JSON.stringify(payload), params);

  check(res, {
    'post status was 201': (r) => r.status === 201,
  });
  
  if (res.status == 201) {
    trend1.add(res.timings.duration);
  }
  
  // sleep to wait for replication
  sleep(1);
  const getUrl = 'http://localhost:8889/key/' + counter.toString();
  const resGet = http.get(getUrl,  params);
  
  check(resGet, {
      'get status was 200': (r) => r.status === 200,
    });
    
    if (resGet.status == 200) {
        trend2.add(resGet.timings.duration);
    }
    
    counter++;

}
