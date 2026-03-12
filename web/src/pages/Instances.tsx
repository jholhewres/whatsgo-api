import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import type { Instance } from '../lib/api';
import {
  listInstances,
  createInstance,
  connectInstance,
  deleteInstance,
  logoutInstance,
} from '../lib/api';
import { useApi } from '../hooks/useApi';
import StatusBadge from '../components/StatusBadge';
import QRCodeModal from '../components/QRCodeModal';

export default function Instances() {
  const { data: instances, loading, error, execute } = useApi<Instance[]>();
  const [newName, setNewName] = useState('');
  const [creating, setCreating] = useState(false);
  const [qrModal, setQrModal] = useState<{
    name: string;
    qrCode: string | null;
    pairingCode: string | null;
  } | null>(null);

  const load = () => execute(() => listInstances());

  useEffect(() => {
    load();
    const interval = setInterval(load, 10000);
    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newName.trim()) return;
    setCreating(true);
    try {
      await createInstance(newName.trim());
      setNewName('');
      await load();
    } catch {
      // error handled by useApi
    } finally {
      setCreating(false);
    }
  };

  const handleConnect = async (name: string) => {
    setQrModal({ name, qrCode: null, pairingCode: null });
    try {
      const resp = await connectInstance(name);
      if (resp.qr_code) {
        setQrModal({ name, qrCode: resp.qr_code, pairingCode: null });
      } else if (resp.pairing_code) {
        setQrModal({ name, qrCode: null, pairingCode: resp.pairing_code });
      } else {
        setQrModal(null);
        load();
      }
    } catch {
      setQrModal(null);
    }
  };

  const handleLogout = async (name: string) => {
    if (!confirm(`Logout instance "${name}"?`)) return;
    try {
      await logoutInstance(name);
      load();
    } catch {
      // handled
    }
  };

  const handleDelete = async (name: string) => {
    if (!confirm(`Delete instance "${name}"? This cannot be undone.`)) return;
    try {
      await deleteInstance(name);
      load();
    } catch {
      // handled
    }
  };

  return (
    <div className="page">
      <div className="page-header">
        <h2>Instances</h2>
        <form className="create-form" onSubmit={handleCreate}>
          <input
            type="text"
            placeholder="Instance name"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
          />
          <button type="submit" className="btn btn-primary" disabled={creating}>
            {creating ? 'Creating...' : 'Create'}
          </button>
        </form>
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      {loading && !instances && <div className="loading"><div className="spinner" /></div>}

      <div className="card-grid">
        {instances?.map((inst) => (
          <div key={inst.id} className="card">
            <div className="card-header">
              <h3>{inst.name}</h3>
              <StatusBadge status={inst.status} />
            </div>
            <div className="card-body">
              <div className="info-row">
                <span className="label">Phone:</span>
                <span>{inst.phone || 'Not connected'}</span>
              </div>
              {inst.token && (
                <div className="info-row">
                  <span className="label">Token:</span>
                  <span className="token">{inst.token.slice(0, 8)}...</span>
                </div>
              )}
              <div className="info-row">
                <span className="label">Created:</span>
                <span>{new Date(inst.created_at).toLocaleDateString()}</span>
              </div>
            </div>
            <div className="card-actions">
              {inst.status !== 'open' && (
                <button className="btn btn-primary btn-sm" onClick={() => handleConnect(inst.name)}>
                  Connect
                </button>
              )}
              <Link to={`/instance/${inst.name}`} className="btn btn-secondary btn-sm">
                Settings
              </Link>
              {inst.status === 'open' && (
                <button className="btn btn-warning btn-sm" onClick={() => handleLogout(inst.name)}>
                  Logout
                </button>
              )}
              <button className="btn btn-danger btn-sm" onClick={() => handleDelete(inst.name)}>
                Delete
              </button>
            </div>
          </div>
        ))}

        {instances?.length === 0 && (
          <div className="empty-state">
            <p>No instances yet. Create one to get started.</p>
          </div>
        )}
      </div>

      {qrModal && (
        <QRCodeModal
          qrCode={qrModal.qrCode}
          pairingCode={qrModal.pairingCode}
          onClose={() => {
            setQrModal(null);
            load();
          }}
        />
      )}
    </div>
  );
}
