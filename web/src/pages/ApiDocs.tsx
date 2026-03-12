import { useState } from 'react';

interface Endpoint {
  method: string;
  path: string;
  description: string;
  body?: string;
}

interface Section {
  title: string;
  endpoints: Endpoint[];
}

const sections: Section[] = [
  {
    title: 'Instance',
    endpoints: [
      { method: 'POST', path: '/api/v1/instance/create', description: 'Create a new instance', body: '{ "name": "my-instance" }' },
      { method: 'POST', path: '/api/v1/instance/{name}/connect', description: 'Connect (QR or pairing code)', body: '{ "phone_number": "+5511999999999" }' },
      { method: 'POST', path: '/api/v1/instance/{name}/restart', description: 'Restart connection' },
      { method: 'GET', path: '/api/v1/instance/{name}/status', description: 'Get connection status' },
      { method: 'GET', path: '/api/v1/instance', description: 'List all instances' },
      { method: 'DELETE', path: '/api/v1/instance/{name}/logout', description: 'Logout from WhatsApp' },
      { method: 'DELETE', path: '/api/v1/instance/{name}', description: 'Delete instance' },
    ],
  },
  {
    title: 'Messages',
    endpoints: [
      { method: 'POST', path: '/api/v1/instance/{name}/message/send-text', description: 'Send text message', body: '{ "number": "5511999999999", "text": "Hello!" }' },
      { method: 'POST', path: '/api/v1/instance/{name}/message/send-media', description: 'Send media (image/video/doc/audio)', body: '{ "number": "5511999999999", "media_type": "image", "media": "https://...", "caption": "Check this" }' },
      { method: 'POST', path: '/api/v1/instance/{name}/message/send-location', description: 'Send location', body: '{ "number": "5511999999999", "latitude": -23.55, "longitude": -46.63, "name": "SP" }' },
      { method: 'POST', path: '/api/v1/instance/{name}/message/send-contact', description: 'Send contact card', body: '{ "number": "5511999999999", "contact_name": "John", "phones": [{"number": "+5511888888888"}] }' },
      { method: 'POST', path: '/api/v1/instance/{name}/message/send-reaction', description: 'Send reaction', body: '{ "number": "5511999999999", "message_id": "ABC123", "emoji": "\\ud83d\\udc4d" }' },
      { method: 'POST', path: '/api/v1/instance/{name}/message/send-sticker', description: 'Send sticker', body: '{ "number": "5511999999999", "sticker": "base64..." }' },
    ],
  },
  {
    title: 'Chat',
    endpoints: [
      { method: 'POST', path: '/api/v1/instance/{name}/chat/check-number', description: 'Check if numbers are on WhatsApp', body: '{ "numbers": ["5511999999999"] }' },
      { method: 'POST', path: '/api/v1/instance/{name}/chat/mark-read', description: 'Mark messages as read', body: '{ "chat_jid": "5511999999999@s.whatsapp.net", "message_ids": ["ABC"] }' },
      { method: 'POST', path: '/api/v1/instance/{name}/chat/delete-message', description: 'Delete message for everyone', body: '{ "chat_jid": "...", "message_id": "ABC" }' },
      { method: 'POST', path: '/api/v1/instance/{name}/chat/edit-message', description: 'Edit a sent message', body: '{ "chat_jid": "...", "message_id": "ABC", "text": "edited" }' },
      { method: 'POST', path: '/api/v1/instance/{name}/chat/send-presence', description: 'Send typing/recording', body: '{ "chat_jid": "...", "presence": "composing" }' },
      { method: 'POST', path: '/api/v1/instance/{name}/chat/block', description: 'Block/unblock user', body: '{ "jid": "...", "action": "block" }' },
      { method: 'GET', path: '/api/v1/instance/{name}/chat/contacts', description: 'List contacts' },
      { method: 'GET', path: '/api/v1/instance/{name}/chat/profile/{jid}', description: 'Get user profile' },
    ],
  },
  {
    title: 'Group',
    endpoints: [
      { method: 'POST', path: '/api/v1/instance/{name}/group/create', description: 'Create group', body: '{ "name": "My Group", "participants": ["5511999999999"] }' },
      { method: 'GET', path: '/api/v1/instance/{name}/group', description: 'List joined groups' },
      { method: 'GET', path: '/api/v1/instance/{name}/group/{jid}', description: 'Get group info' },
      { method: 'PUT', path: '/api/v1/instance/{name}/group/{jid}/name', description: 'Update group name' },
      { method: 'PUT', path: '/api/v1/instance/{name}/group/{jid}/description', description: 'Update description' },
      { method: 'PUT', path: '/api/v1/instance/{name}/group/{jid}/photo', description: 'Update group photo' },
      { method: 'PUT', path: '/api/v1/instance/{name}/group/{jid}/settings', description: 'Update settings' },
      { method: 'POST', path: '/api/v1/instance/{name}/group/{jid}/participants', description: 'Manage participants', body: '{ "action": "add", "participants": ["5511999999999"] }' },
      { method: 'GET', path: '/api/v1/instance/{name}/group/{jid}/invite-link', description: 'Get invite link' },
      { method: 'DELETE', path: '/api/v1/instance/{name}/group/{jid}/leave', description: 'Leave group' },
    ],
  },
  {
    title: 'Webhook',
    endpoints: [
      { method: 'POST', path: '/api/v1/instance/{name}/webhook', description: 'Set webhook', body: '{ "url": "https://...", "events": [], "enabled": true }' },
      { method: 'GET', path: '/api/v1/instance/{name}/webhook', description: 'Get webhook config' },
      { method: 'DELETE', path: '/api/v1/instance/{name}/webhook', description: 'Remove webhook' },
    ],
  },
  {
    title: 'Settings',
    endpoints: [
      { method: 'GET', path: '/api/v1/instance/{name}/settings', description: 'Get instance settings' },
      { method: 'PUT', path: '/api/v1/instance/{name}/settings', description: 'Update settings', body: '{ "always_online": true, "reject_call": false }' },
    ],
  },
];

const methodColors: Record<string, string> = {
  GET: '#22c55e',
  POST: '#3b82f6',
  PUT: '#eab308',
  DELETE: '#ef4444',
};

export default function ApiDocs() {
  const [expanded, setExpanded] = useState<string | null>(null);

  return (
    <div className="page">
      <div className="page-header">
        <h2>API Documentation</h2>
      </div>

      <div className="docs-info">
        <p>Base URL: <code>{window.location.origin}</code></p>
        <p>Auth: <code>X-API-Key: &lt;your-key&gt;</code> or <code>Authorization: Bearer &lt;token&gt;</code></p>
      </div>

      <div className="docs-events">
        <h3>Webhook Events</h3>
        <div className="event-list">
          {[
            ['connection.update', 'Connection state changed (open/close/connecting)'],
            ['message.received', 'New message received'],
            ['message.sent', 'Message sent by this device'],
            ['message.updated', 'Message status updated (delivered/read)'],
            ['message.deleted', 'Message deleted'],
            ['message.reaction', 'Reaction received'],
            ['presence.update', 'User presence changed'],
            ['group.update', 'Group metadata changed'],
            ['group.participants', 'Member joined/left/promoted/demoted'],
            ['call.received', 'Incoming call'],
            ['qrcode.updated', 'New QR code generated'],
          ].map(([name, desc]) => (
            <div key={name} className="event-item">
              <code>{name}</code>
              <span>{desc}</span>
            </div>
          ))}
        </div>
      </div>

      {sections.map((section) => (
        <div key={section.title} className="docs-section">
          <h3>{section.title}</h3>
          {section.endpoints.map((ep) => {
            const key = `${ep.method}-${ep.path}`;
            const isOpen = expanded === key;
            return (
              <div
                key={key}
                className={`endpoint ${isOpen ? 'expanded' : ''}`}
                onClick={() => setExpanded(isOpen ? null : key)}
              >
                <div className="endpoint-header">
                  <span
                    className="method"
                    style={{ backgroundColor: methodColors[ep.method] || '#6b7280' }}
                  >
                    {ep.method}
                  </span>
                  <code className="path">{ep.path}</code>
                  <span className="desc">{ep.description}</span>
                </div>
                {isOpen && ep.body && (
                  <div className="endpoint-body">
                    <pre>{ep.body}</pre>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      ))}
    </div>
  );
}
