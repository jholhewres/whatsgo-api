import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import type { InstanceSettings, Webhook } from '../lib/api';
import {
  getSettings,
  updateSettings,
  getWebhook,
  setWebhook,
  deleteWebhook,
  getStatus,
} from '../lib/api';
import StatusBadge from '../components/StatusBadge';

const EVENT_TYPES = [
  'connection.update',
  'message.received',
  'message.sent',
  'message.updated',
  'message.deleted',
  'message.reaction',
  'presence.update',
  'group.update',
  'group.participants',
  'call.received',
  'qrcode.updated',
];

export default function InstanceDetail() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();

  const [status, setStatus] = useState<{ status: string; phone: string } | null>(null);
  const [settings, setSettings] = useState<InstanceSettings | null>(null);
  const [webhook, setWebhookState] = useState<Webhook | null>(null);
  const [webhookUrl, setWebhookUrl] = useState('');
  const [webhookEvents, setWebhookEvents] = useState<string[]>([]);
  const [webhookEnabled, setWebhookEnabled] = useState(true);
  const [saving, setSaving] = useState(false);
  const [msg, setMsg] = useState('');

  useEffect(() => {
    if (!name) return;
    loadAll();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [name]);

  const loadAll = async () => {
    if (!name) return;
    try {
      const [s, st, wh] = await Promise.allSettled([
        getStatus(name),
        getSettings(name),
        getWebhook(name),
      ]);
      if (s.status === 'fulfilled') {
        const inst = s.value.instance;
        setStatus({ status: inst.status, phone: inst.phone });
      }
      if (st.status === 'fulfilled') setSettings(st.value);
      if (wh.status === 'fulfilled' && wh.value) {
        setWebhookState(wh.value);
        setWebhookUrl(wh.value.url || '');
        setWebhookEvents(wh.value.events || []);
        setWebhookEnabled(wh.value.enabled ?? true);
      }
    } catch {
      // partial failures are ok
    }
  };

  const handleSaveSettings = async () => {
    if (!name || !settings) return;
    setSaving(true);
    try {
      await updateSettings(name, settings);
      setMsg('Settings saved');
      setTimeout(() => setMsg(''), 3000);
    } catch (err) {
      setMsg(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  };

  const handleSaveWebhook = async () => {
    if (!name) return;
    setSaving(true);
    try {
      await setWebhook(name, { url: webhookUrl, events: webhookEvents, enabled: webhookEnabled });
      setMsg('Webhook saved');
      setTimeout(() => setMsg(''), 3000);
    } catch (err) {
      setMsg(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteWebhook = async () => {
    if (!name || !confirm('Remove webhook?')) return;
    try {
      await deleteWebhook(name);
      setWebhookState(null);
      setWebhookUrl('');
      setWebhookEvents([]);
      setMsg('Webhook removed');
      setTimeout(() => setMsg(''), 3000);
    } catch (err) {
      setMsg(err instanceof Error ? err.message : 'Failed');
    }
  };

  const toggleEvent = (evt: string) => {
    setWebhookEvents((prev) =>
      prev.includes(evt) ? prev.filter((e) => e !== evt) : [...prev, evt]
    );
  };

  if (!name) return null;

  return (
    <div className="page">
      <div className="page-header">
        <button className="btn btn-secondary" onClick={() => navigate('/')}>
          &larr; Back
        </button>
        <h2>{name}</h2>
        {status && <StatusBadge status={status.status} />}
      </div>

      {msg && <div className="alert alert-info">{msg}</div>}

      {status && (
        <div className="section">
          <h3>Status</h3>
          <div className="info-grid">
            <div className="info-row">
              <span className="label">Status:</span>
              <span>{status.status}</span>
            </div>
            <div className="info-row">
              <span className="label">Phone:</span>
              <span>{status.phone || 'N/A'}</span>
            </div>
          </div>
        </div>
      )}

      {settings && (
        <div className="section">
          <h3>Settings</h3>
          <div className="settings-grid">
            <label className="toggle-row">
              <input
                type="checkbox"
                checked={settings.always_online}
                onChange={(e) => setSettings({ ...settings, always_online: e.target.checked })}
              />
              <span>Always Online</span>
            </label>
            <label className="toggle-row">
              <input
                type="checkbox"
                checked={settings.read_messages}
                onChange={(e) => setSettings({ ...settings, read_messages: e.target.checked })}
              />
              <span>Auto Read Messages</span>
            </label>
            <label className="toggle-row">
              <input
                type="checkbox"
                checked={settings.read_receipts}
                onChange={(e) => setSettings({ ...settings, read_receipts: e.target.checked })}
              />
              <span>Send Read Receipts</span>
            </label>
            <label className="toggle-row">
              <input
                type="checkbox"
                checked={settings.reject_call}
                onChange={(e) => setSettings({ ...settings, reject_call: e.target.checked })}
              />
              <span>Reject Calls</span>
            </label>
            {settings.reject_call && (
              <div className="input-group">
                <label>Reject Call Message</label>
                <input
                  type="text"
                  value={settings.reject_call_message}
                  onChange={(e) => setSettings({ ...settings, reject_call_message: e.target.value })}
                  placeholder="I'm busy, call me later"
                />
              </div>
            )}
            <label className="toggle-row">
              <input
                type="checkbox"
                checked={settings.groups_ignore}
                onChange={(e) => setSettings({ ...settings, groups_ignore: e.target.checked })}
              />
              <span>Ignore Group Messages</span>
            </label>
            <label className="toggle-row">
              <input
                type="checkbox"
                checked={settings.webhook_base64}
                onChange={(e) => setSettings({ ...settings, webhook_base64: e.target.checked })}
              />
              <span>Webhook Base64 Media</span>
            </label>
          </div>
          <button className="btn btn-primary" onClick={handleSaveSettings} disabled={saving}>
            {saving ? 'Saving...' : 'Save Settings'}
          </button>
        </div>
      )}

      <div className="section">
        <h3>Webhook</h3>
        <div className="input-group">
          <label>URL</label>
          <input
            type="url"
            value={webhookUrl}
            onChange={(e) => setWebhookUrl(e.target.value)}
            placeholder="https://example.com/webhook"
          />
        </div>
        <label className="toggle-row">
          <input
            type="checkbox"
            checked={webhookEnabled}
            onChange={(e) => setWebhookEnabled(e.target.checked)}
          />
          <span>Enabled</span>
        </label>
        <div className="events-grid">
          <label className="label">Events (empty = all):</label>
          {EVENT_TYPES.map((evt) => (
            <label key={evt} className="toggle-row compact">
              <input
                type="checkbox"
                checked={webhookEvents.includes(evt)}
                onChange={() => toggleEvent(evt)}
              />
              <span>{evt}</span>
            </label>
          ))}
        </div>
        <div className="btn-group">
          <button className="btn btn-primary" onClick={handleSaveWebhook} disabled={saving}>
            {saving ? 'Saving...' : 'Save Webhook'}
          </button>
          {webhook && (
            <button className="btn btn-danger" onClick={handleDeleteWebhook}>
              Remove Webhook
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
