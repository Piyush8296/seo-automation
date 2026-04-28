const BASE = '/api'

async function handleResponse(res) {
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error || `HTTP ${res.status}`)
  }
  return res.status === 204 ? null : res.json()
}

export const api = {
  startAudit: (req) =>
    fetch(`${BASE}/audits`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    }).then(handleResponse),

  listAudits: () =>
    fetch(`${BASE}/audits`).then(handleResponse),

  getAudit: (id) =>
    fetch(`${BASE}/audits/${id}`).then(handleResponse),

  deleteAudit: (id) =>
    fetch(`${BASE}/audits/${id}`, { method: 'DELETE' }).then(handleResponse),

  cancelAudit: (id) =>
    fetch(`${BASE}/audits/${id}/cancel`, { method: 'POST' }).then(handleResponse),

  diffAudits: (a, b) =>
    fetch(`${BASE}/audits/diff?a=${a}&b=${b}`).then(handleResponse),

  getCheckCatalog: () =>
    fetch(`${BASE}/checks`).then(handleResponse),

  getExternalCheckCatalog: () =>
    fetch(`${BASE}/external-checks/catalog`).then(handleResponse),

  getLocalSEO: () =>
    fetch(`${BASE}/local-seo/gbp`).then(handleResponse),

  submitGBPAction: (req) =>
    fetch(`${BASE}/local-seo/gbp/actions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    }).then(handleResponse),

  getSearchIntegrations: () =>
    fetch(`${BASE}/search-integrations`).then(handleResponse),

  connectSearchOAuth: (req) =>
    fetch(`${BASE}/search-integrations/oauth/connect`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    }).then(handleResponse),

  submitSearchIntegrationAction: (req) =>
    fetch(`${BASE}/search-integrations/actions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    }).then(handleResponse),

  getCrawlerEvidence: () =>
    fetch(`${BASE}/crawler-evidence`).then(handleResponse),

  runCrawlerEvidence: (req) =>
    fetch(`${BASE}/crawler-evidence/run`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    }).then(handleResponse),

  getSettings: () =>
    fetch(`${BASE}/settings`).then(handleResponse),

  updateSettings: (cfg) =>
    fetch(`${BASE}/settings`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(cfg),
    }).then(handleResponse),

  eventsURL: (id) => `${BASE}/audits/${id}/events`,
  reportURL: (id) => `${BASE}/audits/${id}/report.html`,
  reportDownloadURL: (id) => `${BASE}/audits/${id}/report.html`,
  reportJSONURL: (id) => `${BASE}/audits/${id}/report.json`,

  getReportJSON: (id) =>
    fetch(`${BASE}/audits/${id}/report.json`).then(handleResponse),
}
