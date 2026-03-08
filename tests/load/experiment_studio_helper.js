import http from 'k6/http';
import { check, fail, group, sleep } from 'k6';

// k6 helper: truthful Experiment Studio workflow smoke test.
//
// Usage:
//   k6 run tests/load/experiment_studio_helper.js
//   k6 run --env API_BASE_URL=http://localhost:8081 --env FRONTEND_BASE_URL=http://localhost:3000 tests/load/experiment_studio_helper.js

const API_BASE_URL = __ENV.API_BASE_URL || 'http://localhost:8081';
const FRONTEND_BASE_URL = __ENV.FRONTEND_BASE_URL || 'http://localhost:3000';
const ADMIN_EMAIL = __ENV.ADMIN_EMAIL || 'admin@paywall.local';
const ADMIN_PASS = __ENV.ADMIN_PASS || 'admin12345';
const EXPERIMENT_PREFIX = __ENV.EXPERIMENT_PREFIX || 'k6 Experiment Studio';
const STEP_SLEEP_SECONDS = Number(__ENV.STEP_SLEEP_SECONDS || '0');

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    http_req_failed: ['rate==0'],
  },
};

function maybeSleep() {
  if (STEP_SLEEP_SECONDS > 0) sleep(STEP_SLEEP_SECONDS);
}

function parseJson(res, label) {
  try {
    return res.json();
  } catch (error) {
    fail(`${label}: expected JSON response, got status=${res.status} body=${res.body}`);
  }
}

function unwrapData(payload) {
  return payload && typeof payload === 'object' && payload.data !== undefined ? payload.data : payload;
}

function expectStatus(res, expected, label) {
  const ok = check(res, {
    [`${label} returned HTTP ${expected}`]: (response) => response.status === expected,
  });
  if (!ok) fail(`${label}: expected HTTP ${expected}, got ${res.status} body=${res.body}`);
}

function jsonHeaders(extra = {}) {
  return { 'Content-Type': 'application/json', ...extra };
}

function bearerHeaders(token) {
  return jsonHeaders({ Authorization: `Bearer ${token}` });
}

function studioCookieHeaders(token) {
  return { Cookie: `admin_access_token=${token}` };
}

function requestJson(method, url, body, params = {}) {
  return http.request(method, url, body ? JSON.stringify(body) : null, params);
}

function assertExperimentShape(experiment, label) {
  if (!experiment || typeof experiment !== 'object') fail(`${label}: missing experiment payload`);
  if (!experiment.id) fail(`${label}: experiment id missing`);
  if (!Array.isArray(experiment.arms) || experiment.arms.length < 2) {
    fail(`${label}: expected at least two experiment arms`);
  }
}

export function setup() {
  const healthRes = http.get(`${API_BASE_URL}/health`);
  expectStatus(healthRes, 200, 'API health');

  const loginRes = requestJson('POST', `${API_BASE_URL}/v1/admin/auth/login`, {
    email: ADMIN_EMAIL,
    password: ADMIN_PASS,
  }, { headers: jsonHeaders() });
  expectStatus(loginRes, 200, 'Admin login');

  const loginData = unwrapData(parseJson(loginRes, 'Admin login'));
  if (!loginData.access_token) fail('Admin login did not return access_token');

  return { token: loginData.access_token };
}

export default function (data) {
  const token = data.token;
  const stamp = `${Date.now()}-${__VU}-${__ITER}`;
  const createdName = `${EXPERIMENT_PREFIX} ${stamp}`;
  let experiment = null;

  group('create draft experiment', () => {
    const createRes = requestJson('POST', `${API_BASE_URL}/v1/admin/experiments`, {
      name: createdName,
      description: 'k6 helper workflow for Experiment Studio',
      status: 'draft',
      algorithm_type: 'thompson_sampling',
      is_bandit: true,
      min_sample_size: 120,
      confidence_threshold_percent: 95,
      start_at: null,
      end_at: null,
      arms: [
        { name: 'Control', description: 'Baseline', is_control: true, traffic_weight: 1 },
        { name: 'Variant A', description: 'Alternative', is_control: false, traffic_weight: 1 },
      ],
    }, { headers: bearerHeaders(token), tags: { flow: 'studio_create' } });

    expectStatus(createRes, 201, 'Create experiment');
    experiment = unwrapData(parseJson(createRes, 'Create experiment'));
    assertExperimentShape(experiment, 'Create experiment');
    if (experiment.status !== 'draft') fail(`Create experiment: expected draft, got ${experiment.status}`);
  });

  maybeSleep();

  group('frontend studio dashboard includes experiment', () => {
    const dashboardRes = http.get(`${FRONTEND_BASE_URL}/api/admin/studio/dashboard`, {
      headers: studioCookieHeaders(token),
      tags: { flow: 'studio_dashboard' },
    });
    expectStatus(dashboardRes, 200, 'Studio dashboard');

    const dashboard = unwrapData(parseJson(dashboardRes, 'Studio dashboard'));
    if (!Array.isArray(dashboard.experiments)) fail('Studio dashboard: experiments array missing');

    const found = dashboard.experiments.some((item) => item.id === experiment.id);
    if (!found) fail(`Studio dashboard: created experiment ${experiment.id} not found`);
  });

  maybeSleep();

  group('frontend studio snapshot resolves draft', () => {
    const snapshotRes = http.get(
      `${FRONTEND_BASE_URL}/api/admin/studio/snapshot?experimentId=${encodeURIComponent(experiment.id)}`,
      { headers: studioCookieHeaders(token), tags: { flow: 'studio_snapshot_draft' } },
    );
    expectStatus(snapshotRes, 200, 'Studio snapshot draft');

    const snapshot = unwrapData(parseJson(snapshotRes, 'Studio snapshot draft'));
    if (snapshot.experiment.id !== experiment.id) fail('Studio snapshot draft: experiment id mismatch');
  });

  maybeSleep();

  group('update draft metadata', () => {
    const updatedName = `${createdName} Updated`;
    const updateRes = requestJson('PUT', `${API_BASE_URL}/v1/admin/experiments/${experiment.id}`, {
      name: updatedName,
      description: 'Updated by k6 Experiment Studio helper',
      algorithm_type: 'ucb',
      is_bandit: true,
      min_sample_size: 250,
      confidence_threshold_percent: 97,
      start_at: null,
      end_at: null,
    }, { headers: bearerHeaders(token), tags: { flow: 'studio_update' } });

    expectStatus(updateRes, 200, 'Update experiment');
    experiment = unwrapData(parseJson(updateRes, 'Update experiment'));
    assertExperimentShape(experiment, 'Update experiment');

    if (experiment.name !== updatedName) fail(`Update experiment: expected updated name, got ${experiment.name}`);
    if (experiment.algorithm_type !== 'ucb') fail(`Update experiment: expected ucb, got ${experiment.algorithm_type}`);
    if (experiment.min_sample_size !== 250) fail('Update experiment: min_sample_size was not updated');
    if (experiment.confidence_threshold_percent !== 97) {
      fail(`Update experiment: expected confidence_threshold_percent=97, got ${experiment.confidence_threshold_percent}`);
    }
  });

  maybeSleep();

  group('frontend snapshot reflects updated draft metadata', () => {
    const snapshotRes = http.get(
      `${FRONTEND_BASE_URL}/api/admin/studio/snapshot?experimentId=${encodeURIComponent(experiment.id)}`,
      { headers: studioCookieHeaders(token), tags: { flow: 'studio_snapshot_updated' } },
    );
    expectStatus(snapshotRes, 200, 'Studio snapshot updated draft');

    const snapshot = unwrapData(parseJson(snapshotRes, 'Studio snapshot updated draft'));
    if (snapshot.experiment.name !== experiment.name) fail('Studio snapshot updated draft: name mismatch');
    if (snapshot.experiment.status !== 'draft') fail(`Studio snapshot updated draft: expected draft, got ${snapshot.experiment.status}`);
  });

  maybeSleep();

  group('launch draft experiment', () => {
    const launchRes = requestJson('POST', `${API_BASE_URL}/v1/admin/experiments/${experiment.id}/resume`, null, {
      headers: { Authorization: `Bearer ${token}` },
      tags: { flow: 'studio_launch' },
    });
    expectStatus(launchRes, 200, 'Launch experiment');

    experiment = unwrapData(parseJson(launchRes, 'Launch experiment'));
    if (experiment.status !== 'running') fail(`Launch experiment: expected running, got ${experiment.status}`);
    if (!experiment.start_at) fail('Launch experiment: start_at was not set');
  });

  maybeSleep();

  group('running snapshot and runtime probes', () => {
    const snapshotRes = http.get(
      `${FRONTEND_BASE_URL}/api/admin/studio/snapshot?experimentId=${encodeURIComponent(experiment.id)}`,
      { headers: studioCookieHeaders(token), tags: { flow: 'studio_snapshot_running' } },
    );
    expectStatus(snapshotRes, 200, 'Studio snapshot running');

    const snapshot = unwrapData(parseJson(snapshotRes, 'Studio snapshot running'));
    if (snapshot.experiment.status !== 'running') fail(`Studio snapshot running: got ${snapshot.experiment.status}`);

    const runtimeUrls = [
      `${API_BASE_URL}/v1/bandit/statistics?experiment_id=${encodeURIComponent(experiment.id)}&win_probs=true`,
      `${API_BASE_URL}/v1/bandit/experiments/${experiment.id}/metrics`,
      `${API_BASE_URL}/v1/bandit/experiments/${experiment.id}/objectives`,
      `${API_BASE_URL}/v1/bandit/experiments/${experiment.id}/window/info`,
    ];

    runtimeUrls.forEach((url) => {
      const res = http.get(url, { tags: { flow: 'studio_runtime_probe' } });
      expectStatus(res, 200, `Runtime probe ${url}`);
    });
  });

  maybeSleep();

  group('pause experiment', () => {
    const pauseRes = requestJson('POST', `${API_BASE_URL}/v1/admin/experiments/${experiment.id}/pause`, null, {
      headers: { Authorization: `Bearer ${token}` },
      tags: { flow: 'studio_pause' },
    });
    expectStatus(pauseRes, 200, 'Pause experiment');

    experiment = unwrapData(parseJson(pauseRes, 'Pause experiment'));
    if (experiment.status !== 'paused') fail(`Pause experiment: expected paused, got ${experiment.status}`);
  });

  maybeSleep();

  group('resume paused experiment', () => {
    const resumeRes = requestJson('POST', `${API_BASE_URL}/v1/admin/experiments/${experiment.id}/resume`, null, {
      headers: { Authorization: `Bearer ${token}` },
      tags: { flow: 'studio_resume' },
    });
    expectStatus(resumeRes, 200, 'Resume experiment');

    experiment = unwrapData(parseJson(resumeRes, 'Resume experiment'));
    if (experiment.status !== 'running') fail(`Resume experiment: expected running, got ${experiment.status}`);
  });

  maybeSleep();

  group('complete experiment', () => {
    const completeRes = requestJson('POST', `${API_BASE_URL}/v1/admin/experiments/${experiment.id}/complete`, null, {
      headers: { Authorization: `Bearer ${token}` },
      tags: { flow: 'studio_complete' },
    });
    expectStatus(completeRes, 200, 'Complete experiment');

    experiment = unwrapData(parseJson(completeRes, 'Complete experiment'));
    if (experiment.status !== 'completed') fail(`Complete experiment: expected completed, got ${experiment.status}`);
    if (!experiment.end_at) fail('Complete experiment: end_at was not set');
  });

  maybeSleep();

  group('completed snapshot and dashboard refresh', () => {
    const snapshotRes = http.get(
      `${FRONTEND_BASE_URL}/api/admin/studio/snapshot?experimentId=${encodeURIComponent(experiment.id)}`,
      { headers: studioCookieHeaders(token), tags: { flow: 'studio_snapshot_completed' } },
    );
    expectStatus(snapshotRes, 200, 'Studio snapshot completed');

    const snapshot = unwrapData(parseJson(snapshotRes, 'Studio snapshot completed'));
    if (snapshot.experiment.status !== 'completed') {
      fail(`Studio snapshot completed: expected completed, got ${snapshot.experiment.status}`);
    }

    const dashboardRes = http.get(`${FRONTEND_BASE_URL}/api/admin/studio/dashboard`, {
      headers: studioCookieHeaders(token),
      tags: { flow: 'studio_dashboard_final' },
    });
    expectStatus(dashboardRes, 200, 'Studio dashboard final');

    const dashboard = unwrapData(parseJson(dashboardRes, 'Studio dashboard final'));
    const found = Array.isArray(dashboard.experiments) && dashboard.experiments.some((item) => item.id === experiment.id);
    if (!found) fail(`Studio dashboard final: experiment ${experiment.id} missing`);
  });

  console.log(`Experiment Studio helper finished successfully for experiment_id=${experiment.id}`);
}