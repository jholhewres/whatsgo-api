const API_BASE = import.meta.env.VITE_API_URL || '';

interface RequestOptions {
  method?: string;
  body?: unknown;
  token?: string;
}

export async function api<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  const token = opts.token || localStorage.getItem('whatsgo_api_key') || '';
  if (token) {
    headers['X-API-Key'] = token;
  }

  const res = await fetch(`${API_BASE}${path}`, {
    method: opts.method || 'GET',
    headers,
    body: opts.body ? JSON.stringify(opts.body) : undefined,
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || `HTTP ${res.status}`);
  }

  return res.json();
}

// Types
export interface Instance {
  id: string;
  name: string;
  token: string;
  status: string;
  phone: string;
  business_name: string;
  whatsmeow_jid: string;
  created_at: string;
  updated_at: string;
}

export interface InstanceSettings {
  instance_id: string;
  reject_call: boolean;
  reject_call_message: string;
  groups_ignore: boolean;
  always_online: boolean;
  read_messages: boolean;
  read_receipts: boolean;
  webhook_base64: boolean;
}

export interface Webhook {
  id: string;
  instance_id: string;
  url: string;
  events: string[];
  headers: Record<string, string>;
  enabled: boolean;
}

export interface ConnectResponse {
  status: string;
  qr_code?: string;
  pairing_code?: string;
}

// API functions
export const listInstances = () => api<Instance[]>('/api/v1/instance');
export const createInstance = (name: string) =>
  api<{ instance: Instance }>('/api/v1/instance/create', { method: 'POST', body: { name } });
export const connectInstance = (name: string, phone?: string) =>
  api<ConnectResponse>(`/api/v1/instance/${name}/connect`, {
    method: 'POST',
    body: phone ? { phone_number: phone } : {},
  });
export const getStatus = (name: string) =>
  api<{ instance: Instance }>(`/api/v1/instance/${name}/status`);
export const restartInstance = (name: string) =>
  api<{ status: string }>(`/api/v1/instance/${name}/restart`, { method: 'POST' });
export const logoutInstance = (name: string) =>
  api<{ status: string }>(`/api/v1/instance/${name}/logout`, { method: 'DELETE' });
export const deleteInstance = (name: string) =>
  api<{ status: string }>(`/api/v1/instance/${name}`, { method: 'DELETE' });

export const getWebhook = (name: string) => api<Webhook>(`/api/v1/instance/${name}/webhook`);
export const setWebhook = (name: string, data: { url: string; events: string[]; enabled: boolean }) =>
  api<{ status: string }>(`/api/v1/instance/${name}/webhook`, { method: 'POST', body: data });
export const deleteWebhook = (name: string) =>
  api<{ status: string }>(`/api/v1/instance/${name}/webhook`, { method: 'DELETE' });

export const getSettings = (name: string) => api<InstanceSettings>(`/api/v1/instance/${name}/settings`);
export const updateSettings = (name: string, data: Partial<InstanceSettings>) =>
  api<{ status: string }>(`/api/v1/instance/${name}/settings`, { method: 'PUT', body: data });
